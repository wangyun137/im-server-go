package component

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	"tree-server/define"
	"tree-server/statistics"
	"tree-server/structure"
)

type ParentHandler struct {
	MasterIP   string //10.0.0.11
	SlaveIP    string //10.0.0.11
	ParentPort uint16 //11002

	masterConn *net.Conn
	slaveConn  *net.Conn
	queueLen   int //1000
	sendQueue  chan *structure.Packet

	PushServerPort uint16 //9005
}

var parentHandler ParentHandler

//构造一个queueLen字段为1000,其余字段都为空的ParentHandler
func NewParentHandler(qLen int) (*ParentHandler, error) {
	if qLen <= 0 {
		err := errors.New("component: NewPareneHandler failed, queue length should be positive")
		fmt.Println(err.Error())
		return nil, err
	}
	handler := &parentHandler
	handler.queueLen = qLen
	return handler, nil
}

//创建一个长度为1000，元素为Packet的缓冲通道
func (this *ParentHandler) Init() {
	if this.sendQueue == nil {
		this.sendQueue = make(chan *structure.Packet, this.queueLen)
	}
}

//GetToParentQueueCount()返回一个都为空的QueueCount
//从sendQueue中接收数据
func (this *ParentHandler) SendQueuePop() *structure.Packet {
	//接收通道中的Packet
	packet := <-this.sendQueue
	//QueueCount数减1
	statistics.GetToParentQueueCount().Dec()
	return packet
}

func GetParentHandler() *ParentHandler {
	return &parentHandler
}

func (this *ParentHandler) Start() {
	//创建一个在10.0.0.11:11002上的tcp连接
	this.ConnParent()
	go this.ReceiveFromParent()
	go this.SendOut()
}

func (this *ParentHandler) ReceiveFromParent() {
	for {
		//MAX_BUF = 2048
		data := make([]byte, define.MAX_BUF)
		//PACKET_HEAD_LEN = 19
		headBuf := make([]byte, define.PACKET_HEAD_LEN)

		if this.masterConn == nil {
			time.Sleep(30 * time.Second)
			continue
		}
		//从10.0.0.11:11002上的tcp连接中读取head信息并填充到headbuf
		_, err := (*this.masterConn).Read(headBuf)
		if err != nil && err != io.EOF {
			continue
		} else if err == io.EOF {
			return
		}
		head := structure.Head{}
		err = head.Decode(headBuf)
		if err != nil {
			err = errors.New("component: ReceiveFromParent failed, can not decode the head buf, " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		if head.DataLen > 0 {
			if int(head.DataLen) > len(data) {
				data = make([]byte, head.DataLen)
			}
			//从10.0.0.11:11002上的tcp连接中读取Data部分的数据
			_, err = (*this.masterConn).Read(data[:head.DataLen])
			if err != nil && err != io.EOF {
				continue
			} else if err == io.EOF {
				return
			}
		}

		packet := structure.Packet{head, data[:head.DataLen]}

		this.dispatch(&packet)
	}
}

func (this *ParentHandler) dispatch(packet *structure.Packet) {
	switch packet.Number {
	case define.FORWARD:
		//GetForwardCount()返回的是ForwardCount,然后将ToDaughter加1
		statistics.GetForwardCount().AddToDaughter()
		GetDaughterHandler().PushToDaughter(packet)
	case define.QUERY_USER_RESPONSE:
		GetDaughterHandler().PushToDaughter(packet)
	case define.ALLOC_SERVER:
		this.handleAllocServer(packet)
	}
}

func (this *ParentHandler) handleAllocServer(packet *structure.Packet) {
	//ALLOC_SERVER_RESPONSE = 0x0a
	packet.Number = define.ALLOC_SERVER_RESPONSE
	packet.Data = make([]byte, 6)
	//PushServerPort:9005
	binary.BigEndian.PutUint16(packet.Data[4:6], this.PushServerPort)
	//返回一个个字段都为空值的STree
	stree := GetSTree()
	if pConn, ok := stree.UserMap.UserMap[packet.Uuid]; ok {
		ip := net.ParseIP(strings.Split((*pConn).RemoteAddr().String(), ":")[0]).To4()
		copy(packet.Data[0:4], ip)
	} else {
		conn := stree.GetMinLoadConn()
		ip := net.ParseIP(strings.Split((*conn).RemoteAddr().String(), ":")[0]).To4()
		copy(packet.Data[0:4], ip)
	}
	this.SendToParent(packet)
}

func (this *ParentHandler) SendToParent(packet *structure.Packet) error {
	if packet == nil {
		err := errors.New("component: ParentHandler.SendToParent() failed, got a nil pointer")
		fmt.Println(err.Error())
	}
	//向sendQueue通道中发送packet
	this.sendQueue <- packet
	//QueueCount的Number加1
	statistics.GetToParentQueueCount().Add()
	return nil
}

func (this *ParentHandler) SendOut() error {
	for {
		//		packet := <-this.sendQueue
		//从sendQueue中接收数据
		packet := this.SendQueuePop()

		if packet == nil {
			err := errors.New("component: ParentHandler.SendOut() failed, got a nil pointer form ParentHandler.sendQueue")
			fmt.Println(err.Error())
			continue
		}
		buf, err := packet.Encode()
		if err != nil {
			err = errors.New("component: ParentHandler.SendOut failed, can not got the buf from packet")
			fmt.Println(err.Error())
			continue
		}
		if this.masterConn == nil {
			fmt.Println("\n", time.Now().String(), " STree's parent(Root)'s conn is invalid", "\n")
			time.Sleep(30 * time.Second)
			continue
		}
		//向10.0.0.11:11002中写入Packet的buf
		_, err = (*this.masterConn).Write(buf)
		if err != nil {
			err = errors.New("component: ParentHandler.SendOut failed, can not send packet to parent's master connection, will switch to slave connection")
			fmt.Println(err.Error())
			continue
		}
	}
}

func (this *ParentHandler) ConnParent() error {
	//10.0.0.11:11002
	masterAddr := this.MasterIP + ":" + strconv.Itoa(int(this.ParentPort))
	//	slaveAddr := this.SlaveIP + ":" + strconv.Itoa(int(this.ParentPort))
	//创建一个在10.0.0.11:11002上的tcp连接
	conn1, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.masterConn = &conn1

	//	conn2, err := net.Dial("tcp", slaveAddr)
	//	if err != nil {
	//		log.Error(err.Error())
	//		return err
	//	}
	//	this.slaveConn = &conn2
	return nil
}
