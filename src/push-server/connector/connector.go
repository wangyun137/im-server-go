package connector

import (
	"errors"
	"net"
	"push-server/component"
	"strconv"
)

type Connector struct {
	component.Handler
	Listener net.Listner
	IP       string
	Port     int
	Count    int64
}

func NewConnector() *Connector {
	return &Connector{}
}

func (this *Connector) Init(IP string, Port int) {
	this.IP = IP
	this.Port = Port
}

func (this *Connector) Start() error {
	var err error
	this.Listener, err = net.Listen("tcp", this.IP+":"+strconv.Itoa(this.Port))
	if err != nil {
		err = errors.New("Connector Start Listen fail")
		return err
	}
	var registerComponent *component.RegisterComponent
	for {
		registerComponent = component.GetRegisterComponent()
		if registerComponent != nil {
			break
		}
	}
	go this.accept(registerComponent)
	return nil
}

func (this *Connector) accept(registerComponent *component.RegisterComponent) {
	for {
		Conn, err := this.Listener.Accept()
		if err != nil {
			continue
		}
		registerComponent.PushBack(&Conn)
	}
}
