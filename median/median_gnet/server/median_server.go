package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/network_practice/median/median_gnet"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_type_length_value"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

type medianServer struct {
	gnet.NetServer

	data      []int
	machineId int
}

func (ms *medianServer) onQuery(s gnet.NetSession) {
	respone := &median_gnet.QueryResponse{
		Id:    median_gnet.IdQueryResponse,
		Count: len(ms.data),
		Min:   ms.data[0],
		Max:   ms.data[len(ms.data)-1],
	}

	for _, v := range ms.data {
		respone.Sum += v
	}

	s.Send(respone)
}

func (ms *medianServer) onGenerate(min, max, count int, s gnet.NetSession) {
	ms.generate(min, max, count, os.Stdout)
	ms.onQuery(s)
}

func (ms *medianServer) onSearch(guess int, s gnet.NetSession) {
	response := &median_gnet.SearchResponse{Id: median_gnet.IdSearchResponse}
	for i := 0; i < len(ms.data); i++ {
		if ms.data[i] < guess {
			response.Smaller++
		} else if ms.data[i] == guess {
			response.Same++
		} else {
			break
		}
	}
	s.Send(response)
}

func (ms *medianServer) generate(min, max, count int, write io.Writer) {
	ms.data = make([]int, count)
	for i := 0; i < count; i++ {
		ms.data[i] = rand.Intn(max-min) + min
	}
	sort.Ints(ms.data)

	if write == nil {
		return
	}

	wr := bufio.NewWriter(write)
	for i := 0; i < len(ms.data); i++ {
		wr.WriteString(fmt.Sprintf("%d\t", ms.data[i]))
		if (i+1)%10 == 0 {
			wr.WriteByte('\n')
		}
	}
	wr.Flush()
}

func (ms *medianServer) Run() {
	ms.Serve()
}

func NewMedianServer(machineId int, port string) *medianServer {
	// 0 < machineId <= 65535
	seed := int64(time.Now().UnixNano()<<16 | int64(machineId))
	logger.Infoln("seed:", seed)
	rand.Seed(seed)

	ms := &medianServer{machineId: machineId}

	module := gnet.NewModule(gnet.WithCoder("msgpack"), gnet.WithPacker("tlv"),
		gnet.WithPool("poolRaceOther"))
	operator := gnet.NewOperator(ms.onMessage)
	ms.NetServer = gnet.NewNetServer("tcp", "median", module, operator)

	if err := ms.Listen(":" + port); err != nil {
		panic(err)
	}
	return ms
}

func (ms *medianServer) onMessage(ev gnet.Event) {
	switch msg := ev.Message().(type) {
	case *median_gnet.GenerateRequest:
		ms.onGenerate(msg.Min, msg.Max, msg.Count, ev.Session())
	case *median_gnet.SearchRequest:
		ms.onSearch(msg.Guess, ev.Session())
	default:
		logger.Errorln("unknown msg type")
		ev.Session().Stop()
	}
}

func main() {
	gen := flag.Bool("g", false, "gen without request")
	port := flag.String("p", "2007", "server address")
	flag.Parse()

	//FIXME
	medianServer := NewMedianServer(syscall.Getpid(), *port)
	if *gen {
		medianServer.generate(0, 100, 30, os.Stdout)
	}
	medianServer.Run()
}
