package api

import (
	"errors"
	"fmt"
	"io"
	"net"
	"push-server/protocol"
	"sync"
	"time"
)

type PushObj struct {
	MsgType  uint8
	Token    uint8
	Number   uint16
	DestUuid string
	DeviceId uint32
	PushInfo
}

type PushInfo interface {
	Encode() ([]byte, error)
}

type Callback interface {
	Receive(*protocol.Message) error
}

type CallbackPush interface {
	Receive(*protocol.Message) error
}

type Pusher struct {
	Number   uint16
	UserUuid string
	Connector
	conn           *net.TCPConn
	pushAddr       string
	pushQueue      chan *PushObj
	queueLen       int
	receiveRWMutex sync.RWMutex
	receiveMap     map[uint8][]Callback
	pushRWMutex    sync.RWMutex
	pushMap        map[uint16][]CallbackPush
	blockingChan   chan *protocol.Message
	ctrlChan       chan *protocol.Message
	stopped        bool
	stopRWMutex    sync.RWMutex
}

var pusherInstance Pusher

func NewPushObj(msgType, token uint8, number uint16, destUuid string, deviceId uint32, info PushInfo) (*PushObj, error) {
	if len(destUuid) != USER_UUID_LEN {
		err := errors.New("Pusher NewPushObj Error: invalid argument - the destUuid's length is not equal 32")
		return nil, err
	}
	pushObj := &PushObj{
		MsgType:  msgType,
		Token:    token,
		Number:   number,
		DestUuid: destUuid,
		DeviceId: deviceId,
		PushInfo: info,
	}
	return pushObj, nil
}

func NewPusher(number uint16, userUuid string, pushQueueLen int) (*Pusher, error) {
	if pushQueueLen <= 0 || len(userUuid) != USER_UUID_LEN {
		err := errors.New("Pusher NewPusher Error: invalid argument - the pushQueueLen is lower than 0 or userUuid is not equal 32")
		return nil, err
	}

	pusher := &pusherInstance
	pusher.Number = number
	pusher.UserUuid = userUuid
	pusher.queueLen = pushQueueLen
	pusher.stopped = true

	pusher.initMembers()
	return pusher, nil
}

func (pusher *Pusher) initMembers() {
	pusher.pushQueue = make(chan *PushObj, pusher.queueLen)
	pusher.blockingChan = make(chan *protocol.Message, 1)
	pusher.receiveMap = make(map[uint8][]Callback)
	pusher.ctrlChan = make(chan *protocol.Message, 1)
	pusher.pushMap = make(map[uint16][]CallbackPush)
}

func GetPusher() *pusher {
	return &pusherInstance
}

func (pusher *Pusher) RegisterPushReceiver(number uint16, receiver CallbackPush) error {
	if pusher.hasPushReceiver(number, receiver) {
		err := errors.New("Pusher RegisterPushReceiver Error: receiver has already registered with same number")
		return err
	}
	err := pusher.addPushReceiver(number, receiver)
	if err != nil {
		return err
	}
	return nil
}

func (pusher *Pusher) hasPushReceiver(number uint16, receiver CallbackPush) bool {
	pusher.pushRWMutex.RLock()
	defer pusher.pushRWMutex.RUnlock()
	if receiverArr, ok := pusher.pushMap[number]; ok {
		for _, v := range receiverArr {
			if v == receiver {
				return true
			}
		}
	}
}

func (pusher *Pusher) addPushReceiver(number uint16, receiver CallbackPush) error {
	pusher.pushRWMutex.Lock()
	defer pusher.pushRWMutex.Unlock()
	if pusher.pushMap == nil {
		err := errors.New("Pusher addPushReceiver Error:the Pusher has not initialized")
		return err
	}
	pusher.pushMap[number] = append(pusher.pushMap[number], receiver)
	return nil
}

func (pusher *Pusher) RegisterReceiver(msgType uint8, receiver Callback) error {
	if pusher.hasReceiver(msgType, receiver) {
		err := errors.New("Pusher RegisterReceiver Error: receiver has already registered with same msgType")
		return err
	}
	err := pusher.addReceiver(msgType, receiver)
	if err != nil {
		return err
	}
	return nil
}

func (pusher *Pusher) hasReceiver(msgType uint8, receiver Callback) bool {
	pusher.receiveRWMutex.RLock()
	defer pusher.receiveRWMutex.RUnlock()
	if receiverArr, ok := pusher.receiveMap[msgType]; ok {
		for _, value := range receiverArr {
			if value == receiver {
				return true
			}
		}
	}
	return false
}

