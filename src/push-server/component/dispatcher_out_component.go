package component

import (
	"errors"
	"fmt"
	"net"
	"push-server/buffpool"
	"push-server/model"
	"push-server/protocol"
	"push-server/utils"
	snode_component "snode-server/component"
	"time"
	stree_define "tree-server/define"
)

type DispatcherOutComponent struct {
	PriorityQueue
}

func GetDispatcherOutComponent() *DispatcherOutComponent {
	return &dispatcherOutComponent
}

func (this *DispatcherOutComponent) Init() error {
	return nil
}

func (this *DispatcherOutComponent) Start() error {
	go this.process()
	return nil
}

var dispatcherOutComponent DispatcherOutComponent

func NewDispatcherOutComponent(nQueueLen, eQueueLen, level int) (*DispatcherOutComponent, error) {
	pQueue, err := NewPriorityQueue(nQueueLen, eQueueLen, level)
	if err != nil {
		return nil, err
	}
	dispOut := &dispatcherOutComponent
	dispOut.PriorityQueue = *pQueue
	return dispOut, nil
}

func (this *DispatcherOutComponent) Pop() *MessagePacket {
	packet := this.PriorityQueue.Pop()
	GetDispOutTotalQueueCount().Dec()
	return packet
}
func (this *DispatcherOutComponent) EPushBack(msg *protocol.Message, flag bool) error {
	err := this.PriorityQueue.EPushBack(msg, flag)
	GetDispOutEmergentQueueCount().Add()
	GetDispOutTotalQueueCount().Add()
	return err
}
func (this *DispatcherOutComponent) NPushBack(msg *protocol.Message, flag bool) error {
	err := this.PriorityQueue.NPushBack(msg, flag)
	GetDispOutNormalQueueCount().Add()
	GetDispOutTotalQueueCount().Add()
	return err
}

