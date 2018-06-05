package sudoku

import (
	"strings"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/util"
	"github.com/laurentlp/sudoku-solver/solver"
)

func Solve(grid string) string {
	if result, err := solver.Solve(grid); err != nil {
		panic(err)
	} else {
		return result
	}
}

func Vertify(answer string) bool{
	return true
}

type sudokuServer struct {
	gnetOption *gnet.GnetOption
	gnetCallback *gnet.CallBackOption
	*net.TcpServer
}

func NewSudokuServer(option *gnet.GnetOption, callBackOption *gnet.CallBackOption, addr string) *sudokuServer{
	s := &sudokuServer{gnetOption:option, gnetCallback:callBackOption}
	s.TcpServer = gnet.NewServer(addr, s.gnetCallback, s.gnetOption, s.onMessage)
	return s
}

func (s *sudokuServer) onMessage(ev iface.Event) {
	switch msg := ev.Message().(type) {
	case []byte:
		var reqId string
		var reqContent string
		req := util.BytesToString(msg)

		if idx := strings.Index(req, ":"); idx != -1 {
			reqId = req[:idx]
			reqContent = req[idx+1:]
		} else {
			reqContent = req
		}

		 result := Solve(reqContent)
		 if reqId != "" {
		 	result = reqId + ":" + result
		 }
		 ev.Session().Send(result)
	}
}

