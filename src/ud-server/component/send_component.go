package component

import (
	"fmt"
	"ud-server/model"
)

type SendComponent struct {
	Messages   chan *model.Request
	BuffLength int
}

var Send *SendComponent

func NewSendComponent() *SendComponent {
	return &SendComponent{}
}

func GetSendComponent() *SendComponent {
	if Send == nil {
		Send = NewSendComponent()
	}

	return Send
}

//BuffLength = 100000
func (this *SendComponent) Init(BuffLength int) {
	this.BuffLength = BuffLength
	if this.Messages == nil {
		this.Messages = make(chan *model.Request, this.BuffLength)
	}
}

func (this *SendComponent) Start() {
	go this.Handle()
}

func (this *SendComponent) Handle() {
	for {
		this.MsgHandle(this.Pop())
	}
}

func (this *SendComponent) MsgHandle(audio *model.Request) {
	if user := model.GetUsers().Get(audio.DestUuid, audio.DestDeviceId); user != nil && user.Conn != nil && user.BindUuid == audio.FromUuid && user.BindDeviceId == audio.FromDeviceId {
		audio.DestUuid, audio.FromUuid = audio.FromUuid, audio.DestUuid
		audio.DestDeviceId, audio.FromDeviceId = audio.FromDeviceId, audio.DestDeviceId

		buff, err := audio.Encode()
		if err == nil {
			n, err := GetReceiveComponent().Write(buff, user.IP, user.Port)
			//	n, err := user.Conn.Write(buff)
			if n == 0 || err != nil {
				fmt.Println("Send Component MsgHandle Error : send", audio.DestUuid, audio.DestDeviceId, err)
				return
			}

			fmt.Println("Send Component MsgHandle Ok :send ", audio.DestUuid, audio.DestDeviceId)
		}
	} else {
		fmt.Println("Send Component MsgHandle Error :not found user", audio.DestUuid, audio.DestDeviceId)
	}
}

func (this *SendComponent) Push(msg *model.Request) {
	this.Messages <- msg
}

func (this *SendComponent) Pop() *model.Request {
	return <-this.Messages
}
