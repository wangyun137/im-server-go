package component

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"ud-server/define"
	"ud-server/model"
)

type TcpComponent struct {
	Conn net.Listener
	IP   string
	Port uint16
}

var Tcp *TcpComponent

func NewTcpComponent() *TcpComponent {
	return &TcpComponent{}
}

func GetTcpComponent() *TcpComponent {
	if Tcp == nil {
		Tcp = NewTcpComponent()
	}

	return Tcp
}

func (this *TcpComponent) Init(IP string, Port uint16) {
	this.IP = IP
	this.Port = Port
	var err error

	this.Conn, err = net.Listen("tcp", this.IP+":"+strconv.Itoa(int(this.Port)))
	if err != nil {
		fmt.Println("Tcp Component Error: can not listen")
		return
	}
	fmt.Println("Tcp Component Ok - listen")
}

func (this *TcpComponent) Start() {
	go this.Handle()
}

func (this *TcpComponent) Handle() {
	defer func() {
		if this.Conn != nil {
			this.Conn.Close()
		}
	}()

	for {
		Conn, err := this.Conn.Accept()
		if err != nil {
			continue
		}
		go this.ConnHandle(&Conn)
	}
}

func (this *TcpComponent) ConnHandle(Conn *net.Conn) {
	defer func() {
		(*Conn).Close()
	}()
	//LEN_LENGTH = 2
	buff := make([]byte, model.LEN_LENGTH)
	msg := &model.Operation{}

	for {
		n, err := (*Conn).Read(buff)
		if n == 0 || err != nil {
			return
		}

		length := binary.BigEndian.Uint16(buff)
		if length <= 0 {
			return
		}

		databuff := make([]byte, length)

		n, err = (*Conn).Read(databuff[model.LEN_LENGTH:])
		if n == 0 || err != nil {
			return
		}

		copy(databuff[:model.LEN_LENGTH], buff)

		err = msg.Decode(databuff)
		if err != nil {
			continue
		}

		switch msg.Type {
		case define.DELETE:
			this.DeleteHandle(msg)
		}
	}
}

func (this *TcpComponent) DeleteHandle(msg *model.Operation) {
	GetDeleteComponent().Push(msg)
}
