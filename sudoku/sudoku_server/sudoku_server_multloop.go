package main

import (
	"github.com/MaxnSter/gnet"
	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker/worker_session_norace"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_other"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_self"
	"github.com/MaxnSter/network_practice/sudoku"
)

func main() {
	// single event loop, 多个goroutine处理所有请求, 每个client的所有请求对应固定goroutine,
	// 多个client时可利用多核,单个客户端时仍不可利用多核
	option := &gnet.GnetOption{Packer: "line", Coder: "byte", WorkerPool: "poolRaceSelf"}
	s := sudoku.NewSudokuServer(option, nil, "127.0.0.1:2007")
	s.StartAndRun()
}
