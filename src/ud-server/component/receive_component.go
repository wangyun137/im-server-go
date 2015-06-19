package component

import (
	"net"
	"strconv"
	"strings"
	"ud-server/define"
	"ud-server/model"
)

type ReceiveComponent struct {
	IP   string
	Port uint16
	Conn *net.UDPConn
}

var Receive *ReceiveComponent

func NewReceiveComponent() *ReceiveComponent {
	return &ReceiveComponent{}
}

func GetReceiveComponent() *ReceiveComponent {
	if Receive == nil {
		Receive = NewReceiveComponent()
	}

	return Receive
}

func (this *ReceiveComponent) Init(IP string, Port uint16) {
	this.IP = IP
	this.Port = Port
}

func (this *ReceiveComponent) Start() {
	var err error

	this.Conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(this.IP),
		Port: int(this.Port),
	})

	if err != nil {
		return
	}
	defer this.Conn.Close()

	this.Handle()
}

func (this *ReceiveComponent) Handle() {
	buff := make([]byte, 1400*define.B)
	for {
		n, remoteaddr, err := this.Conn.ReadFromUDP(buff)
		if err != nil {
			continue
		}

		this.Unpacket(buff[:n], remoteaddr)
	}
}

func (this *ReceiveComponent) Unpacket(buff []byte, remoteaddr *net.UDPAddr) {
	if len(buff) != 0 {
		msgaudio := model.Request{}
		err := msgaudio.Decode(buff)
		if err != nil {
			return
		}

		switch msgaudio.Type {
		case define.SEND:
			GetSendComponent().Push(&msgaudio)
		case define.REGISTER:
			this.RegisterMsgHandle(&msgaudio, remoteaddr)
		default:
		}
	}
}

func (this *ReceiveComponent) RegisterMsgHandle(msg *model.Request, remoteaddr *net.UDPAddr) {
	register := model.Register{}
	register.Request = *msg
	register.IP = remoteaddr.IP
	register.Port = uint16(remoteaddr.Port)

	GetRegisterComponent().Push(&register)
}

func (this *ReceiveComponent) Write(buff []byte, IP net.IP, Port uint16) (int, error) {
	return this.Conn.WriteToUDP(buff, &net.UDPAddr{IP: IP, Port: int(Port)})
}

func (this *ReceiveComponent) GetIPAndPort() (net.IP, int) {
	addr := this.Conn.LocalAddr().String()
	addrs := strings.Split(addr, ":")
	if len(addrs) != 2 {
		return nil, 0
	}

	port, err := strconv.Atoi(addrs[1])
	if err != nil {
		return nil, 0
	}

	return net.ParseIP(addrs[0]), port
}
