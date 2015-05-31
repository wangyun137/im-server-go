package component

import (
	"connect-server/utils/config"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"push-server/protocol"
	"runtime"
	"strconv"
	"sync"
	"time"
	"tree-server/define"
	"tree-server/structure"
)

const (
	USERSNUMBER_LENGTH   = 4
	MAXCONNNUMBER_LENGTH = 4
	CONNNUMBER_LENGTH    = 4
)

type NodeComponent struct {
	StreeID           string
	MainStreeIP       string
	MainStreePort     int
	SpareStreeIP      string
	SpareStreePort    int
	GOMAXPROCS        int
	StreeConn         *net.TCPConn
	HeartBeatTime     int
	CacheSize         int
	ReConnectTime     int
	SendMsgs          chan *structure.Packet
	QueryMsgs         chan *structure.Packet
	ForwardMsgs       chan *structure.Packet
	StreeConnLock     sync.RWMutex
	UsersNumber       uint32
	MaxConnNumber     uint32
	ConnNumber        uint32
	UsersNumberLock   sync.RWMutex
	MaxConnNumberLock sync.RWMutex
	ConnNumberLock    sync.RWMutex
}

var Node *NodeComponent

//生成一个字段全部为空值的NodeComponent
func NewNodeComponent() *NodeComponent {
	return &NodeComponent{}
}

func GetNodeComponent() *NodeComponent {
	if Node == nil {
		Node = NewNodeComponent()
	}

	return Node
}

func (this *NodeComponent) SetUsersNumber(number uint32) {
	this.UsersNumberLock.Lock()
	defer this.UsersNumberLock.Unlock()

	this.UsersNumber = number
}

func (this *NodeComponent) GetUsersNumber() uint32 {
	this.UsersNumberLock.RLock()
	defer this.UsersNumberLock.RUnlock()

	return this.UsersNumber
}

func (this *NodeComponent) SetMaxConnNumber(number uint32) {
	this.MaxConnNumberLock.Lock()
	defer this.MaxConnNumberLock.Unlock()

	this.MaxConnNumber = number
}

func (this *NodeComponent) GetMaxConnNumber() uint32 {
	this.MaxConnNumberLock.RLock()
	defer this.MaxConnNumberLock.RUnlock()

	return this.MaxConnNumber
}

func (this *NodeComponent) SetConnNumber(number uint32) {
	this.ConnNumberLock.Lock()
	defer this.ConnNumberLock.Unlock()

	this.ConnNumber = number
}

func (this *NodeComponent) GetConnNumber() uint32 {
	this.ConnNumberLock.RLock()
	defer this.ConnNumberLock.RUnlock()

	return this.ConnNumber
}

