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
	m := gnet.NewModule(gnet.WithCoder("byte"), gnet.WithPacker("line"),
		gnet.WithPool("poolNoRace"))
	s := sudoku.NewSudokuServer(m, "0.0.0.0:"+*port)
	s.Serve()
}
