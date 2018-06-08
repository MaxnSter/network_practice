package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/MaxnSter/gnet"
	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	_ "github.com/MaxnSter/gnet/pack/pack_line"
	"github.com/MaxnSter/gnet/util"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_other"
	"github.com/MaxnSter/network_practice/sudoku"
)

func main() {
	fileName := flag.String("f", "", "input file")
	isLocal := flag.Bool("l", false, "run batch as local")
	clientNum := flag.Int("c", 1, "clientNum")
	addr := flag.String("addr", "", "sudoku server address")

	flag.Parse()

	if *fileName == "" {
		flag.Usage()
		os.Exit(1)
	}
	f, err := os.OpenFile(*fileName, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if *isLocal {
		runLocal(f)
		return
	}

	if !*isLocal {
		if *addr == "" {
			flag.Usage()
			os.Exit(1)
		}
		runClient(*clientNum, *addr, f)
		return
	}

	//f, err := os.OpenFile("/home/maxnster/study/golang/src/github.com/MaxnSter/network_practice/sudoku/sudoku_client/test1",
	//	os.O_RDONLY, 0644)
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()
	//runClient(1, "127.0.0.1:2007", f)
}

func runLocal(rd io.Reader) {

	scan := bufio.NewScanner(rd)
	count := 0
	start := time.Now()

	for scan.Scan() {
		sudoku.Solve(scan.Text())
		count++
	}

	elapsed := time.Now().Sub(start)
	fmt.Printf("%f seconds, %d total sudoku, %f us per sudoku\n", elapsed.Seconds(),
		count,
		float64(elapsed.Nanoseconds())/1000/float64(count))

}

func runClient(clinetNum int, addr string, rd io.Reader) {

	gStart = time.Now()
	gClientNum = clinetNum

	scan := bufio.NewScanner(rd)
	input := make([]string, 0)
	for scan.Scan() {
		input = append(input, scan.Text())
	}

	wg := sync.WaitGroup{}
	for i := 0; i < clinetNum; i++ {
		wg.Add(1)
		go func(i int) {
			NewSudokuClient(i, input).StartAndRun(addr)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

type baseSudokuClient struct {
	*net.TcpClient

	id    int
	input []string
	count int
	start time.Time
	end   time.Time
}

func NewSudokuClient(id int, input []string) *baseSudokuClient {
	return &baseSudokuClient{input: input, id: id}
}

func (s *baseSudokuClient) StartAndRun(addr string) {
	callback := gnet.NewCallBackOption(gnet.WithOnConnectCB(s.onConnect))
	gnetOption := &gnet.GnetOption{Packer: "line", Coder: "byte", WorkerPool: "poolRaceOther"}

	s.TcpClient = gnet.NewClient(addr, callback, gnetOption, s.onMessage)
	s.TcpClient.StartAndRun()
}

func (s *baseSudokuClient) onConnect(session *net.TcpSession) {
	s.start = time.Now()
	for _, req := range s.input {
		session.Send(req)
	}
}

func (s *baseSudokuClient) onMessage(ev iface.Event) {
	switch msg := ev.Message().(type) {
	case []byte:
		if sudoku.Vertify(util.BytesToString(msg)) {
			s.count++

			if s.count == len(s.input) {
				s.end = time.Now()
				s.TcpClient.Stop()
				done("client"+strconv.Itoa(s.id), s.count, s.start, s.end)
			}
		}
	}
}

var gClientNum int
var gFinished int
var gStart time.Time

func done(name string, reqCount int, from, to time.Time) {
	elapsed := to.Sub(from)
	fmt.Printf("%s done, %f seconds, %d total request, %f per us\n",
		name, elapsed.Seconds(), reqCount, float64(elapsed.Nanoseconds())/1000/float64(reqCount))

	gFinished++
	if gFinished == gClientNum {
		totalElapsed := time.Now().Sub(gStart)

		fmt.Printf("all client done, %f secnods, %f per client\n", totalElapsed.Seconds(),
			totalElapsed.Seconds()/float64(gClientNum))
	}

}
