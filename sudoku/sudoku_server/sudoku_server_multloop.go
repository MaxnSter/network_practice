package main

import (
	"flag"

	"github.com/MaxnSter/gnet"
	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_norace"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_self"
	"github.com/MaxnSter/network_practice/sudoku"
)

func main() {

	port := flag.String("p", "2007", "listen port")
	flag.Parse()

	// single event loop, 多个goroutine处理所有请求, 每个client的所有请求对应固定goroutine,
	// 多个client时可利用多核,单个客户端时仍不可利用多核
	m := gnet.NewModule(gnet.WithCoder("byte"), gnet.WithPacker("line"),
		gnet.WithPool("poolRaceSelf"))
	s := sudoku.NewSudokuServer(m, "0.0.0.0:"+*port)
	s.Serve()
}
