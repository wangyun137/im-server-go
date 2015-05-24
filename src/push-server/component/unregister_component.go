package component

import (
	"errors"
	"fmt"
	"push-server/buffpool"
	"push-server/protocol"
	"push-server/utils"
	snode_component "sudo-snode/component"
)

type UnregisterComponent struct {
	msgQueue chan *protocol.Message
	queueLen int
}

var unRegisterComponent UnregisterComponent

func NewUnRegisterComponent(qLen int) (*UnregisterComponent, error) {
	if qLen <= 0 {
		err := errors.New("NewUnregisterComponent() Error:invalid argument")
		return nil, err
	}
	unRegComp := &unRegisterComponent
	unRegComp.queueLen = qLen
	return unRegComp, nil
}

func GetUnregisterComponent() *UnregisterComponent {
	return &unRegisterComponent
}

func (this *UnregisterComponent) Init() error {
	if this.msgQueue != nil {
		err := errors.New("UnregisterComponent Init() Error:UnregisterComponent already initialized")
		return err
	}
	this.msgQueue = make(chan *protocol.Message, this.queueLen)
	return nil
}

func (this *UnregisterComponent) Start() error {
	if this.msgQueue == nil {
		err := errors.New("UnregisterComponent Start() Error:please initialize this object first")
		return err
	}
	go this.Process()
	return nil
}

func (this *UnregisterComponent) Process() error {
	var buffPool *buffpool.BuffPool
	for {
		buffPool = buffpool.GetBuffPool()
		if buffPool != nil {
			break
		} else {
			fmt.Println("UnregisterComponent Process() Error:can not get BuffPool, will try again")
		}
	}
	for {
		msg := <-this.msgQueue
		if msg == nil {
			err := errors.New("UnregisterComponent Process() Error:received msg is nil, will try again")
			continue
		}
		if msg.MsgType != protocol.MT_UNREGISTER {
			err := errors.New("UnregisterComponent Process() Error:received non-MT_UNREGISTER message, will discard it and try again")
			continue
		}

		if msg.PushType == protocol.PT_SERVICE {
			err := this.unregisterService(buffPool, msg)
			if err != nil {
				err := errors.New("UnregisterComponent Process() Error: " + err.Error() + "; will handler next request")
				return err
			}
		} else if msg.PushType == protocol.PT_CLIENT {
			err := this.unregisterClient(buffPool, msg)
			if err != nil {
				err := errors.New("UnregisterComponent Process() Error, " + err.Error() + "; will handler next request")
				return err
			}
		}
	}
	return nil
}

func (this *UnregisterComponent) unregisterService(buffPool *buffpool.BuffPool, msg *protocol.Message) error {
	if buffPool == nil || msg == nil {
		err := errors.New("UnregisterComponent unregisterService() Error: invalid argument")
		return err
	}
	conn := buffPool.QueryService(msg.Number, msg.UserUuid)
	if conn == nil {
		return nil
	}
	//only remove conn from connPool, but do not disconnect this connection
	err := buffPool.DeleteService(msg.Number, msg.UserUuid)
	if err != nil {
		err = errors.New("UnregisterComponent unregisterService() Error: " + err.Error())
		//send failed response to client
		resp, err2 := protocol.NewMessage(protocol.VERSION, protocol.MT_ERROR, protocol.PT_PUSHSERVER, msg.Token, msg.Number, uint16(0),
			P_UID, msg.UserUuid, msg.DeviceId, nil)
		if err2 != nil {
			err2 = errors.New("UnregisterComponent.unregisterService() Error: " + err2.Error())
			return err2
		}
		err3 := sendResponse(conn, resp)
		if err3 != nil {
			err3 = errors.New("UnregisterComponent.unregisterService() Error: " + err3.Error())
			return err3
		}
		return err
	}

	resp, err := protocol.NewMessage(protocol.VERSION, protocol.MT_UNREGISTER, protocol.PT_PUSHSERVER, msg.Token, msg.Number, uint16(0),
		P_UID, msg.UserUuid, msg.DeviceId, nil)
	if err != nil {
		err = errors.New("UnregisterComponent unregisterService() Error:conn is removed from BuffPool, but can not send response to service" + err.Error())
		return err
	}
	err2 := sendResponse(conn, resp)
	if err2 != nil {
		err2 = errors.New("UnregisterComponent unregisterService() Error: conn is reomved from BuffPool, but failed when send response to service" + err2.Error())
		return err2
	}
	return nil
}

