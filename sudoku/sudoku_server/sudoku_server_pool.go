package main

import (
	"flag"

	"github.com/MaxnSter/gnet"
	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_norace"
	"github.com/MaxnSter/network_practice/sudoku"
)

func main() {

	port := flag.String("p", "2007", "listen port")
	flag.Parse()

	// single event loop, goroutine pool处理所有请求,每个请求对应不同goroutine
	// 单个或多个client时可利用多核
	option := &gnet.GnetOption{Packer: "line", Coder: "byte", WorkerPool: "poolNoRace"}
	s := sudoku.NewSudokuServer(option, &gnet.CallBackOption{}, "127.0.0.1:"+*port)
	s.StartAndRun()
}