func (pusher *Pusher) addReceiver(msgType uint8, receiver Callback) error {
	pusher.receiveRWMutex.Lock()
	defer pusher.receiveRWMutex.Unlock()
	if pusher.receiveMap == nil {
		err := errors.New("Pusher addReceiver Error:the Pusher has not initialized")
		return err
	}
	pusher.receiveMap[msgType] = append(pusher.receiveMap[msgType], receiver)
}

func (pusher *Pusher) Start(pushAddr string) error {
	err := pusher.init(pushAddr)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	pusher.stopRWMutex.Lock()
	defer pusher.stopRWMutex.Unlock()
	pusher.stopped = false

	go pusher.receive()
	go pusher.publish()
	return nil
}

func (pusher *Pusher) init(addr string) error {
	pusher.stopRWMutex.RLock()
	defer pusher.stopRWMutex.RUnlock()
	if !pusher.stopped {
		err := errors.New("Pusher init Error: Pusher has already started")
		return err
	}
	conn, err := pusher.connect(addr)
	pusher.conn = conn
	if err != nil {
		return err
	}
	pusher.pushAddr = addr

	err = pusher.register()
	if err != nil {
		return err
	}
	return nil

}

func (puhser *Pusher) connect(addr string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		err = errors.New("Pusher connect Error: " + err.Error())
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		err = errors.New("Pusher connect Error: " + err.Error())
		return nil, err
	}
	return conn, nil
}

func (pusher *Pusher) register() error {
	msg, err := protocol.NewMessage(uint8(protocol.VERSION), uint8(protocol.MT_REGISTER), uint8(protocol.PT_SERVICE), uint8(0), pusher.Number, uint16(P_SN_LEN),
		pusher.UserUuid, P_DEST_UUID, uint32(0), []byte(P_SN))
	if err != nil {
		err := errors.New("Pusher register Error: " + err.Error())
		return err
	}

	_, _, err = pusher.Connector.Register(msg, pusher.conn)
	if err != nil {
		err = errors.New("Pusher register Error: " + err.Error())
		return err
	}
	return nil
}

func (pusher *Pusher) receive() {
	for {
		pusher.stopRWMutex.RLock()
		if pusher.stopped {
			pusher.stopRWMutex.RUnlock()
			return
		} else {
			pusher.stopRWMutex.RUnlock()
		}
		msg, err := pusher.Connector.ReadMessage(pusher.conn)
		if err != nil && err != io.EOF && pusher.Conn == nil {
			err = errors.New("Pusher receive Error: " + err.Error())
			continue
		} else if err != nil && (err == io.EOF || pusher.conn == nil) {
			err = errors.New("Pusher receive Error: the conn is nil")
			fmt.Println(err.Error())

			pusher.stopRWMutex.RLock()
			if !pusher.stopped {
				pusher.Reconnect()
				pusher.stopRWMutex.RUnlock()
				continue
			} else {
				pusher.stopRWMutex.RUnlock()
				return
			}
		}
		pusher.dispatch(msg)
	}
}

