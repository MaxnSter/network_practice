package main

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/network_practice/sudoku"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker/worker_session_norace"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_other"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_self"
)

func main() {
	// single event loop, 单个goroutine处理所有请求,无法利用多核计算
	option := &gnet.GnetOption{Packer: "line", Coder: "byte", WorkerPool: "poolRaceOther"}
	s := sudoku.NewSudokuServer(option, &gnet.CallBackOption{}, "127.0.0.1:2007")
	s.StartAndRun()
}
