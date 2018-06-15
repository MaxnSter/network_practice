package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/MaxnSter/network_practice/median/kth"
	pb "github.com/MaxnSter/network_practice/median/median_grpc"
	"google.golang.org/grpc"
)

type collector struct {
	connections []pb.MedianClient
	wg          sync.WaitGroup

	min, max   int64
	count, sum int64
}

func (c *collector) Connect(addrs ...string) {
	connCh := make(chan *grpc.ClientConn, len(addrs))
	for _, addr := range addrs {
		go func(addr string) {
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			connCh <- conn
		}(addr)
	}

	for i := 0; i < len(addrs); i++ {
		conn := <-connCh
		c.connections = append(c.connections, pb.NewMedianClient(conn))
	}
	close(connCh)
}

func (c *collector) reset() {
	c.min = math.MaxInt64
	c.max = math.MinInt64
	c.count = 0
	c.sum = 0
}

func (c *collector) Generate(min, max, count int64) {
	c.reset()
	req := &pb.GenerateRequest{Count: count, Max: max, Min: min}
	resCh := make(chan *pb.QueryResponse, len(c.connections))

	for _, conn := range c.connections {

		conn := conn
		go func() {
			response, err := conn.Generate(context.Background(), req)
			if err != nil {
				panic(err)
			}

			resCh <- response
		}()

	}

	for i := 0; i < len(c.connections); i++ {
		response := <-resCh
		c.sum += response.Sum
		c.count += response.Count
		if c.min > response.Min {
			c.min = response.Min
		}
		if c.max < response.Max {
			c.max = response.Max
		}
	}
	close(resCh)

	fmt.Printf("min:%d, max:%d, count:%d, sum:%d, avg:%d\n",
		c.min, c.max, c.count, c.sum, c.sum/c.count)
}

func (c *collector) FindKthMax(k int) (int, bool) {
	return kth.FindKth(c.search, k, int(c.count), int(c.min), int(c.max))
}

func (c *collector) search(guest int) (smaller, same int) {
	req := &pb.SeachRequest{Guess: int64(guest)}
	resCh := make(chan *pb.SearchResponse, len(c.connections))

	for _, conn := range c.connections {
		conn := conn
		go func() {
			res, err := conn.Search(context.Background(), req)
			if err != nil {
				panic(err)
			}

			resCh <- res
		}()
	}

	for i := 0; i < len(c.connections); i++ {
		res := <-resCh
		smaller += int(res.Smaller)
		same += int(res.Same)
	}
	close(resCh)

	return smaller, same
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("[median_client] addr1 aadr2...")
		os.Exit(1)
	}

	c := &collector{}
	c.Connect(os.Args[1:]...)
	fmt.Println("all connected")
	c.Generate(100, 200, 50)
	fmt.Println(c.FindKthMax(40))
}
