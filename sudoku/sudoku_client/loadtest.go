package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/util"
	"github.com/MaxnSter/gnet/worker_pool"
	"github.com/MaxnSter/network_practice/sudoku"
)

func main() {

	logger.Infoln("pid = ", syscall.Getpid())

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

	// 为了将pipeline测试和压力测试写在一个文件里
	// 所以看起来有点乱
	if *pipelineMode {
		//pipeline测试
		NewLoadTestWithHook(*conn, *rps, *addr, *file,
			func(client *loadTestClient) {
				//建立连接后,发送管道长度的消息
				client.send(*rps)
			},
			func(client *loadTestClient, ev gnet.Event) {
				//每接收到一个消息,发送一个,保持infly始终等于我们指定的pipeline长度
				client.send(1)
			},
		).RunLoadTest(true)
	} else {
		//压力测试
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

	module  gnet.Module
	pool    worker_pool.Pool
	timers  timer.TimerManager
	clients []*loadTestClient
}

func (l *loadTest) RunLoadTest(recordOnly bool) {
	wg := sync.WaitGroup{}

	for _, c := range l.clients {
		wg.Add(1)
		c := c
		go func() {
			c.run()

			//client shutdown
			wg.Done()
		}()
	}

	l.timers.Start()
	//pipeline test recordOnly
	if !recordOnly {
		l.timers.AddTimer(time.Now(), time.Second/kHZ, nil, l.tick)
	}
	l.timers.AddTimer(time.Now(), time.Second, nil, l.tock)

	//pool保证多次重复start()调用安全,只会生成一个eventLoop
	l.pool.Start()
	wg.Wait()
}

func (l *loadTest) tick(_ time.Time, _ iface.Context) {
	// 这三个表示保证不会有因向下取整而少发的问题
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
	// 收集当前所有client的数据
	latencies := make([]int, 0)
	infly := 0

	for _, c := range l.clients {
		c.report(&latencies, &infly)
	}

	logger.Infoln(sudoku.NewPercentile(latencies, infly).Report())
}

// for pipeline test
func NewLoadTestWithHook(conn int, rps int, addr string, file string,
	onConnectHook func(*loadTestClient), onMessageHook func(*loadTestClient, gnet.Event)) *loadTest {

	l := &loadTest{
		conn: conn,
		rps:  rps,
		addr: addr,
		file: file,

		input:  sudoku.ReadInput(file),
		module: gnet.NewModule(gnet.WithPacker("line"), gnet.WithCoder("byte")),
	}

	l.pool = worker_pool.MustGetWorkerPool("poolRaceOther")
	l.timers = timer.NewTimerManager(l.pool)
	l.clients = make([]*loadTestClient, 0)

	for i := 0; i < l.conn; i++ {
		c := NewLoadTestClientWithCallBack(addr, l.module, l.pool, "conn"+strconv.Itoa(i), l.input, onConnectHook)
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

		input:  sudoku.ReadInput(file),
		module: gnet.NewModule(gnet.WithPacker("line"), gnet.WithCoder("byte")),
	}

	l.pool = worker_pool.MustGetWorkerPool("poolRaceOther")
	//all clients and timer share one pool. completely thread safe
	l.timers = timer.NewTimerManager(l.pool)
	l.clients = make([]*loadTestClient, 0)

	for i := 0; i < l.conn; i++ {
		c := NewLoadTestClient(addr, l.module, l.pool, "conn"+strconv.Itoa(i), l.input)
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
	addr  string

	loadCounter
	gnet.NetClient
	gnet.NetSession

	hook func(*loadTestClient, gnet.Event)
}

func NewLoadTestClientWithCallBack(addr string, module gnet.Module, sharePool worker_pool.Pool,
	name string, input []string, hook func(*loadTestClient)) *loadTestClient {

	c := &loadTestClient{
		name:  name,
		input: input,
		addr:  addr,
	}

	c.loadCounter = loadCounter{
		sendTime:  make(map[int]time.Time),
		latencies: make([]int, 0),
	}

	//all clients share one pool. completely thread safe
	module.SetSharePool(sharePool)
	op := gnet.NewOperator(c.onMessage)
	op.SetOnConnected(func(s gnet.NetSession) {
		c.NetSession = s
		if hook != nil {
			hook(c)
		}
	})

	c.NetClient = gnet.NewNetClient("tcp", "sudoku_pipeline_client", module, op)
	return c
}

func NewLoadTestClient(addr string, gnetModule gnet.Module, sharePool worker_pool.Pool,
	name string, input []string) *loadTestClient {
	c := &loadTestClient{
		name:  name,
		input: input,
		addr:  addr,
	}

	c.loadCounter = loadCounter{
		sendTime:  make(map[int]time.Time),
		latencies: make([]int, 0),
	}

	//all clients share one pool. completely thread safe
	gnetModule.SetSharePool(sharePool)
	operate := gnet.NewOperator(c.onMessage)
	operate.SetOnConnected(func(s gnet.NetSession) {
		c.NetSession = s
	})

	c.NetClient = gnet.NewNetClient("tcp", "sudoku_loadtest_client", gnetModule, operate)
	return c
}

func (c *loadTestClient) setOnMessageHook(hook func(client *loadTestClient, ev gnet.Event)) {
	c.hook = hook
}

func (c *loadTestClient) run() {
	c.Connect(c.addr)
}

// 上传自己的延迟记录和未相应个数
func (c *loadTestClient) report(latency *[]int, infly *int) {
	*latency = append(*latency, c.latencies...)
	*infly += len(c.sendTime)
	c.latencies = c.latencies[:0]
}

func (c *loadTestClient) send(n int) {

	if c.NetSession == nil {
		logger.Infoln("waiting for connected...")
		return
	}

	now := time.Now()
	for i := 0; i < n; i++ {
		content := c.input[c.count%len(c.input)]
		req := fmt.Sprintf("%s-%08d:%s", c.name, c.count, content)
		c.NetSession.Send(req)

		c.sendTime[c.count] = now
		c.count++
	}
}

func (c *loadTestClient) onMessage(ev gnet.Event) {
	switch msg := ev.Message().(type) {
	case []byte:
		if !c.verify(msg, time.Now()) {
			logger.Errorln("error happened, shutdown client")
			c.NetClient.Stop()
		}

		if c.hook != nil {
			c.hook(c, ev)
		}
	default:
		logger.Errorln("not known msg")
		c.NetClient.Stop()
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
