package main

import (
	"fmt"
	"push-server/buffpool"
	"push-server/server"
	"runtime"
	"time"
)

func main() {
	pusherServer := server.PushServer{}
	err := pusherServer.ReadConf("conf/push-server.conf")
	if err != nil {
		fmt.Println("PushServer Error:" + err.Error())
		return
	}
	runtime.GOMAXPROCS(pusherServer.GoMaxProcs)

	connect, err := pusherServer.Init()
	if err != nil {
		fmt.Println("PushServer Error:" + err.Error())
		return
	}

	err = pusherServer.Start(connect)
	if err != nil {
		fmt.Println("PushServer Error:" + err.Error())
		return
	}

	go func() {
		for {
			buffpool.GetBuffPool().PrintServiceConnections()
			buffpool.GetBuffPool().PrintClientConnections()
			buffpool.GetBuffPool().PrintTmpClientConnections()
			time.Sleep(5 * time.Second)
		}
	}()

	signal := make(chan bool, 0)
	<-signal
}
