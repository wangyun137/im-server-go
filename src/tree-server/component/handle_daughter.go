package component

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	push_protocol "push-server/protocol"
	"strconv"
	"time"
	"tree-server/define"
	"tree-server/statistics"
	"tree-server/structure"
)

type DaughterHandler struct {
	IP          string
	ServicePort uint16
	queueLen    int
	toDaughter  chan *structure.Packet
}

var daughterHandler DaughterHandler

func NewDaughterHandler(qLen int) (*DaughterHandler, error) {
	if qLen <= 0 {
		err := errors.New("component: DaughterHandler() failed, queue lenght should be positive")
		fmt.Println(err.Error())
		return nil, err
	}
	dh := &daughterHandler
	dh.queueLen = qLen
	return dh, nil
}

//从toDaughter通道中接收Packet
func (this *DaughterHandler) ToDaughterPop() *structure.Packet {
	packet := <-this.toDaughter
	//GetToDaughterQueueCount是QueueCount,将Number-1
	statistics.GetToDaughterQueueCount().Dec()
	return packet
}

//返回一个各项都为空值的DaughterHandler
func GetDaughterHandler() *DaughterHandler {
	return &daughterHandler
}

//开启一个tcp监听，并不断等待有新的连接
//10.0.0.11:11001
func (this *DaughterHandler) ListenService() {
	addr := this.IP + ":" + strconv.Itoa(int(this.ServicePort))
	//监听addr上的tcp连接
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		//不断等待有新的连接
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		go this.handleService(&conn)
	}
}

func (this *DaughterHandler) Init() {
	if this.toDaughter == nil {
		this.toDaughter = make(chan *structure.Packet, this.queueLen)
	}
	return
}

func (this *DaughterHandler) PushToDaughter(packet *structure.Packet) error {
	if packet == nil {
		err := errors.New("component: DaughterHandler.PushToDaughter() failed, got a nil packet as argument")
		fmt.Println(err.Error())
		return err
	}
	//向toDaughter中写入packet
	this.toDaughter <- packet
	statistics.GetToDaughterQueueCount().Add()
	return nil
}

func (this *DaughterHandler) Start() {
	go this.ListenService()
	go this.SendToDaughter()
}