func (pusher *Pusher) Reconnect() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", pusher.pushAddr)
	if err != nil {
		err = errors.New("Pusher Reconnect Error: " + err.Error())
		fmt.Println(err.Error())
		return
	}
	for {
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			err = errors.New("Pusher Reconnect Error: " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		pusher.conn = conn
		err = pusher.register()
		if err != nil {
			err = errors.New("Pusher Reconnect Error: " + err.Error())
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}
}

func (pusher *Pusher) dispatch(msg *protocol.Message) {
	switch msg.MsgType {
	case protocol.MT_PUBLISH:
		pusher.handleMTPublish(msg)
	case protocol.MT_UNREGISTER:
		pusher.ctrlChan <- msg
	case protocol.MT_PUSH:
		pusher.handleMessage("handle MT_PUSH ", msg)
	case protocol.MT_ERROR:
		pusher.handleMessage("handle MT_ERROR", msg)
	case protocol.MT_BROADCAST_ONLINE:
		pusher.handleMessage("handle MT_BROADCAST_ONLINE", msg)
	case protocol.MT_BROADCAST_OFFLINE:
		pusher.handleMessage("handle MT_BROADCAST_OFFLINE", msg)
	case protocol.MT_QUERY_USER_RESPONSE:
		pusher.handleMessage("Handle MT_QUERY_USER_RESPONSE", msg)
	default:

	}
}

func (pusher *Pusher) handleMTPublish(msg *protocol.Message) {
	if (msg.Number >> 15) == 0x0 {
		receiveArr := pusher.getReceiver(msg.MsgType)
		for _, receiver := range receiveArr {
			receiver.Receive(msg)
		}
	} else {
		pusher.blockingChan <- msg
	}
}

func (pusher *Pusher) handleMessage(tag string, msg *protocol.Mesasge) {
	if msg.PushType != protocol.PT_PUSHSERVER {
		fmt.Println(tag + " Error: the message's pushType is not PushServer")
		return
	}
	receiveAddr := pusher.getReceiver(msg.MsgType)
	for _, receiver := range receiveAddr {
		receiver.Receive(msg)
	}
}

func (pusher *Pusher) getReceiver(msgType uint8) []Callback {
	pusher.receiveRWMutex.RLock()
	defer pusher.receiveRWMutex.Unlock()
	return pusher.receiveMap[msgType]
}

func (pusher *Pusher) publish() {
	for {
		pusher.stopRWMutex.RLock()
		if pusher.stopped {
			pusher.stopRWMutex.RUnlock()
			return
		} else {
			pusher.stopRWMutex.RUnlock()
		}

		msg, err := pusher.createMsg()
		if err == nil && msg == nil {
			return
		} else if err != nil {
			err = errors.New("Pusher publish Error: " + err.Error())
			continue
		}

		err = pusher.Connector.Publish(msg, pusher.conn)
		if err != nil && err != io.EOF {
			err = errors.New("Pusher publish Error: " + err.Error())
			continue
		} else if err != nil && (err == io.EOF || pusher.conn == nil) {
			err = errors.New("Pusher publish Error: the conn is invalid")
			return
		}
	}
}

func (pusher *Pusher) createMsg() (msg *protocol.Message, err error) {
	pushObj := <-pusher.pushQueue
	if pushObj == nil {
		return nil, nil
	}
	var buffer []byte
	if pushObj.PushInfo == nil {
		buffer = nil
	} else {
		buffer, err = pushObj.PushInfo.Encode()
		if err != nil {
			err = errors.New("Pusher createMsg Error: " + err.Error())
			return nil, err
		}
	}
	msg, err = protocol.NewMessage(uint8(protocol.VERSION), pushObj.MsgType, uint8(protocol.PT_SERVICE), pushObj.Token,
		pushObj.Number, uint16(len(buffer)), pusher.UserUuid, pushObj.DestUuid, pushObj.DeviceId, buffer)
	if err != nil {
		err = errors.New("Pusher createMsg Error: " + err.Error())
		return nil, err
	}
	return msg, nil
}

func (pusher *Pusher) Push(pushObj *PushObj) error {
	if pushObj == nil {
		err := errors.New("Pusher Push Error: invalid argument - the pushObj is nil")
		return nil, err
	}
	pusher.pushQueue <- pushObj
	return nil
}

func (pusher *Pusher) PushBlocking(pushObj *PushObj) (*protocol.Message, error) {
	if pushObj == nil {
		err := errors.New("Pusher PushBlocking Error: invalid argument - the pushObj is nil")
		return nil, err
	}
	pusher.pushQueue <- pushObj
	msg := <-pusher.blockingChan
}

func (pusher *Pusher) PushCtrlBlocking(pushObj *PushObj) (*protocol.Message, error) {
	if pushObj == nil {
		err := errors.New("Pusher PushCtrlBlocking Error: invalid argument - thee pushObj is nil")
		return nil, err
	}
	pusher.pushQueue <- pushObj
	msg := <-pusher.ctrlChan
	return msg, nil
}

func (pusher *Pusher) Stop() error {
	err
}

func (pusher *Pusher) Destroy() error {
	if pusher.conn == nil {
		err := errors.New("Pusher Destroy Error: invalid argument - the conn is nil")
		return err
	}

	pushObj, err := NewPushObj(uint8(protocol.MT_UNREGISTER), uint(0), pusher.Number, P_DEST_UUID, uint32(0), nil)
	if err != nil {
		err = errors.New("Pusher Destroy Error: " + err.Error())
		return err
	}

	response, err := pusher.PushCtrlBlocking(pushObj)
	if err != nil {
		err = errors.New("Pusher Destroy Error: " + err.Error())
		return err
	}
	if response.MsgType == protocol.MT_ERROR && response.PushType == protocol.PT_PUSHSERVER && response.DestUuid == pusher.UserUuid {
		err = errors.New("Pusher Destroy Error: Push-Server refused the unregister request")
		return err
	}
	if response.MsgType != protocol.MT_UNREGISTER || response.PushType != protocol.PT_PUSHSERVER || response.DestUuid != pusher.UserUuid ||
		response.Number != pusher.Number {
		err = errors.New("Pusher Destroy Error: receive wrong unregister response")
		return err
	}
	pusher.conn.Close()
	return nil
}
