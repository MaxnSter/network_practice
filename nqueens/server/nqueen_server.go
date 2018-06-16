package main

import (
	"context"
	"flag"
	"math/bits"
	"net"
	"time"

	pb "github.com/MaxnSter/network_practice/nqueens/nqueens_proto"
	"google.golang.org/grpc"
)

//go:generate protoc -I ../nqueens_proto ../nqueens_proto/nqueen.proto --go_out=plugins=grpc:../nqueens_proto

const (
	maxQueen = 20
)

type nQueensBackTracker struct {
	count  int
	queens int

	columns      [maxQueen]uint //列掩码
	diagonal     [maxQueen]uint //对角线掩码
	antiDiagonal [maxQueen]uint //斜对角掩码
}

func NewNQueensBackTracker(queens int) *nQueensBackTracker {
	return &nQueensBackTracker{
		queens: queens}
}

// 从第几行开始搜索
func (t *nQueensBackTracker) search(row int) {
	avail := t.columns[row] | t.diagonal[row] | t.antiDiagonal[row]
	// 此时avail上位数为1的位置可放皇后
	avail = ^avail

	for avail > 0 {
		// 找出avail上从最右边的1
		i := uint(bits.TrailingZeros(avail))

		// 找不到位置,回溯
		if i >= uint(t.queens) {
			break
		}

		if row == t.queens-1 {
			// 到了最后一行,说明找到了一个解
			t.count++
		} else {
			var mask uint = 1 << i
			// 下一行设置掩码
			// 下一行对角线,右移
			// 下一行斜对角线,左移
			t.columns[row+1] = t.columns[row] | mask
			t.diagonal[row+1] = (t.diagonal[row] | mask) >> 1
			t.antiDiagonal[row+1] = (t.antiDiagonal[row] | mask) << 1

			// 继续搜索
			t.search(row + 1)
		}

		avail &= avail - 1
	}
}

type NQueensServer struct {
}

func (s *NQueensServer) Solve(ctx context.Context, req *pb.SubProblemRequest) (*pb.SubProblemResponse, error) {
	tStart := time.Now()
	resp := &pb.SubProblemResponse{}

	bt := NewNQueensBackTracker(int(req.Nqueens))
	var m0 uint = 1 << uint(req.FirstNow)
	bt.columns[1] = m0
	bt.diagonal[1] = m0 >> 1
	bt.antiDiagonal[1] = m0 << 1
	bt.search(1)

	resp.Count = int64(bt.count)
	resp.Senconds = time.Now().Sub(tStart).Seconds()
	return resp, nil
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

	rpc := grpc.NewServer()
	pb.RegisterNQueensServiceServer(rpc, &NQueensServer{})
	rpc.Serve(lis)
}
