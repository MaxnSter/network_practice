package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/network_practice/soduku"
)

func main() {

	fileName := flag.String("f", "", "input file")
	isLocal := flag.Bool("l", false, "run batch as local")
	clientNum := flag.Int("c", 1, "clientNum")
	addr := flag.String("addr", "", "sudoku server address")

	flag.Parse()

	if *isLocal {
		if *fileName == "" {
			flag.Usage()
			os.Exit(1)
		}

		f, err := os.OpenFile("./network_practice/soduku/"+*fileName, os.O_RDONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		runLocal(f)
		return
	}

	if !*isLocal {
		if *addr == "" {
			flag.Usage()
			os.Exit(1)
		}

		runClient(*clientNum, *addr)
		return
	}

}

func runLocal(rd io.Reader) {
	scan := bufio.NewScanner(rd)
	count := 0
	start := time.Now()

	for scan.Scan() {
		soduku.Solve(scan.Text())
		count++
	}

	elapsed := time.Now().Sub(start)
	fmt.Printf("%f seconds, %d total sudoku, %f us per sudoku\n", elapsed.Seconds(),
		count,
		float64(elapsed.Nanoseconds())/float64(count))
}

func runClient(clinetNum int, addr string) {
	wg := sync.WaitGroup{}
	for i := 0; i < clinetNum; i++ {
		wg.Add(1)
		go func() {
			NewSudokuClient().StartAndRun(addr)
			wg.Done()
		}()
	}

	wg.Wait()
}

type sudokuClient struct {
	*net.TcpClient
}

func NewSudokuClient() *sudokuClient {
	return &sudokuClient{}
}

func (s *sudokuClient) StartAndRun(addr string) {
	callback := gnet.NewCallBackOption(gnet.WithOnConeectCB(s.onConnect))
	gnetOption := gnet.NewGnetOption(gnet.WithWorkerPool("poolRaceOther"))

	s.TcpClient = gnet.NewClient(addr, callback, gnetOption, s.onMessage)
	s.TcpClient.StartAndRun()
}

func (s *sudokuClient) onConnect(session *net.TcpSession) {

}

func (s *sudokuClient) onMessage(ev iface.Event) {

}
