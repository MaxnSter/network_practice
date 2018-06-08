package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_other"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/util"
	"github.com/MaxnSter/gnet/worker"
	"github.com/MaxnSter/network_practice/sudoku"
)

func main() {
	conn := flag.Int("c", 1, "client number")
	rps := flag.Int("r", 100, "request per second")
	addr := flag.String("addr", ":2007", "server address")
	file := flag.String("f", "", "sudoku input file")
	pipelineMode := flag.Bool("p", false, "pipeline mode")

	flag.Parse()
	if *file == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *pipelineMode {
		NewLoadTestWithHook(*conn, *rps, *addr, *file,
			func(client *loadTestClient) {
				client.send(*rps)
			},
			func(client *loadTestClient, ev iface.Event) {
				client.send(1)
			},
		).RunLoadTest(true)
	} else {
		NewLoadTest(*conn, *rps, *addr, *file).RunLoadTest(false)
	}
}

const (
	kHZ = 100
)

type loadTest struct {
	conn int
	rps  int
	addr string
	file string

	input []string
	ticks int64
	soFar int64

	gnetOption *gnet.GnetOption
	pool       iface.WorkerPool
	timers     *timer.TimerManager
	clients    []*loadTestClient
}

func (l *loadTest) RunLoadTest(recordOnly bool) {
	wg := sync.WaitGroup{}

	for _, c := range l.clients {
		wg.Add(1)
		c := c
		go func() {
			c.run()
			wg.Done()
		}()
	}

	l.timers.Start()
	if !recordOnly {
		l.timers.AddTimer(time.Now(), time.Second/kHZ, nil, l.tick)
	}
	l.timers.AddTimer(time.Now(), time.Second, nil, l.tock)

	l.pool.Start()
	wg.Wait()
}

func (l *loadTest) tick(_ time.Time, _ iface.Context) {
	l.ticks++
	reqs := int64(l.rps)*l.ticks/kHZ - l.soFar
	l.soFar += reqs

	if reqs > 0 {
		for _, c := range l.clients {
			c.send(int(reqs))
		}
	}
}

func (l *loadTest) tock(_ time.Time, _ iface.Context) {
	latencies := make([]int, 0)
	infly := 0

	for _, c := range l.clients {
		c.report(&latencies, &infly)
	}

	// 打印测试情况
	logger.Infoln(sudoku.NewPercentile(latencies, infly).Report())
}

func NewLoadTestWithHook(conn int, rps int, addr string, file string,
	onConnectHook func(*loadTestClient), onMessageHook func(*loadTestClient, iface.Event)) *loadTest {

	l := &loadTest{
		conn: conn,
		rps:  rps,
		addr: addr,
		file: file,

		input:      sudoku.ReadInput(file),
		gnetOption: &gnet.GnetOption{Coder: "byte", Packer: "line"},
	}

	l.pool = worker.MustGetWorkerPool("poolRaceOther")
	l.timers = timer.NewTimerManager(l.pool)
	l.clients = make([]*loadTestClient, 0)

	for i := 0; i < l.conn; i++ {
		c := NewLoadTestClientWithCallBack(addr, l.gnetOption, l.pool, "conn"+strconv.Itoa(i), l.input, onConnectHook)
		c.setOnMessageHook(onMessageHook)
		l.clients = append(l.clients, c)
	}
	return l
}

func NewLoadTest(conn int, rps int, addr string, file string) *loadTest {
	l := &loadTest{
		conn: conn,
		rps:  rps,
		addr: addr,
		file: file,

		input:      sudoku.ReadInput(file),
		gnetOption: &gnet.GnetOption{Coder: "byte", Packer: "line"},
	}

	l.pool = worker.MustGetWorkerPool("poolRaceOther")
	l.timers = timer.NewTimerManager(l.pool)
	l.clients = make([]*loadTestClient, 0)

	for i := 0; i < l.conn; i++ {
		c := NewLoadTestClient(addr, l.gnetOption, l.pool, "conn"+strconv.Itoa(i), l.input)
		l.clients = append(l.clients, c)
	}
	return l
}

