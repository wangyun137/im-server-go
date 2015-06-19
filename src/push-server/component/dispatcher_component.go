package component

import (
	connect_protocol "connect-server/protocol"
	"encoding/binary"
	"errors"
	"fmt"
	"push-server/buffpool"
	"push-server/protocol"
	snode_component "snode-server/component"
	stree_define "tree-server/define"
)

type DispatcherComponent struct {
	PriorityQueue
}

var dispatcher DispatcherComponent

//111111111111111
const MSGNUM_BASE = 0x3fff

func NewDispatcherComponent(nQueueLen, eQueueLen, level int) (*DispatcherComponent, error) {
	pQueue, err := NewPriorityQueue(nQueueLen, eQueueLen, level)
	if err != nil {
		return nil, err
	}
	dispatcher.PriorityQueue = *pQueue
	return &dispatcher, nil
}

func (this *DispatcherComponent) Init() error {
	return nil
}

func (this *DispatcherComponent) Start() error {
	go this.process()
	return nil
}

func (this *DispatcherComponent) process() {
	var responseComponent *ResponseComponent
	for {
		responseComponent = GetResponseComponent()
		if responseComponent != nil {
			break
		} else {
			fmt.Println("DispatcherComponent.Process() Error: can not get ResponComponent sigleton, will try again")
		}
	}

	var buffPool *buffpool.BuffPool
	for {
		buffPool = buffpool.GetBuffPool()
		if buffPool != nil {
			break
		}
	}

	for {
		packet := this.Pop()
		msg := packet.Message
		flag := packet.Flag
		this.dispatch(msg, flag, buffPool, responseComponent)
	}
}

func (this *DispatcherComponent) dispatch(msg *protocol.Message, flag bool, buffPool *buffpool.BuffPool, respComp *ResponseComponent) {
	switch msg.MsgType {
	case protocol.MT_UNREGISTER:
		this.handleMTUnregister(msg)
	case protocol.MT_PUSH:
		this.handleMTPush(msg, buffPool, respComp)
	case protocol.MT_KEEP_ALIVE:
		this.handleKeepAlive(msg)

	//MT_DEL_CACHE_ENTRY from other servers, such as gconnect, MT_PUSH_WITH_RESPONSE from clients for contact sync
	case protocol.MT_DEL_CACHE_ENTRY, protocol.MT_PUSH_WITH_RESPONSE:
		if GetRunWithSnode().Flag {
			//since the response is sent by client, the packet's uuid should be the push-message's UserId
			snode_component.GetNodeComponent().Forward(stree_define.DEL_CACHE_ENTRY, msg.UserUuid, msg.DeviceId, msg.Data)
		} else {
			GetCacheComponent().RemoveCached(msg)
		}
	case protocol.MT_PUBLISH: //only PUBLISH type may forward a message
		cached := uint16(1 << 15)
		if (msg.Number & cached) == cached { //only service can send MT_PUBLISH and Number >> 14 == 1
			if GetRunWithSnode().Flag {
				GetCacheComponent2().MsgPush2(msg, flag)
			} else {
				this.handleContactSync(msg, flag)
			}
		} else {
			this.handleMTPublish(msg, flag, buffPool)
		}

	case protocol.MT_QUERY_USER:
		if GetRunWithSnode().Flag {
			this.handleQueryUser(msg)
		}
	}
}

func (this *DispatcherComponent) handleQueryUser(msg *protocol.Message) {
	if msg == nil {
		return
	}
	data := make([]byte, stree_define.QUERY_USER_REQUEST_LEN)
	copy(data, msg.Data[:4])
	serverID, err := connect_protocol.StringToUuid(msg.UserUuid)
	if err != nil {
		err = errors.New("DispatcherComponent.handleQueryUser() Error: " + err.Error())
		fmt.Println(err.Error())
		return
	}
	copy(data[4:20], serverID)
	binary.BigEndian.PutUint16(data[20:], msg.Number)
	snode_component.GetNodeComponent().Forward(stree_define.QUERY_USER, msg.DestUuid, msg.DeviceId, data)
	return
}

func (this *DispatcherComponent) handleKeepAlive(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("DispatcherComponent.handleKeepAlive() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}
	buffPool := buffpool.GetBuffPool()
	err := buffPool.UpdateClientStatus(msg.Number, msg.UserUuid, msg.DeviceId, msg.Token, true)
	if err != nil {
		err = errors.New("DispatcherComponent.handleKeepAlive() Error: can not update client's connections status" + err.Error())
		fmt.Println(err.Error())
		return err
	}
	msg.PushType = protocol.PT_PUSHSERVER
	msg.UserUuid, msg.DestUuid = msg.DestUuid, msg.UserUuid
	GetDispatcherOutComponent().EPushBack(msg, false)
	return nil
}
func (this *DispatcherComponent) handleContactSync(msg *protocol.Message, flag bool) error {
	if msg == nil {
		err := errors.New("DispatcherComponent.handleContactSync() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}

	GetDispatcherOutComponent().NPushBack(msg, flag)
	return nil
}

func (this *DispatcherComponent) handleMTUnregister(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("DispatcherComponent.handleMTUnregister() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}
	unRegister := GetUnregisterComponent()
	if unRegister == nil {
		err := errors.New("DispatcherComponent.handleMTUnregister() Error: can not get unregisterComponent")
		fmt.Println(err.Error())
		return err
	}
	err := unRegister.PushBack(msg)
	if err != nil {
		err := errors.New("DispatcherComponent.handleMTUnregister() Error: can not push the request to UnregisterComponent.msgQueue")
		fmt.Println(err.Error())
		return err
	}
	return nil
}
func (this *DispatcherComponent) handleMTPush(msg *protocol.Message, buffPool *buffpool.BuffPool, respComp *ResponseComponent) error {
	if msg.PushType == protocol.PT_PUSHSERVER {
		//this case will be directly sent to DispOut
		//		GetDispatcherOutComponent().NPushBack(msg, false)
	} else if msg.PushType == protocol.PT_CLIENT || msg.PushType == protocol.PT_SERVICE {
		respComp.AcceptResponseInfo(msg)
		return nil
	}
	return nil
}

func (this *DispatcherComponent) handleMTPublish(msg *protocol.Message, flag bool, buffPool *buffpool.BuffPool) error {
	if msg.DeviceId == 0 && flag && GetRunWithSnode().Flag {
		buf, err := msg.Encode()
		if err != nil {
			err = errors.New("DispatcherComponnet.handleMTPublish() Error: " + err.Error())
			fmt.Println(err.Error())
			return err
		}
		snode_component.GetNodeComponent().Forward(stree_define.BROADCAST_TO_CLIENT, msg.DestUuid, msg.DeviceId, buf)
		return nil
	} else {
		GetDispatcherOutComponent().NPushBack(msg, flag)
	}
	return nil
}
