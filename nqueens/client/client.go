package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/MaxnSter/network_practice/nqueens/nqueens_proto"
	"google.golang.org/grpc"
)

type nQueensClient struct {
	connections []pb.NQueensServiceClient

	queens int
	count  int64
}

func newNQueensClient(queens int) *nQueensClient {
	return &nQueensClient{queens: queens}
}

func (c *nQueensClient) connect(addrs ...string) {
	connCh := make(chan *grpc.ClientConn, len(addrs))
	for _, addr := range addrs {
		addr := addr
		go func() {
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			connCh <- conn
		}()

	}

	for i := 0; i < len(addrs); i++ {
		c.connections = append(c.connections, pb.NewNQueensServiceClient(<-connCh))
	}
	close(connCh)
}

func (c *nQueensClient) run() {
	reqCount := (c.queens + 1) / 2
	connLen := len(c.connections)
	respCh := make(chan int64, reqCount)

	for i := 0; i < reqCount; i++ {
		i := i
		go func() {
			req := &pb.SubProblemRequest{Nqueens: int32(c.queens), FirstNow: int32(i)}
			resp, err := c.connections[i%connLen].Solve(context.Background(), req)
			if err != nil {
				panic(err)
			}

			fmt.Printf("response for col:%d count:%d seconds:%v\n", i, resp.Count, resp.Senconds)
			if reqCount%2 != 0 && i == reqCount/2 {
				respCh <- resp.Count
			} else {
				respCh <- resp.Count * 2
			}
		}()
	}

	for i := 0; i < reqCount; i++ {
		c.count += <-respCh
	}
	close(respCh)
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("[client] addr1 addr2 ...")
		os.Exit(1)
	}

	client := newNQueensClient(8)
	client.connect(os.Args[1:]...)
	client.run()
	fmt.Printf("%d queens, slove:%d\n", 8, client.count)
}