type loadCounter struct {
	count     int
	sendTime  map[int]time.Time
	latencies []int
}

type loadTestClient struct {
	name  string
	input []string

	loadCounter
	*net.TcpClient
	*net.TcpSession

	hook func(*loadTestClient, iface.Event)
}

func NewLoadTestClientWithCallBack(addr string, gnetOption *gnet.GnetOption, sharePool iface.WorkerPool,
	name string, input []string, hook func(*loadTestClient)) *loadTestClient {

	c := &loadTestClient{
		name:  name,
		input: input,
	}

	c.loadCounter = loadCounter{
		sendTime:  make(map[int]time.Time),
		latencies: make([]int, 0),
	}

	onConnect := func(session *net.TcpSession) {
		c.TcpSession = session
		if hook != nil {
			hook(c)
		}
	}
	cb := gnet.NewCallBackOption(gnet.WithOnConnectCB(onConnect))

	c.TcpClient = gnet.NewClientSharePool(addr, cb, gnetOption, sharePool, c.onMessage)
	return c
}

func NewLoadTestClient(addr string, gnetOption *gnet.GnetOption, sharePool iface.WorkerPool,
	name string, input []string) *loadTestClient {
	c := &loadTestClient{
		name:  name,
		input: input,
	}

	c.loadCounter = loadCounter{
		sendTime:  make(map[int]time.Time),
		latencies: make([]int, 0),
	}

	callbacks := gnet.NewCallBackOption(gnet.WithOnConnectCB(func(session *net.TcpSession) {
		c.TcpSession = session
	},
	))

	c.TcpClient = gnet.NewClientSharePool(addr, callbacks, gnetOption, sharePool, c.onMessage)
	return c
}

func (c *loadTestClient) setOnMessageHook(hook func(client *loadTestClient, ev iface.Event)) {
	c.hook = hook
}

func (c *loadTestClient) run() {
	c.TcpClient.StartAndRun()
}

// 上传自己的延迟记录和未相应个数
func (c *loadTestClient) report(latency *[]int, infly *int) {
	*latency = append(*latency, c.latencies...)
	*infly += len(c.sendTime)
	c.latencies = c.latencies[:0]
}

func (c *loadTestClient) send(n int) {

	if c.TcpSession == nil {
		logger.Infoln("waiting for connected...")
		return
	}

	now := time.Now()
	for i := 0; i < n; i++ {
		content := c.input[c.count%len(c.input)]
		req := fmt.Sprintf("%s-%08d:%s", c.name, c.count, content)
		c.TcpSession.Send(req)

		c.sendTime[c.count] = now
		c.count++
	}
}

func (c *loadTestClient) onMessage(ev iface.Event) {
	switch msg := ev.Message().(type) {
	case []byte:
		if !c.verify(msg, time.Now()) {
			logger.Errorln("error happened, shutdown client")
			c.TcpClient.Stop()
		}

		if c.hook != nil {
			c.hook(c, ev)
		}
	default:
		logger.Errorln("not known msg")
		c.TcpClient.Stop()
	}
}

func (c *loadTestClient) verify(response []byte, recvTime time.Time) bool {
	res := util.BytesToString(response)
	colon := strings.Index(res, ":")
	dash := strings.Index(res, "-")

	if colon == -1 || dash == -1 {
		logger.Errorln("base response:", res)
		return false
	}

	id, err := strconv.Atoi(res[dash+1 : colon])
	if err != nil {
		logger.Errorln(err)
		return false
	}

	sendTime, ok := c.sendTime[id]
	if !ok {
		logger.Errorln("id not record:", id)
		return false
	}

	// 记录延迟
	latencyUS := recvTime.Sub(sendTime).Nanoseconds() / 1000
	c.latencies = append(c.latencies, int(latencyUS))

	// 清除记录
	delete(c.sendTime, id)

	return true
}
