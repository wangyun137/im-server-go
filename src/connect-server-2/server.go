package main

import (
	"fmt"
	"push-server/api"
	"push-server/protocol"
	"runtime"
	"strconv"
)

const (
	QSERVER_UUID = "000000000000000000000000000" + strconv.Itoa(10000)
)

type Server struct {
	/*nil*/
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	pusher, err := api.NewPusher(uint16(protocol.N_QSERVER), QSERVER_UUID, 100)
	if err != nil {
		err = fmt.Errorf("Connect-Server main Error: %v", err)
		fmt.Println(err.Error())
		return
	}

	err = pusher.RegisterPushReceiver(uint16(protocol.N_QSERVER))
}

func (server *Server) Receive(msg *protocol.Message) error {
	tcpHandler := NewTCPHandler()
}
