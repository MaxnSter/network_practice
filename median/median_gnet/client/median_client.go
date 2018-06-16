package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/worker_pool"
	"github.com/MaxnSter/network_practice/median/kth"
	"github.com/MaxnSter/network_practice/median/median_gnet"

	_ "github.com/MaxnSter/gnet/codec/codec_msgpack"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_type_length_value"
	_ "github.com/MaxnSter/gnet/net/tcp"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

type collector struct {
	connections []gnet.NetSession

	wg sync.WaitGroup

	max, min, avg   int
	sum, count, kth int

	// search result
	//FIXME store as a context
	smaller, same int
}

func (c *collector) onMessage(ev gnet.Event) {
	switch msg := ev.Message().(type) {
	case *median_gnet.QueryResponse:
		c.onQueryResponse(msg)
	case *median_gnet.SearchResponse:
		c.onSearchResponse(msg)
	}
}

func (c *collector) onQueryResponse(response *median_gnet.QueryResponse) {
	if response.Max > c.max {
		c.max = response.Max
	}

	if response.Min < c.min {
		c.min = response.Min
	}

	c.count += response.Count
	c.sum += response.Sum
	c.wg.Done()
}

func (c *collector) onConnected(s gnet.NetSession) {
	c.connections = append(c.connections, s)
	c.wg.Done()
}

func (c *collector) onSearchResponse(response *median_gnet.SearchResponse) {
	c.same += response.Same
	c.smaller += response.Smaller
	c.wg.Done()
}

func (c *collector) Generate(count, min, max int) {
	req := &median_gnet.GenerateRequest{Id: median_gnet.IdGenerateRequest}
	req.Count = count
	req.Min = min
	req.Max = max

	for _, s := range c.connections {
		c.wg.Add(1)
		s.Send(req)
	}

	c.wg.Wait()
	c.avg = c.sum / c.count

	fmt.Printf("all genrate done\nmax:%d, min:=%d, sum=%d, count=%d, avg=%d\n",
		c.max, c.min, c.sum, c.count, c.avg)
}

func (c *collector) FindKth(k int) {
	//FIXME assert not in pool
	kMax, result := kth.FindKth(c.search, k, c.count, c.min, c.max)

	if !result {
		panic("find kth failed")
	}
	fmt.Printf("%d max from server:%d\n", k, kMax)
}

func (c *collector) search(guest int) (smaller, same int) {
	//FIXME
	c.same = 0
	c.smaller = 0

	req := &median_gnet.SearchRequest{Id: median_gnet.IdSearchRequest}
	req.Guess = guest

	for _, s := range c.connections {
		c.wg.Add(1)
		s.Send(req)
	}

	c.wg.Wait()
	return c.smaller, c.same
}

func (c *collector) Connect(addrs ...string) {
	//FIXME assert not in pool
	Pool := worker_pool.MustGetWorkerPool("poolRaceOther")
	for i, addr := range addrs {
		m := gnet.NewModule()
		m.SetSharePool(Pool)
		m.SetPacker("tlv")
		m.SetCoder("msgpack")

		//FIXME onClose callback
		o := gnet.NewOperator(c.onMessage)
		o.SetOnConnected(c.onConnected)

		c.wg.Add(1)
		client := gnet.NewNetClient("tcp", "medianClient#"+strconv.Itoa(i), m, o)
		go client.Connect(addr)
	}
	c.wg.Wait()
}

func (c *collector) Run(min, max, count, kth int) {
	//FIXME assert not in pool
	c.Generate(count, min, max)

	logger.Infoln("start find kth:", kth)
	c.FindKth(kth)
}

func NewCollector() *collector {
	//FIXME init max,min,sum...
	return &collector{}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("[median_client] addr1 aadr2...")
		os.Exit(1)
	}

	c := NewCollector()
	c.Connect(os.Args[1:]...)
	c.Run(0, 200, 50, 10)
}