func (this *NodeComponent) ConfInit() error {
	var err error
	config := config.NewConfig()

	err = config.Read("conf/node.conf")
	if err != nil {
		fmt.Println("NodeComponent Error: read conf file error")
		return errors.New("NodeComponent Error:ConfInit")
	}
	//conf里没有
	this.StreeID, err = config.GetString("StreeID")
	if err != nil {
		fmt.Println("NodeComponent Error:Parse StreeID")
		return errors.New("NodeComponent Error: StreeID")
	}
	//MainStreeIP:10.0.0.7
	this.MainStreeIP, err = config.GetString("MainStreeIP")
	if err != nil {
		fmt.Println("NodeComponent Error:Parse MainStreeIP")
		return errors.New("NodeComponent Error: ConfInit")
	}
	//MainStreePort:10000
	this.MainStreePort, err = config.GetInt("MainStreePort")
	if err != nil {
		fmt.Println("NodeComponent Error:Parse MainStreePort")
		return errors.New("NodeComponent Error: ConfInit")
	}
	//SpareStreeIP:10.0.0.7
	this.SpareStreeIP, err = config.GetString("SpareStreeIP")
	if err != nil {
		fmt.Println("NodeComponent Error:Parse SpareStreePort")
		return errors.New("NodeComponent Error: ConfInit")
	}
	//SpareStreePort:10001
	this.SpareStreePort, err = config.GetInt("SpareStreePort")
	if err != nil {
		fmt.Println("NodeComponent Error: Parse SpareStreePort")
		return errors.New("NodeComponent Error: ConfInit")
	}
	//conf里无，HeartBeatTime = 1
	this.HeartBeatTime, err = config.GetInt("HeartBeatTime")
	if err != nil {
		fmt.Println("NodeComponent Error: Parse HeartBeatTime, Use Heart Beat Time Default Value 1")
		this.HeartBeatTime = 1
	}
	//conf里无，CacheSize = 10000
	this.CacheSize, err = config.GetInt("CacheSize")
	if err != nil {
		fmt.Println("NodeComponent Error: Parse CacheSize, Use Cache Size Default Value 10000")
		this.CacheSize = 10000
	}
	//conf里无，ReconnectTime = 1
	this.ReConnectTime, err = config.GetInt("ReConnectTime")
	if err != nil {
		fmt.Println("NodeComponent Error: Parse ReConnectTime, Use ReConnectTime Default Value 1")
		this.ReConnectTime = 1
	}
	//conf里无，GOMAXPROCS为当前最大CPU数
	this.GOMAXPROCS, err = config.GetInt("GOMAXPROCS")
	if err != nil {
		fmt.Println("NodeComponent Error: Parse GOMAXPROCS, Use GOMAXPROCS Default Value 8")
		this.GOMAXPROCS = runtime.NumCPU()
	}

	runtime.GOMAXPROCS(this.GOMAXPROCS)
	return nil
}
func (this *NodeComponent) LogInfo() {
	fmt.Println("MainStreeIP	:	" + this.MainStreeIP)
	fmt.Println("MainStreePort	:	" + strconv.Itoa(this.MainStreePort))
	fmt.Println("SpareStreeIP	:	" + this.SpareStreeIP)
	fmt.Println("SpareStreePort	:	" + strconv.Itoa(this.SpareStreePort))
	fmt.Println("HeartBeatTime	:	" + strconv.Itoa(this.HeartBeatTime))
	fmt.Println("CacheSize	:	" + strconv.Itoa(this.CacheSize))
}

func (this *NodeComponent) Init() {
	this.ConfInit()
	//
	if this.SendMsgs == nil {
		this.SendMsgs = make(chan *structure.Packet, this.CacheSize)
	}

	if this.QueryMsgs == nil {
		this.QueryMsgs = make(chan *structure.Packet, this.CacheSize)
	}

	if this.ForwardMsgs == nil {
		this.ForwardMsgs = make(chan *structure.Packet, this.CacheSize)
	}
}

func (this *NodeComponent) ConnectMainStree() error {
	if this.MainStreeIP == "" {
		return errors.New("NodeComponent Error: MainStreeIP is empty")
	}

	var err error
	this.StreeConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(this.MainStreeIP),
		Port: this.MainStreePort,
	})

	if err != nil {
		fmt.Println("NodeComponent Error: Connect Main STree")
		return err
	}
	return nil
}

func (this *NodeComponent) ConnectSpareStree() error {
	if this.SpareStreeIP == "" {
		return errors.New("NodeComponent Error:SpareStreeIP is empty")
	}
	var err error
	this.StreeConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(this.SpareStreeIP),
		Port: this.SpareStreePort,
	})

	if err != nil {
		fmt.Println("NodeComponent Error: Connect Spare STree")
		return err
	}

	return nil
}

func (this *NodeComponent) Start() {

	for {
		err := this.ReConnect()
		if err == nil {
			break
		}
		<-time.After(time.Second * time.Duration(this.ReConnectTime))
	}

	go this.HeartBeat()
	go this.ReceiveMsg()
}

func (this *NodeComponent) Push(msg *structure.Packet) {
	this.SendMsgs <- msg
}

func (this *NodeComponent) Pop() *structure.Packet {
	return <-this.SendMsgs
}

func (this *NodeComponent) Handle() {
	for {
		this.MsgHandle(this.Pop())
	}
}

