package main

import (
	"connect-server/model"
	"connect-server/qserver"
	"connect-server/utils/config"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config := config.NewConfig()

	err := config.Read("conf/server.conf")
	if err != nil {
		err = fmt.Errorf("Main Error: %v", err)
		fmt.Println(err.Error())
		return
	}
	var IP string
	var Port int
	var UDPPort int
	var DispatchIP string
	var TokenPort int

	IP, err = config.GetString("ServerIP")
	if err != nil {
		err = fmt.Errorf("Main Error: %v", err)
		fmt.Println(err.Error())
		return
	}

	Port, err = config.GetInt("ServerPort")
	if err != nil {
		err = fmt.Errorf("Main Error: %v", err)
		fmt.Println(err.Error())
		return
	}

	UDPPort, err = config.GetInt("UDPServerPort")
	if err != nil {
		err = fmt.Errorf("Main Error: %v", err)
		fmt.Println(err.Error())
		return
	}

	DispatchIP, err = config.GetString("DispatchIP")
	if err != nil {
		err = fmt.Errorf("Main Error: %v", err)
		fmt.Println(err.Error())
		return
	}

	TokenPort, err = config.GetInt("TokenPort")
	if err != nil {
		err = fmt.Errorf("Main Error: %v", err)
		fmt.Println(err.Error())
		return
	}

	userCache := model.NewUserCache()
	udpServer := qserver.NewUDPServer(IP, UDPPort, userCache)
	if udpServer == nil {
		fmt.Println("Start UDPServer error")
		return
	}
	go udpServer.Start()

	tcpServer := qserver.NewQServer(IP, Port, userCache, DispatchIP, TokenPort)
	tcpServer.Start()

}
