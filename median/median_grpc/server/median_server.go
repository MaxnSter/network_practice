package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"syscall"
	"time"

	pb "github.com/MaxnSter/network_practice/median/median_grpc"
	"google.golang.org/grpc"
)

//go:generate protoc -I .. ../median.proto --go_out=plugins=grpc:..
type medianServer struct {
	data      []int64
	machineId int
}

func NewMedianServer(machineId int) *medianServer {
	seed := time.Now().UnixNano()<<16 | int64(machineId)
	rand.Seed(seed)
	return &medianServer{machineId: machineId}
}

func (ms *medianServer) output(writer io.Writer) {
	wr := bufio.NewWriter(writer)
	for i, v := range ms.data {
		wr.WriteString(strconv.FormatInt(v, 10))
		if i+1%10 == 0 {
			wr.WriteByte('\n')
		} else {
			wr.WriteByte('\t')
		}
	}
	wr.WriteByte('\n')
	wr.Flush()
}

func (ms *medianServer) Generate(ctx context.Context, req *pb.GenerateRequest) (*pb.QueryResponse, error) {
	// i don't block here, ignore ctx
	var sum int64
	ms.data = make([]int64, req.Count)
	for i := 0; i < len(ms.data); i++ {
		ms.data[i] = rand.Int63n(req.Max-req.Min) + req.Min
		sum += ms.data[i]
	}

	sort.Slice(ms.data, func(i, j int) bool {
		return ms.data[i] < ms.data[j]
	})

	// record data
	go ms.output(os.Stdout)

	return &pb.QueryResponse{
		Sum:   sum,
		Count: int64(len(ms.data)),
		Min:   ms.data[0],
		Max:   ms.data[len(ms.data)-1],
	}, nil
}

func (ms *medianServer) Search(ctx context.Context, req *pb.SeachRequest) (*pb.SearchResponse, error) {
	response := &pb.SearchResponse{}
	for _, v := range ms.data {
		if v < req.Guess {
			response.Smaller++
		} else if v == req.Guess {
			response.Same++
		} else {
			break
		}
	}
	return response, nil
}

var (
	port = flag.String("p", "2007", "listen port")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMedianServer(grpcServer, NewMedianServer(syscall.Getpid()))
	fmt.Printf("serve at %s\n", lis.Addr().String())
	grpcServer.Serve(lis)
}
