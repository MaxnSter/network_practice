package sudoku

import (
	"fmt"
	"math"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"syscall"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/util"
)

func Solve(grid string) string {

	for i := 0; i < math.MaxInt16<<3; i++ {
	}

	return grid
	// 这个库有问题,有goroutine leak
	// 用上面循环模拟cpu计算
	//if result, err := solver.Solve(grid); err != nil {
	//	panic(err)
	//} else {
	//	sb := new(strings.Builder)
	//	for _, v := range result {
	//		sb.WriteString(v)
	//	}
	//	return sb.String()
	//}
}

func Vertify(answer string) bool {
	return true
}

type sudokuServer struct {
	gnetOption   *gnet.GnetOption
	gnetCallback *gnet.CallBackOption
	*net.TcpServer
}

func NewSudokuServer(option *gnet.GnetOption, callBackOption *gnet.CallBackOption, addr string) *sudokuServer {
	s := &sudokuServer{gnetOption: option, gnetCallback: callBackOption}
	s.TcpServer = gnet.NewServer(addr, s.gnetCallback, s.gnetOption, s.onMessage)

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()
	fmt.Println("pid = ", syscall.Getpid())

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