func (this *UnregisterComponent) unregisterClient(buffPool *buffpool.BuffPool, msg *protocol.Message) error {
	fmt.Println("in UnregisterComponent.unregisterClient")
	if buffPool == nil || msg == nil {
		err := errors.New("component: UnregisterComponent.unregisterClient() failed, invalid argument")
		fmt.Println(err.Error())
		return err
	}
	clientConn := buffPool.GetClient(msg.Number, msg.UserId, msg.DeviceId, msg.Token)
	if clientConn == nil {
		return nil
	}

	//only remove conn from connPool, but do not disconnect this connection
	err := buffPool.RemoveClient(msg.Number, msg.UserId, msg.DeviceId, msg.Token)
	(*clientConn.Conn).Close()
	if err != nil {
		err = errors.New("component: UnregisterComponent.unregisterClient() failed, " + err.Error())
		fmt.Println(err.Error())
		//send failed response to client
		resp, err2 := protocol.NewMessage(define.VERSION, define.MT_ERROR, define.PT_PUSHSERVER, msg.Token, msg.Number, uint16(0),
			P_UID, msg.UserId, msg.DeviceId, nil)
		if err2 != nil {
			err2 = errors.New("component: UnregisterComponent.unregisterClient() failed, " + err2.Error())
			fmt.Println(err2.Error())
			return err2
		}
		err3 := sendResponse(clientConn.Conn, resp)
		if err3 != nil {
			err3 = errors.New("component: UnregisterComponent.unregisterClient() failed, " + err3.Error())
			fmt.Println(err3.Error())
			return err3
		}
		return err
	}

	//offline broadcast
	err = this.offlineBroadcast(buffPool, msg.UserId, msg.DeviceId)
	if err != nil {
		err = errors.New("component: UnregisterComponent.unregisterClient() error, offlineBroadcast send failed, " + err.Error())
		fmt.Println(err.Error())
	}

	resp, err := protocol.NewMessage(define.VERSION, define.MT_UNREGISTER, define.PT_PUSHSERVER, msg.Token, msg.Number, uint16(0),
		P_UID, msg.UserId, msg.DeviceId, nil)
	if err != nil {
		err = errors.New("UnregisterComponent.unregisterClient() Error: conn is removed from BuffPool, but can not send response to service" + err.Error())
		fmt.Println(err.Error())
		return err
	}
	err2 := sendResponse(clientConn.Conn, resp)
	if err2 != nil {
		err2 = errors.New("UnregisterComponent.unregisterClient() Error:conn is reomved from BuffPool, but failed when send response to service" + err2.Error())
		return err2
	}
	return nil
}

func (this *UnregisterComponent) offlineBroadcast(buffPool *buffpool.BuffPool, userUuid string, devID uint32) error {

	if utils.GetRunWithSnode().Flag {
		snode_component.GetNodeComponent().Logout(userID, devID)
		snode_component.GetNodeComponent().DecUsersNumber()
		snode_component.GetNodeComponent().DecConnNumber()
	}

	if buffPool == nil {
		err := errors.New("UnRegisterComponent offlineBroadcast Error: invalid argument")
		return err
	}

	var dispatcherOut *DispatcherOutComponent
	for {
		dispatcherOut = GetDispatcherOutComponent()
		if dispatcherOut != nil {
			break
		} else {
			fmt.Println("UnregisterComponent.offlineBroadcast() Error:can not get a dispatcher, will try again")
		}
	}

	brMsg := protocol.BRMsg{
		UserUuid: userUuid,
		DeviceId: devID,
	}

	brData, err := brMsg.Encode()
	if err != nil {
		err := errors.New("UnregisterComponent.offlienBroadcast() Error: " + err.Error())
		return err
	}

	servNoUuidArr := buffPool.GetAllServices()
	for num, uuidArr := range servNoUuidArr {
		for _, uuid := range uuidArr {
			offlineMsg, err := protocol.NewMessage(protocol.VERSION, protocol.MT_BROADCAST_OFFLINE, protocol.PT_PUSHSERVER, 0, num, uint16(len(brData)),
				P_UID, uuid, 0, brData)
			if err != nil {
				err = errors.New("UnregisterComponent offlineBroadcast() Error: " + err.Error())
				return err
			}
			dispatcherOut.NPushBack(offlineMsg, false) //not need forward
		}
	}

	return nil
}

func (this *UnregisterComponent) PushBack(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("UnregisterComponent.PushBack() Error: invalid argument")
		return err
	}
	this.msgQueue <- msg
	return nil
}
