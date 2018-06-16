package main

import (
	"flag"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/network_practice/sudoku"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_norace"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_self"
)

func main() {

	port := flag.String("p", "2007", "listen port")
	flag.Parse()

	// single event loop, 单个goroutine处理所有请求,无法利用多核计算
	option := &gnet.GnetOption{Packer: "line", Coder: "byte", WorkerPool: "poolRaceOther"}
	s := sudoku.NewSudokuServer(option, &gnet.CallBackOption{}, "0.0.0.0:"+*port)
	s.StartAndRun()
}
