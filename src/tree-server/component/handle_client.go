package component

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"tree-server/define"
	"tree-server/statistics"
	"tree-server/structure"
)

type ClientHandler struct {
	IP                 string
	ClientPort         uint16
	clientQueueLen     int
	clientQueue        chan *net.Conn
	allocRespQueueLen  int
	allocResponseQueue chan *structure.Packet
}

var clientHandler ClientHandler

//cLen,aLen都为1000
func NewClientHandler(cLen, aLen int) (*ClientHandler, error) {
	if cLen <= 0 || aLen <= 0 {
		err := errors.New("component: NewClientHandler() failed, queue length should be positvie")
		fmt.Println(err.Error())
		return nil, err
	}
	client := &clientHandler
	client.clientQueueLen = cLen
	client.allocRespQueueLen = aLen
	return client, nil
}

func (this *ClientHandler) Init() {
	if this.clientQueue == nil {
		this.clientQueue = make(chan *net.Conn, this.clientQueueLen)
	}
	if this.allocResponseQueue == nil {
		this.allocResponseQueue = make(chan *structure.Packet, this.allocRespQueueLen)
	}
	return
}

func GetClientHandler() *ClientHandler {
	return &clientHandler
}

//向allocResponseQueue写入Packet数据
func (this *ClientHandler) PushAllocResponse(packet *structure.Packet) error {
	if packet == nil {
		err := errors.New("component: Got a nil pointer as the argument")
		fmt.Println(err.Error())
		return err
	}
	this.allocResponseQueue <- packet
	statistics.GetAllocRespQueueCount().Add()
	return nil
}

//从allocResponseQueue中读取Packet数据
func (this *ClientHandler) AllocResponsePop() *structure.Packet {
	packet := <-this.allocResponseQueue
	statistics.GetAllocRespQueueCount().Dec()
	return packet
}

//向clientQueue中写入*net.Conn数据
func (this *ClientHandler) ClientQueuePush(conn *net.Conn) {
	this.clientQueue <- conn
	statistics.GetClientQueueCount().Add()
}

//从clientQueue中读取*net.Conn数据
func (this *ClientHandler) ClientQueuePop() *net.Conn {
	conn := <-this.clientQueue
	statistics.GetClientQueueCount().Dec()
	return conn
}

func (this *ClientHandler) Start() {
	go this.HandleClient()
	go this.ListenClient()
}

//监听10.0.0.11:11000上的tcp连接
func (this *ClientHandler) ListenClient() {
	addr := this.IP + ":" + strconv.Itoa(int(this.ClientPort))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		//		this.clientQueue <-&conn
		this.ClientQueuePush(&conn)
	}
}

func (this *ClientHandler) HandleClient() {
	for {
		//        conn := <-this.clientQueue
		conn := this.ClientQueuePop()
		if conn == nil {
			continue
		}
		req, err := this.sendAllocReq(conn)
		if err != nil {
			continue
		}

		//        resp := <-this.allocResponseQueue
		resp := this.AllocResponsePop()
		if resp.Uuid != req.Uuid {
			err := errors.New("component: HandleClient() failed, got a wrong response")
			fmt.Println(err.Error())
			(*conn).Close()
			continue
		}
		resp.Number = define.LOGIN_RESPONSE
		buf2, err := resp.Encode()
		if err != nil {
			err := errors.New("component: HandleClient() failed, make the buf of response")
			fmt.Println(err.Error())
			(*conn).Close()
			continue
		}
		_, err = (*conn).Write(buf2)
		if err != nil {
			err := errors.New("component: HandleClient() failed, can not send response to client")
			fmt.Println(err.Error())
			(*conn).Close()
			continue
		}
	}
}

func (this *ClientHandler) sendAllocReq(conn *net.Conn) (*structure.Packet, error) {
	buf := make([]byte, 19)
	_, err := (*conn).Read(buf)
	if err != nil {
		err = errors.New("component: sendAllocReq() failed, can not read a log in message from conn")
		fmt.Println(err.Error())
		(*conn).Close()
		return nil, err
	}
	req := structure.Packet{}
	err = req.Decode(buf)
	if err != nil {
		err = errors.New("component: sendAllReq() failed, can not get message from received buf")
		fmt.Println(err.Error())
		(*conn).Close()
		return nil, err
	}
	req.Number = define.ALLOC_SERVER
	req.DataLen = 0
	req.Data = nil
	GetDaughterHandler().PushToDaughter(&req)
	return &req, nil
}
