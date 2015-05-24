package main

import (
	"connect-server/utils/config"
	"fmt"
	"runtime"
	"ud-server/component"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	config := config.NewConfig()

	err := config.Read("conf/uds.conf")
	if err != nil {
		fmt.Println("read conf file error")
	}

	var IP string
	var Port int
	var TCPPort int
	var RegisterResponseCount int
	var BuffLength int

	IP, err = config.GetString("IP")
	if err != nil {
		fmt.Println("parse  IP error")
		return
	}

	Port, err = config.GetInt("Port")
	if err != nil {
		fmt.Println("parse Port error")
		return
	}

	TCPPort, err = config.GetInt("TCPPort")
	if err != nil {
		fmt.Println("parse TCPPort error")
		return
	}
	fmt.Println(TCPPort)

	RegisterResponseCount, err = config.GetInt("RegisterResponseCount")
	if err != nil {
		fmt.Println("Err - parse RegisterResponseCount, set default value 10")
		RegisterResponseCount = 10
	}

	BuffLength, err = config.GetInt("BuffLength")
	if err != nil {
		fmt.Println("Err - parse BuffLength, set default value 100000")
		BuffLength = 100000
	}

	send := component.GetSendComponent()
	send.Init(BuffLength)
	send.Start()

	register := component.GetRegisterComponent()
	register.Init(RegisterResponseCount, BuffLength)
	register.Start()

	deleteComponent := component.GetDeleteComponent()
	deleteComponent.Init(BuffLength)
	deleteComponent.Start()

	tcp := component.GetTcpComponent()
	tcp.Init(IP, uint16(TCPPort))
	tcp.Start()

	receive := component.GetReceiveComponent()
	receive.Init(IP, uint16(Port))
	receive.Start()
}