func (this *DaughterHandler) SendToDaughter() {
	for {
		//	fmt.Println("\n", time.Now().String(), " SendToDaughter Directly", "\n")
		//        packet := <-this.toDaughter
		//从toDaughter中接收数据
		packet := this.ToDaughterPop()
		//返回一个*net.Conn
		conn := GetSTree().UserMap.Get(packet.Uuid)
		if conn == nil && packet.Number == define.ALLOC_SERVER {
			fmt.Println("\n", "Geta Min LoadConn ", "\n")
			conn = GetSTree().GetMinLoadConn()
		}
		if conn == nil {
			err := errors.New("component: STree.SendToDaughter() failed, can not find user in user's list")
			fmt.Println(err.Error())
			fmt.Println("\n", "destID = ", packet.Uuid, "\n")
			continue
		}
		buf, err := packet.Encode()
		if err != nil {
			err = errors.New("component: STree.SendToDaughter() failed, can not get buf from packet, " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		//向conn中写入Packet的buf
		_, err = (*conn).Write(buf)
		if err != nil {
			fmt.Println("\n", time.Now().String(), " in SendToDaughter, Got an error ", err, "\n")
			fmt.Println(err.Error())
			continue
		}
		msg := push_protocol.Message{}
		err = msg.Decode(packet.Data)
		if err != nil {
			fmt.Println("MSG Decode Failed, ", err)
			return
		}
		fmt.Println("\n", time.Now().String(), " SendToDaughter Directly OK, packet.Uuid: ", packet.Uuid, " msg.UserId: ", msg.UserId, " msg.DestId: ", msg.DestId, "\n")
	}
}

func (this *DaughterHandler) handleService(conn *net.Conn) {
	stree := GetSTree()
	stree.ServiceMap.Lock.Lock()
	stree.ServiceMap.ServiceMap[conn] = ServiceInfo{}
	stree.ServiceMap.Lock.Unlock()

	this.ReceiveFromDaughter(conn)
}

func (this *DaughterHandler) ReceiveFromDaughter(conn *net.Conn) {
	for {
		//PACKET_HEAD_LEN = 19
		headBuf := make([]byte, define.PACKET_HEAD_LEN)
		//MAX_BUF = 2048
		data := make([]byte, define.MAX_BUF)
		//读取头部信息
		_, err := (*conn).Read(headBuf)
		if err != nil && err != io.EOF {
			fmt.Println(err.Error())
			continue
		} else if err == io.EOF {
			fmt.Println("\n", time.Now().String(), " In DaughterHander.ReceiverFromDaughter, connection ", (*conn).RemoteAddr(), " is invalid", "\n")
			(*conn).Close()
			return
		}
		head := structure.Head{}
		err = head.Decode(headBuf)
		if err != nil {
			err = errors.New("component: STree.Receive failed, can not decode the head buf, " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		if head.DataLen > 0 {
			if int(head.DataLen) > len(data) {
				data = make([]byte, head.DataLen)
			}
			//从conn中读取Data[]部分
			_, err = (*conn).Read(data[:head.DataLen])
			if err != nil && err != io.EOF {
				fmt.Println("Got a error ", err)
				continue
			} else if err == io.EOF {
				fmt.Println("\n", time.Now().String(), " In DaughterHander.ReceiverFromDaughter, connection ", (*conn).RemoteAddr(), " is invalid", "\n")
				(*conn).Close()
				return
			}
		}

		packet := &structure.Packet{head, data[:head.DataLen]}
		if packet.Number == define.FORWARD {
			msg := push_protocol.Message{}
			err = msg.Decode(packet.Data)
			if err != nil {
				fmt.Println("MSG Decode Failed, ", err)
				return
			}
			fmt.Println("\n", time.Now().String(), " ReceiveFromDaughter, packet.Uuid: ", packet.Uuid, " msg.UserId: ", msg.UserId, " msg.DestId: ", msg.DestId, "\n")
		}
		this.Dispatch(packet, conn)
	}
}

func (this *DaughterHandler) Dispatch(packet *structure.Packet, conn *net.Conn) {
	//	fmt.Println("\n", time.Now().String(), " IN DaughterHandler.Dispatch, packet.Number = ", packet.Number, "\n")
	switch packet.Number {
	case define.ONLINE:
		this.handleOnline(packet, conn)
	case define.OFFLINE:
		this.handleOffline(packet, conn)
	case define.FORWARD:
		this.handleForward(packet)
	case define.HEART_BEAT:
		this.handleHeartBeat(packet, conn)
	case define.QUERY_USER:
		this.handleQuery(packet)
	case define.ADD_CACHE_ENTRY:
		this.handleAddCacheEntry(packet)
	case define.DEL_CACHE_ENTRY:
		this.handleDelCacheEntry(packet)
	case define.ALLOC_SERVER_RESPONSE: //as root
		GetClientHandler().PushAllocResponse(packet)
	}
}

func (this *DaughterHandler) handleOnline(packet *structure.Packet, conn *net.Conn) {
	GetSTree().addOnlineInfo(packet, conn)
	if GetSTree().Type == "Root" {
		return
	}
	GetParentHandler().SendToParent(packet)
}

func (this *DaughterHandler) handleOffline(packet *structure.Packet, conn *net.Conn) {
	GetSTree().addOfflineInfo(packet, conn)
	if GetSTree().Type == "Root" {
		return
	}
	GetParentHandler().SendToParent(packet)
}

func (this *DaughterHandler) handleForward(packet *structure.Packet) {
	//	fmt.Println("\n", time.Now().String(), " In DaughterHandler.HandleForWard", "\n")
	statistics.GetForwardCount().AddFromDaughter()
	conn := GetSTree().UserMap.Get(packet.Uuid)
	if conn == nil && GetSTree().Type != "Root" {
		//forward this packet to parent
		//		fmt.Println("Send this Packet to Parent")
		GetParentHandler().SendToParent(packet)
	} else {
		//		fmt.Println("Send this packet to Daughter")
		statistics.GetForwardCount().AddToDaughter()
		this.PushToDaughter(packet)
	}
}

func (this *DaughterHandler) handleHeartBeat(packet *structure.Packet, conn *net.Conn) {
	if len(packet.Data) < 12 {
		err := errors.New("component: handleHeartBeat failed, packet.Data part is not enough")
		fmt.Println(err.Error())
	}
	userNumber := binary.BigEndian.Uint32(packet.Data[0:4])
	maxConn := binary.BigEndian.Uint32(packet.Data[4:8])
	currConn := binary.BigEndian.Uint32(packet.Data[8:12])
	info := ServiceInfo{
		UserNumber:        userNumber,
		MaxConnNumber:     maxConn,
		CurrentConnNumber: currConn,
	}
	stree := GetSTree()
	stree.ServiceMap.Lock.Lock()
	stree.ServiceMap.ServiceMap[conn] = info
	stree.ServiceMap.Lock.Unlock()

	stree.UpdateSelfInfo()
}

func (this *DaughterHandler) handleQuery(packet *structure.Packet) {
	fmt.Println("\n", time.Now().String(), "IN Handle Query", "\n")
	//Query whether a user is online?
	stree := GetSTree()
	conn := stree.UserMap.Get(packet.Uuid)
	if conn == nil && stree.Type != "Root" {
		//forward this Query packet to parent
		GetParentHandler().SendToParent(packet)
		return
	}
	head := structure.Head{
		Number:  define.QUERY_USER_RESPONSE,
		Uuid:    packet.Uuid,
		DataLen: 1,
	}
	resp := &structure.Packet{
		Head: head,
		Data: make([]byte, 1),
	}
	if conn == nil {
		//reply to daughter; packet's content should be defined
		//find user online
		resp.Data[0] = uint8(0)
	} else {
		//not find user
		resp.Data[0] = uint8(1)
	}
	this.PushToDaughter(resp)
}

func (this *DaughterHandler) handleAddCacheEntry(packet *structure.Packet) {
	fmt.Println("\n", time.Now().String(), "IN handleAddCacheEntry", "\n")
	if GetSTree().Type != "Root" {
		GetParentHandler().SendToParent(packet)
	} else {
		GetCacheComponent().SendToCacheDB(packet)
	}
}

func (this *DaughterHandler) handleDelCacheEntry(packet *structure.Packet) {
	fmt.Println("\n", time.Now().String(), "IN handleDelCacheEntry", "\n")
	if GetSTree().Type != "Root" {
		GetParentHandler().SendToParent(packet)
	} else {
		GetCacheComponent().SendToCacheDB(packet)
	}
}