func (this *NodeComponent) MsgHandle(msg *structure.Packet) {
	buff, err := msg.Encode()
	if err != nil {
		fmt.Println("NodeComponent Error: Encode Msg Packet")
		return
	}

	n, err := this.Write(buff)
	if n == 0 || err != nil {
		fmt.Println("NodeComponent Error: Write Msg Packet")
		return
	}
}

func (this *NodeComponent) HeartBeat() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("NodeComponent Error: " + err.(error).Error())
		}
	}()

	var DataLength uint16
	//DataLength = 12
	DataLength = USERSNUMBER_LENGTH + MAXCONNNUMBER_LENGTH + CONNNUMBER_LENGTH
	//GetPacket(Number uint8, Uuid string, Data []byte) *structure.Packet
	Packet := GetPacket(define.HEART_BEAT, this.StreeID, make([]byte, DataLength))

	for {
		//add push server info
		index := 0
		//0-4是UsersNumber
		copy(Packet.Data[index:index+USERSNUMBER_LENGTH], Int32ParseBytes(this.UsersNumber))
		index = index + USERSNUMBER_LENGTH
		//4-8是MaxConnNumber
		copy(Packet.Data[index:index+MAXCONNNUMBER_LENGTH], Int32ParseBytes(this.MaxConnNumber))
		index = index + MAXCONNNUMBER_LENGTH
		//8-12是ConnNumber
		copy(Packet.Data[index:index+CONNNUMBER_LENGTH], Int32ParseBytes(this.ConnNumber))
		index = index + CONNNUMBER_LENGTH
		//Packet共有31位
		buff, err := Packet.Encode()
		if err != nil {
			fmt.Println("NodeComponent Error: Encode Heart Beat Packet")
			continue
		}

		n, err := this.Write(buff)
		//如果出错则再次尝试建立tcp连接
		if n == 0 || err != nil {
			fmt.Println("NodeComponent Error: Write Heart Beat Packet")
			err := this.ReConnect()
			if err != nil {
				<-time.After(time.Second * time.Duration(this.ReConnectTime))
			}

			continue
		}

		<-time.After(time.Second * time.Duration(this.HeartBeatTime))
	}
}

func (this *NodeComponent) ReceiveMsg() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("NodeComponent Error: " + err.(error).Error())
		}
	}()

	headbuff := make([]byte, define.PACKET_HEAD_LEN)
	for {

		var head structure.Head
		var Packet structure.Packet
		//如果无连接，尝试重新建立tcp连接
		if this.StreeConn == nil {
			this.ReConnect()
			continue
		}
		//从tcp连接中先读取头部信息
		n, err := this.StreeConn.Read(headbuff)
		//如果读取出错，则尝试重新建立tcp连接
		if n == 0 || n != define.PACKET_HEAD_LEN || err != nil {
			fmt.Println("NodeComponent Error:Read STree msg")
			this.ReConnect()
			continue
		}

		err = head.Decode(headbuff)
		if err != nil {
			fmt.Println("NodeComponent Error: Decode Stree msg")
			return
		}

		databuff := make([]byte, head.DataLen)
		//再从连接中读取剩余数据
		n, err = this.StreeConn.Read(databuff)
		if n == 0 || n != int(head.DataLen) || err != nil {
			fmt.Println("NodeComponent Error: Read Stree Data")
			this.ReConnect()
			continue
		}

		Packet.Head = head
		Packet.Data = databuff
		msg := protocol.Message{}
		//Packet.Data至少44位
		err = msg.Decode(Packet.Data)
		if err != nil {
			fmt.Println("NodeComponent Error:MSG Decode Failed", err)
			return
		}

		this.ReceiveMsgHandle(&Packet)
	}
}

func (this *NodeComponent) ReceiveMsgHandle(msg *structure.Packet) {
	switch msg.Number {
	case define.QUERY_USER:
		this.QueryPush(msg)
	case define.FORWARD, define.FORWARD_WITH_CACHE:
		this.ForwardPush(msg)
	}
}

func (this *NodeComponent) QueryPush(msg *structure.Packet) {
	this.QueryMsgs <- msg
}

