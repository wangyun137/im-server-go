package component

import (
	"fmt"
	"ud-server/model"
)

type DeleteComponent struct {
	Messages   chan *model.Operation
	BuffLength int
}

var Delete *DeleteComponent

func NewDeleteComponent() *DeleteComponent {
	return &DeleteComponent{}
}

func GetDeleteComponent() *DeleteComponent {
	if Delete == nil {
		Delete = NewDeleteComponent()
	}

	return Delete
}

func (this *DeleteComponent) Init(BuffLength int) {
	this.BuffLength = BuffLength
	if this.Messages == nil {
		this.Messages = make(chan *model.Operation, this.BuffLength)
	}
}

func (this *DeleteComponent) Start() {
	go this.Handle()
}

func (this *DeleteComponent) Handle() {
	for {
		this.MsgHandle(this.Pop())
	}
}

func (this *DeleteComponent) MsgHandle(msg *model.Operation) {
	for _, v := range msg.List {

		err := model.GetUsers().Remove(v.Uuid, v.DeviceId)
		if err != nil {
			continue
		}
		fmt.Println("Ok - delete uuid ", v.Uuid, v.DeviceId)
	}
}

func (this *DeleteComponent) Push(msg *model.Operation) {
	this.Messages <- msg
}

func (this *DeleteComponent) Pop() *model.Operation {
	return <-this.Messages
}
