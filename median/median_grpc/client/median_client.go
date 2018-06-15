package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/MaxnSter/network_practice/median/kth"
	pb "github.com/MaxnSter/network_practice/median/median_grpc"
	"google.golang.org/grpc"
)

type collector struct {
	connections []pb.MedianClient

	min, max   int64
	count, sum int64
}

func (c *collector) Connect(addrs ...string) {
	for _, addr := range addrs {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		c.connections = append(c.connections, pb.NewMedianClient(conn))
	}
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
	for _, conn := range c.connections {
		response, err := conn.Generate(context.Background(), req)
		if err != nil {
			panic(err)
		}

		c.sum += response.Sum
		c.count += response.Count
		if c.min > response.Min {
			c.min = response.Min
		}
		if c.max < response.Max {
			c.max = response.Max
		}
	}

	fmt.Printf("min:%d, max:%d, count:%d, sum:%d, avg:%d\n",
		c.min, c.max, c.count, c.sum, c.sum/c.count)
}

func (c *collector) FindKthMax(k int) (int, bool){
	return kth.FindKth(c.search, k, int(c.count), int(c.min), int(c.max))
}

func (c *collector) search(guest int) (smaller, same int) {
	req := &pb.SeachRequest{Guess: int64(guest)}
	for _, conn := range c.connections {
		res, err := conn.Search(context.Background(), req)
		if err != nil {
			panic(err)
		}
		smaller += int(res.Smaller)
		same += int(res.Same)
	}
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
