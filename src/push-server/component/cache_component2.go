package component

import (
	"encoding/binary"
	"fmt"
	"push-server/protocol"
	snode_component "snode-server/component"
	"time"
	stree_define "tree-server/define"
)

const HEAD_RESET = 0xffff >> 2

type CacheComponent2 struct {
	ServerCacheID uint16
	MsgQueueLen   int
	MsgQueue      chan *MessagePacket
}

var cacheComponent CacheComponent2

func NewCacheComponent2(msgQueueLen, respQueueLen int) *CacheComponent2 {
	cache := &cacheComponent
	cache.MsgQueueLen = msgQueueLen
	return cache
}

func GetCacheComponent2() *CacheComponent2 {
	return &cacheComponent
}

func (this *CacheComponent2) Init2() {
	this.MsgQueue = make(chan *MessagePacket, this.MsgQueueLen)
	return
}

func (this *CacheComponent2) Start2() {
	go this.HandleMsg2()
	return
}

func (this *CacheComponent2) MsgPush2(msg *protocol.Message, flag bool) {
	if msg == nil {
		fmt.Println("Got a nil msg pointer")
		return
	}
	packet := &MessagePacket{
		Message: msg,
		Flag:    flag,
	}
	this.MsgQueue <- packet
}

func (this *CacheComponent2) MsgPop2() *MessagePacket {
	packet := <-this.MsgQueue
	return packet
}

func (this *CacheComponent2) HandleMsg2() {
	for {
		packet := this.MsgPop2()
		if packet == nil {
			fmt.Println("CacheComponent2 HandleMsg2 Error:Got a nil msg")
			continue
		}

		this.Dispatch2(packet.Message, packet.Flag)
	}
}

func (this *CacheComponent2) Dispatch2(msg *protocol.Message, flag bool) {
	cached := uint16(1 << 15)
	if (msg.Number & cached) != cached {
		return
	}
	msg.Number &= HEAD_RESET

	msg, err := this.MakeCacheMessage(msg)
	if err != nil {
		fmt.Println("CacheCompoennt Dispatch2 Error:" + err.Error())
		return
	}
	data1, err := msg.Encode()
	if err != nil {
		fmt.Println("CacheCompoennt Dispatch2 Error:", err)
		return
	}
	data2 := make([]byte, len(data1)+8)
	copy(data2[:8], msg.Data[0:8])
	copy(data2[8:], data1)

	snode_component.GetNodeComponent().Forward(stree_define.ADD_CACHE_ENTRY, msg.DestUuid, msg.DeviceId, data2)
	//if this message is a contact sync message, send it to root first.
	if (msg.Number == protocol.N_ACCOUNT_SYS || msg.DeviceId == 0) && flag {
		snode_component.GetNodeComponent().Forward(stree_define.BROADCAST_TO_CLIENT, msg.DestUuid, msg.DeviceId, data1)
	} else {
		GetDispatcherOutComponent().NPushBack(msg, flag)
	}
	return
}

func (this *CacheComponent2) MakeCacheMessage(msg *protocol.Message) (*protocol.Message, error) {
	data := make([]byte, 8+msg.DataLen)
	t := time.Now().UnixNano()
	binary.BigEndian.PutUint64(data[:8], uint64(t))
	msg.DataLen += 8
	msg.MsgType = protocol.MT_PUSH_WITH_RESPONSE
	copy(data[8:], msg.Data)
	msg.Data = data
	return msg, nil
}
