package component

import (
	"fmt"
	"net"
	"ud-server/define"
	"ud-server/model"
)

type RegisterComponent struct {
	Messages              chan *model.Register
	RegisterResponseCount int
	BuffLength            int
}

var Register *RegisterComponent

func NewRegisterComponent() *RegisterComponent {
	return &RegisterComponent{}
}

func GetRegisterComponent() *RegisterComponent {
	if Register == nil {
		Register = NewRegisterComponent()
	}

	return Register
}

func (this *RegisterComponent) Init(RegisterResponseCount int, BuffLength int) {
	this.BuffLength = BuffLength
	this.RegisterResponseCount = RegisterResponseCount
	if this.Messages == nil {
		this.Messages = make(chan *model.Register, this.BuffLength)
	}
}

func (this *RegisterComponent) Start() {
	go this.Handle()
}

func (this *RegisterComponent) Handle() {
	for {
		this.MsgHandle(this.Pop())
	}
}

func (this *RegisterComponent) MsgHandle(request *model.Register) {

	audio := request.Request

	Conn := &net.UDPAddr{
		IP:   request.IP,
		Port: int(request.Port),
	}

	err := model.GetUsers().Add(request.FromUuid, request.FromDeviceId, request.DestUuid, request.DestDeviceId, request.IP, request.Port, Conn)
	if err != nil {
		err = model.GetUsers().Update(request.FromUuid, request.FromDeviceId, request.DestUuid, request.DestDeviceId, request.IP, request.Port, Conn)
		if err != nil {
			audio.Type = define.REGISTER_ERR
			fmt.Println("Register Component MsgHandle - register ", request.FromUuid, request.FromDeviceId)
		} else {
			audio.Type = define.REGISTER_OK
			fmt.Println("Register Component MsgHandle Ok - register ", request.FromUuid, request.FromDeviceId)
		}
	} else {
		audio.Type = define.REGISTER_OK
		fmt.Println("Register Component MsgHandle Ok - register ", request.FromUuid, request.FromDeviceId)
	}

	if user := model.GetUsers().Get(request.DestUuid, request.DestDeviceId); user != nil && user.Conn != nil &&
		user.BindUuid == request.FromUuid && user.BindDeviceId == request.FromDeviceId {

		buff, err := audio.Encode()
		if err != nil {
			err = fmt.Errorf("Register Component MsgHandle Error:%v", err)
			fmt.Println(err.Error())
			return
		}

		for i := 0; i < this.RegisterResponseCount; i++ {
			GetReceiveComponent().Write(buff, request.IP, request.Port)
			GetReceiveComponent().Write(buff, user.IP, user.Port)
		}
	}
}

func (this *RegisterComponent) Push(msg *model.Register) {
	this.Messages <- msg
}

func (this *RegisterComponent) Pop() *model.Register {
	return <-this.Messages
}