func (this *NodeComponent) QueryPop() *structure.Packet {
	return <-this.QueryMsgs
}

func (this *NodeComponent) ForwardPush(msg *structure.Packet) {
	m := protocol.Message{}
	err := m.Decode(msg.Data)
	if err != nil {
		fmt.Println("NodeComponent Error:MSG Decode Failed", err)
		return
	}
	this.ForwardMsgs <- msg
}

func (this *NodeComponent) ForwardPop() *structure.Packet {
	return <-this.ForwardMsgs
}

func (this *NodeComponent) Login(Uuid string) {
	Packet := GetPacket(define.ONLINE, Uuid, make([]byte, 0))

	buff, err := Packet.Encode()
	if err != nil {
		fmt.Println("NodeComponent Error: Encode Login Packet")
		return
	}

	n, err := this.Write(buff)
	if n == 0 || err != nil {
		return
	}

}

func (this *NodeComponent) Logout(Uuid string) {
	Packet := GetPacket(define.OFFLINE, Uuid, make([]byte, 0))

	buff, err := Packet.Encode()
	if err != nil {
		fmt.Println("NodeComponent Error: Encode Logout Packet")
		return
	}

	n, err := this.Write(buff)
	if n == 0 || err != nil {
		return
	}
}

func (this *NodeComponent) Query(Uuid string) {
	Packet := GetPacket(define.QUERY_USER, Uuid, make([]byte, 0))

	buff, err := Packet.Encode()
	if err != nil {
		return
	}

	n, err := this.Write(buff)
	if n == 0 || err != nil {
		return
	}
}

func (this *NodeComponent) NoCacheForward(Uuid string, Data []byte) {
	Packet := GetPacket(define.FORWARD, Uuid, Data)

	buff, err := Packet.Encode()
	if err != nil {
		fmt.Println("NodeComponent Error:Encode No Cache Forward Packet")
		return
	}

	n, err := this.Write(buff)
	if n == 0 || err != nil {
		return
	}

	msg := protocol.Message{}
	err = msg.Decode(Packet.Data)
	if err != nil {
		fmt.Println("NodeComponent Error: MSG Decode Failed, ", err)
		return
	}
}

func (this *NodeComponent) CacheForward(Uuid string, Data []byte) {
	Packet := GetPacket(define.FORWARD_WITH_CACHE, Uuid, Data)

	buff, err := Packet.Encode()
	if err != nil {
		fmt.Println("NodeComponent Error: Encode Cache Forward Packet")
		return
	}

	n, err := this.Write(buff)
	if n == 0 || err != nil {
		return
	}

}

func (this *NodeComponent) Write(buff []byte) (int, error) {
	this.StreeConnLock.Lock()
	defer this.StreeConnLock.Unlock()

	if this.StreeConn != nil {
		n, err := this.StreeConn.Write(buff)
		if n == 0 || err != nil {
			err := this.ReConnect()
			if err != nil {
				fmt.Println("NodeComponent Error: Stree Not Connect")
				return 0, nil
			}

			n, err = this.StreeConn.Write(buff)
			return n, err
		}

		return n, err
	}

	return 0, errors.New("NodeComponent Error: Stree Connect is nil")
}

func (this *NodeComponent) Close() {
	this.StreeConnLock.Lock()
	defer this.StreeConnLock.Unlock()

	if this.StreeConn != nil {
		this.StreeConn.Close()
	}
}

func (this *NodeComponent) ReConnect() error {
	this.Close()

	err := this.ConnectMainStree()
	if err != nil {
		err = this.ConnectSpareStree()
		if err != nil {
			return err
		}
	}

	this.LogInfo()
	return nil
}

//初始化一个structure.Packet并返回
func GetPacket(Number uint8, Uuid string, Data []byte) *structure.Packet {
	return &structure.Packet{
		Head: structure.Head{
			Number:  Number,
			Uuid:    Uuid,
			DataLen: uint16(len(Data)),
		},

		Data: Data,
	}
}

//将uint32序列化到切片中
func Int32ParseBytes(number uint32) []byte {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, number)

	return buff
}