func (this *DispatcherOutComponent) process() {
	var respComp *ResponseComponent
	for {
		respComp = GetResponseComponent()
		if respComp != nil {
			break
		} else {
			fmt.Println("DispatcherOutComponent.Process() Error: can not get ResponComponent sigleton, will try again")
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
		this.dispatch(msg, flag, buffPool, respComp)
	}
}

func (this *DispatcherOutComponent) dispatch(msg *protocol.Message, flag bool, buffPool *buffpool.BuffPool, respComp *ResponseComponent) {
	switch msg.MsgType {
	case protocol.MT_REGISTER: //successful response (to client or service)
		this.handleMTRegister(msg, buffPool, respComp)
	case protocol.MT_PUSH:
		this.handleMTPush(msg, buffPool, respComp)
	case protocol.MT_PUBLISH, protocol.MT_PUSH_WITH_RESPONSE: //only PUBLISH type may forward a message
		if utils.GetRunWithSnode().Flag {
			this.handleMTPublish(msg, flag, buffPool)
		} else {
			cached := uint16(1 << 15)
			if (msg.Number & cached) == cached { //only service can send MT_PUBLISH and Number >> 14 == 1
				this.handleContactSync(msg, flag)
			} else {
				this.handleMTPublish(msg, flag, buffPool)
			}
		}

	case protocol.MT_BROADCAST_ONLINE:
		this.handleMTBroadcast(msg, buffPool)
	case protocol.MT_BROADCAST_OFFLINE:
		this.handleMTBroadcast(msg, buffPool)
	case protocol.MT_ERROR:
		this.handleMTError(msg, buffPool)
	case protocol.MT_KEEP_ALIVE:
		this.handleKeepAlive(msg)
	case protocol.MT_QUERY_USER_RESPONSE:
		if utils.GetRunWithSnode().Flag {
			this.sendToService2(msg, flag, buffPool)
		}
	}
}

func (this *DispatcherOutComponent) handleKeepAlive(msg *protocol.Message) error {
	if msg == nil {
		err := errors.New("DispatcherOutComponent.handleKeepAlive() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}
	buffPool := buffpool.GetBuffPool()
	this.sendToClient(msg, buffPool)
	return nil
}

func (this *DispatcherOutComponent) handleContactSync(msg *protocol.Message, flag bool) error {
	if msg == nil {
		err := errors.New("DispatcherOutComponent.handleContactSync() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}

	GetCacheComponent().Save(*msg)
	GetCacheComponent().SendCached(msg.DestUuid, msg.DeviceId)
	return nil
}

func (this *DispatcherOutComponent) handleMTRegister(msg *protocol.Message, buffPool *buffpool.BuffPool, respComp *ResponseComponent) error {
	if msg.PushType == protocol.PT_PUSHSERVER {
		conn := buffPool.QueryService(msg.Number, msg.DestUuid)
		if conn == nil {
			clientConn := buffPool.GetClient(msg.Number, msg.DestUuid, msg.DeviceId, msg.Token)
			if clientConn == nil || clientConn.Conn == nil {
				err := errors.New("DispatcherOutComponent.handleMTRegister() Error: can not find the dest conn")
				fmt.Println(err.Error())
				return err
			}
			conn = clientConn.Conn
		}
		if conn == nil {
			err := errors.New("DispatcherOutComponent.handleMTRegister() Error: can not find the dest conn")
			fmt.Println(err.Error())
			return err
		}
		err := this.sendMessage(msg, conn)
		if err != nil {
			err = errors.New("DispatcherComponnet.handleMTRegister() Error: can not send message out, " + err.Error())
			fmt.Println(err.Error())
			return err
		}
	}
	return nil
}

func (this *DispatcherOutComponent) handleMTPush(msg *protocol.Message, buffPool *buffpool.BuffPool, respComp *ResponseComponent) error {
	if msg.PushType == protocol.PT_PUSHSERVER {
		conn1, err1 := this.sendToService(msg, buffPool)
		if conn1 != nil && err1 != nil {
			if err1 != nil {
				err := errors.New("Dispatcher.handleMTPush() Error: " + err1.Error())
				fmt.Println(err.Error())
				return err
			}
		} else if conn1 != nil && err1 == nil {
			fmt.Println("\n", time.Now().String(), " in DispatcherOutComponent.handleMTPush, query is send to account system ", "\n")
			return nil
		}

		err2 := this.sendToClient(msg, buffPool)
		if err2 != nil {
			err := errors.New("Dispatcher.handleMTPush() Error: " + err2.Error())
			fmt.Println(err.Error())
			return err
		}
	}
	return nil
}

func (this *DispatcherOutComponent) handleMTPublish(msg *protocol.Message, flag bool, buffPool *buffpool.BuffPool) error {
	if msg == nil {
		fmt.Println("DispacthoutComponent handleMTPublish Error:nil msg")
		return errors.New("DispacthoutComponent handleMTPublish Error:nil msg")
	}
	if msg.PushType == protocol.PT_SERVICE {
		err := this.handlePTService(msg, flag, buffPool)
		if err != nil {
			err = errors.New("Dispatcher.handleMTPublish() Error: " + err.Error())
			fmt.Println(err.Error())
			return err
		}
	} else if msg.PushType == protocol.PT_CLIENT {
		_, err := this.sendToService2(msg, flag, buffPool)
		if err != nil {
			err = errors.New("Dispatcher.handleMTPublish() Error: " + err.Error())
			fmt.Println(err.Error())
			return err
		}
	} else {
		err := errors.New("Dispatcher.handleMTPublish() Error:wrong PushType")
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (this *DispatcherOutComponent) handlePTService(msg *protocol.Message, flag bool, buffPool *buffpool.BuffPool) error {
	err := this.sendToClient2(msg, flag, buffPool)
	if err != nil {
		err = errors.New("Dispatcher.handlePTService() Error: " + err.Error())
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (this *DispatcherOutComponent) handleMTBroadcast(msg *protocol.Message, buffPool *buffpool.BuffPool) error {
	if msg.PushType != protocol.PT_PUSHSERVER {
		err := errors.New("Dispatcher.handleMTBroadcast() Error: wrong PushType")
		fmt.Println(err.Error())
		return err
	}

	conn1, err1 := this.sendToService(msg, buffPool)
	if conn1 != nil && err1 != nil {
		err := errors.New(" Dispatcher.handleMTBroadcast() Error: " + err1.Error())
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (this *DispatcherOutComponent) handleMTError(msg *protocol.Message, buffPool *buffpool.BuffPool) error {
	if msg.PushType != protocol.PT_PUSHSERVER {
		err := errors.New("Dispatcher.handleMTError() Error: wrong PushType")
		fmt.Println(err.Error())
		return err
	}

	conn1, err1 := this.sendToService(msg, buffPool)
	if conn1 != nil && err1 != nil {
		if err1 != nil {
			err := errors.New("Dispatcher.handleMTError() Error: " + err1.Error())
			fmt.Println(err.Error())
			return err
		}
	} else if conn1 != nil && err1 == nil {
		return nil
	}

	err2 := this.sendToClient(msg, buffPool)
	if err2 != nil {
		err := errors.New("Dispatcher.handleMTError() Error: " + err2.Error())
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (this *DispatcherOutComponent) sendToClient2(msg *protocol.Message, forward bool, buffPool *buffpool.BuffPool) error {
	if msg == nil || buffPool == nil {
		err := errors.New("DispatcherOutComponent.sendToClient2() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}

	clientConnArr := make([]*model.Client, 0)
	if msg.Token > 0 {
		clientConn := buffPool.GetClient(msg.Number, msg.DestUuid, msg.DeviceId, msg.Token)
		if clientConn != nil && clientConn.Conn != nil {
			clientConnArr = append(clientConnArr, clientConn)
		}
		if (clientConnArr == nil || len(clientConnArr) <= 0) && forward && GetRunWithSnode().Flag {
			buf, err := msg.Encode()
			if err != nil {
				fmt.Println("DispatcherOutComponent.sendToClient2() Error: " + err.Error())
				fmt.Println("\n", time.Now().String(), "Not find user, not forward due to encode failed.", "\n")
				return err
			}
			snode_component.GetNodeComponent().Forward(stree_define.FORWARD, msg.DestUuid, msg.DeviceId, buf)
			return nil
		}
	} else if msg.Token == 0 && msg.DeviceId > 0 {
		clientConnArr = buffPool.QueryForDeviceId(msg.Number, msg.DestUuid, msg.DeviceId)
		if (clientConnArr == nil || len(clientConnArr) <= 0) && forward && GetRunWithSnode().Flag {
			buf, err := msg.Encode()
			if err != nil {
				fmt.Println("DispatcherOutComponent.sendToClient2() Error:" + err.Error())
				fmt.Println("\n", time.Now().String(), "Not find user, not forward due to encode failed.", "\n")
				return err
			}
			snode_component.GetNodeComponent().Forward(stree_define.FORWARD, msg.DestUuid, msg.DeviceId, buf)
			return nil
		}
	} else if msg.Token == 0 && msg.DeviceId == 0 {
		clientConnArr = buffPool.QueryForUuid(msg.Number, msg.DestUuid)
	}

	err := errors.New("DispatcherOutComponent.sendToClient2() Error: the client's connection not pass authorization yet")
	for _, client := range clientConnArr {
		if client == nil || client.Conn == nil {
			continue
		}
		if !this.checkConnActivity(client.Time, client, buffPool, msg.Number, msg.DestUuid, msg.DeviceId, msg.Token) {
			err := errors.New("DispatcherOutComponent.sendToClient2() Error: client's connection is not alive")
			fmt.Println(err.Error())
			continue
		}
		err2 := this.sendMessage(msg, client.Conn)
		if err2 != nil {
			err2 = errors.New("DispatcherOutComponent.sendToClient2() Error: can not send message to client's connection")
			fmt.Println(err.Error())
			continue
		}
	}
	return nil
}

func (this *DispatcherOutComponent) sendToClient(msg *protocol.Message, buffPool *buffpool.BuffPool) error {
	if msg == nil || buffPool == nil {
		err := errors.New("DispatcherOutComponent.sendToClient() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}
	clientConnArr := make([]*model.Client, 0)

	if msg.Token > 0 {
		clientConn := buffPool.GetClient(msg.Number, msg.DestUuid, msg.DeviceId, msg.Token)
		if clientConn != nil && clientConn.Conn != nil {
			clientConnArr = append(clientConnArr, clientConn)
		}
	} else if msg.Token == 0 && msg.DeviceId > 0 {
		clientConnArr = buffPool.QueryForDeviceId(msg.Number, msg.DestUuid, msg.DeviceId)
	} else if msg.Token == 0 && msg.DeviceId == 0 {
		clientConnArr = buffPool.QueryForUuid(msg.Number, msg.DestUuid)
	}

	if clientConnArr != nil && len(clientConnArr) <= 0 {
		err := errors.New("DispatcherOutComponent.sendToClient() Error: can not find client's connection")
		fmt.Println(err.Error())
		return err
	}

	err := errors.New("DispatcherOutComponent.sendToClient() Error: the client's connection not pass authorization yet")
	for _, client := range clientConnArr {
		if client == nil || client.Conn == nil {
			continue
		}

		if !this.checkConnActivity(client.Time, client, buffPool, msg.Number, msg.DestId, msg.DeviceId, msg.Token) {
			err := errors.New("DispatcherOutComponent.sendToClient() Error: client's connection is not alive")
			fmt.Println(err.Error())
			continue
		}

		err2 := this.sendMessage(msg, client.Conn)
		if err2 != nil {
			err2 = errors.New("DispatcherOutComponent.sendToClient() Error: can not send message to client's connection")
			fmt.Println(err.Error())
			continue
		}
	}

	return err
}

func (this *DispatcherOutComponent) checkConnActivity(Time time.Time, client *model.Client, buffPool *buffpool.BuffPool, number uint16, userUuid string, devID uint32, token uint8) bool {
	//5分钟
	if time.Now().Sub(Time) > define.KEEP_ALIVE_PEORID {
		if client.Conn != nil {
			(*client.Conn).Close()
		}
		if token > 0 {
			err := buffPool.RemoveClient(number, userUuid, devID, token)
			if err != nil {
				err = errors.New("DispatcherComponent.checkConnActivity() Error:can not remove invalid client's connection, " + err.Error())
				fmt.Println(err.Error())
				return false
			}
			clientArr := buffPool.QueryForDeviceId(number, userUuid, devID)
			if clientArr == nil || len(clientArr) == 0 {
				//offline broadcast
				unregister := GetUnregisterComponent()
				if unregister != nil {
					err2 := unregister.offlineBroadcast(buffPool, userUuid, devID)
					if err2 != nil {
						err2 = errors.New("cDispatcherComponent.checkConnActivity() Error: " + err2.Error())
						fmt.Println(err2.Error())
					}
				} else {
					fmt.Println("DispatcherComponent.checkConnActivity() Error: one client disconnected, but offline broadcast send failed, ")
				}
			}

		}
		return false
	}
	return true
}

func (this *DispatcherOutComponent) sendToService2(msg *protocol.Message, forward bool, buffPool *buffpool.BuffPool) (*net.Conn, error) {
	if msg == nil || buffPool == nil {
		err := errors.New("DispatcherOutComponent.sendToService() Error: invalid argument")
		fmt.Println(err.Error())
		return nil, err
	}
	var conn *net.Conn
	if msg.DestUuid == P_UID {
		conn = buffPool.GetService(msg.Number)
	} else {
		conn = buffPool.QueryService(msg.Number, msg.DestUuid)
		if conn == nil && forward && GetRunWithSnode().Flag {
			buf, err := msg.Encode()
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			snode_component.GetNodeComponent().Forward(stree_define.FORWARD, msg.DestUuid, msg.DeviceId, buf)
			return nil, nil
		}
	}

	if conn == nil {
		err := errors.New("DispatcherOutComponent.sendToService() Error:  can not find dest server")
		fmt.Println(err.Error())
		return nil, err
	}

	err := this.sendMessage(msg, conn)
	if err != nil {
		err = errors.New("DispatcherOutComponent.sendToService() Error: can not send message to server's connection")
		fmt.Println(err.Error())
		return conn, err
	}
	if msg.MsgType == protocol.MT_PUBLISH {
		fmt.Println("\n", time.Now().String(), "MT_PUBLISH: sendToService2 Directly OK", "\n")
	}
	return conn, nil
}

func (this *DispatcherOutComponent) sendToService(msg *protocol.Message, buffPool *buffpool.BuffPool) (*net.Conn, error) {
	if msg == nil || buffPool == nil {
		err := errors.New("DispatcherOutComponent.sendToService() Error: invalid argument")
		fmt.Println(err.Error())
		return nil, err
	}
	var conn *net.Conn
	if msg.DestUuid == P_UID {
		conn = buffPool.GetService(msg.Number)
	} else {
		conn = buffPool.QueryService(msg.Number, msg.DestUuid)
		if conn == nil {
			conn = buffPool.GetService(msg.Number)
		}
	}

	if conn == nil {
		err := errors.New("DispatcherOutComponent.sendToService() Error:  can not get a server connection from buffpool")
		fmt.Println(err.Error())
		return nil, err
	}

	err := this.sendMessage(msg, conn)
	if err != nil {
		err = errors.New("DispatcherOutComponent.sendToService() Error:  can not send message to server's connection")
		fmt.Println(err.Error())
		return conn, err
	}
	return conn, nil
}

func (this *DispatcherOutComponent) sendMessage(msg *protocol.Message, conn *net.Conn) error {
	if msg == nil || conn == nil {
		err := errors.New("DispatcherOutComponent sendMessage() Error: invalid argument")
		fmt.Println(err.Error())
		return err
	}

	buf, err := msg.Encode()
	if err != nil {
		err = errors.New("DispatcherOutComponent sendMessage() Error: invalid argument " + err.Error())
		fmt.Println(err.Error())
		return err
	}

	n, err := (*conn).Write(buf)
	if err != nil {
		fmt.Println("\n", time.Now().String(), " Write buf to connection failed, the connection will be closed, n = ", n, err, "\n")
		(*conn).Close()
	}
	return nil
}
