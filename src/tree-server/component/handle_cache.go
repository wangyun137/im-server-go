package component

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"tree-server/define"
	"tree-server/statistics"
	"tree-server/structure"
)

type CacheComponent struct {
	CacheIP        string
	CachePort      uint16
	queueLen       int
	CacheQueue     chan *structure.Packet
	respQueueLen   int
	CacheRespQueue chan *structure.Packet
	CacheConn      *net.Conn
}

var cacheComponent CacheComponent

func NewCacheComponent(qLen, respQLen int) (*CacheComponent, error) {
	if qLen <= 0 || respQLen <= 0 {
		err := errors.New("component: NewCacheComponent() failed, queue length should be positive")
		return nil, err
	}
	cache := &cacheComponent
	cache.queueLen = qLen
	cache.respQueueLen = respQLen
	return cache, nil
}

//GetCacheQueueCount()返回QueueCount，将Number+1
func (this *CacheComponent) CachePush(packet *structure.Packet) {
	this.CacheQueue <- packet
	statistics.GetCacheQueueCount().Add()
}

//从CacheQueue中读取数据，并将Number-1
func (this *CacheComponent) CachePop() *structure.Packet {
	packet := <-this.CacheQueue
	statistics.GetCacheQueueCount().Dec()
	return packet
}

//GetCacheRespQueueCount()返回QueueCount，将Number+1
func (this *CacheComponent) CacheRespPush(packet *structure.Packet) {
	this.CacheRespQueue <- packet
	statistics.GetCacheRespQueueCount().Add()
}

//从CacheRespQueue重读取数据，并将Number - 1
func (this *CacheComponent) CacheRespPop() *structure.Packet {
	packet := <-this.CacheRespQueue
	statistics.GetCacheRespQueueCount().Dec()
	return packet
}

func GetCacheComponent() *CacheComponent {
	return &cacheComponent
}

func (this *CacheComponent) Init() {
	if this.CacheQueue == nil {
		this.CacheQueue = make(chan *structure.Packet, this.queueLen)
	}
	if this.CacheRespQueue == nil {
		this.CacheRespQueue = make(chan *structure.Packet, this.respQueueLen)
	}
}

func (this *CacheComponent) Start() {
	//建立10.0.0.11：11003上的tcp连接
	err := this.ConnCache()
	if err != nil {
		err = errors.New("component: CacheComponent.Start() failed, " + err.Error())
		fmt.Println(err.Error())
	}
	go this.ReceiveFromCacheDB()
	go this.SendOut()
}

func (this *CacheComponent) ReceiveFromCacheDB() {
	for {
		//MAX_BUF= 2048
		data := make([]byte, define.MAX_BUF)
		//PACKET_HEAD_LEN = 19
		headBuf := make([]byte, define.PACKET_HEAD_LEN)

		if this.CacheConn == nil {
			time.Sleep(30 * time.Second)
			continue
		}
		//读取头部信息
		_, err := (*this.CacheConn).Read(headBuf)
		if err != nil && err != io.EOF {
			continue
		} else if err == io.EOF {
			return
		}
		head := structure.Head{}
		err = head.Decode(headBuf)
		if err != nil {
			err = errors.New("component: CacheComponent.ReceiveFromCacheDB() failed, " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		if head.DataLen > 0 {
			if int(head.DataLen) > len(data) {
				data = make([]byte, head.DataLen)
			}
			//读取Packet的Data数据
			_, err = (*this.CacheConn).Read(data[:head.DataLen])
			if err != nil && err != io.EOF {
				err = errors.New("component: CacheComponent.ReceiveFromCacheDB() failed, " + err.Error())
				fmt.Println(err.Error())
				continue
			} else if err == io.EOF {
				fmt.Println("Got io.EOF")
				return
			}
		}

		packet := structure.Packet{head, data[:head.DataLen]}
		this.dispatch(&packet)
	}
}

func (this *CacheComponent) dispatch(packet *structure.Packet) {
	switch packet.Number {
	case define.FORWARD:
		GetDaughterHandler().PushToDaughter(packet)
	//response
	case define.ADD_CACHE_ENTRY_RESPONSE:
		this.CacheRespPush(packet)
	}
}

func (this *CacheComponent) SendOut() {
	for {
		packet := this.CachePop()
		if packet == nil {
			err := errors.New("component: CacheComponent.SendOut() failed, got a nil packet from CacheComponent.CacheQueue")
			fmt.Println(err.Error())
			continue
		}
		buf, err := packet.Encode()
		if err != nil {
			err := errors.New("component: CacheComponent.SendOut() failed, " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		if this.CacheConn == nil {
			time.Sleep(30 * time.Second)
			continue
		}
		_, err = (*this.CacheConn).Write(buf)
		if err != nil && err != io.EOF {
			err := errors.New("component: CacheComponent.SendOut() failed, " + err.Error())
			fmt.Println(err.Error())
			continue
		}
		resp := this.CacheRespPop()
		if resp.Uuid != packet.Uuid {
			err := errors.New("component: CacheComponent.SendOut() failed, Got a response with wrong uuid from DB")
			fmt.Println(err.Error())
			continue
		}
		if len(resp.Data) > 0 && uint8(resp.Data[0]) == 0 {
			err := errors.New("component: CacheComponent.SendOut() failed, DB did not save last entry")
			fmt.Println(err.Error())
		} else if len(resp.Data) > 0 && uint8(resp.Data[0]) == 1 {
			fmt.Println("component: CacheComponent.SendOut() OK, DB saved a cache-entry")
		}
	}
}

func (this *CacheComponent) SendToCacheDB(packet *structure.Packet) error {
	if packet == nil {
		err := errors.New("component: CacheComponent.SendToCacheDB() failed, the argument is a nil packet")
		fmt.Println(err.Error())
		return err
	}
	this.CachePush(packet)
	return nil
}

//建立10.0.0.11：11003上的tcp连接
func (this *CacheComponent) ConnCache() error {
	//10.0.0.11：11003
	addr := this.CacheIP + ":" + strconv.Itoa(int(this.CachePort))
	//建立10.0.0.11：11003上的tcp连接
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		err = errors.New("component: CacheComponent.ConnCache() failed, " + err.Error())
		fmt.Println(err.Error())
		return err
	}
	this.CacheConn = &conn
	return nil
}
