package component

import (
	"errors"
	"net"

	"encoding/binary"
	"fmt"
	"sync"
	"time"
	"tree-server/define"
	"tree-server/structure"
)

type STree struct {
	//master/slave
	RunMode string
	//Root/Branch/Node
	Type string
	SelfInfo
	ServiceMap
	UserMap
	Seconds int
}

type ServiceMap struct {
	ServiceMap map[*net.Conn]ServiceInfo
	Lock       sync.RWMutex
}

type ServiceInfo struct {
	UserNumber        uint32
	MaxConnNumber     uint32
	CurrentConnNumber uint32
}

type SelfInfo struct {
	UserNumber        uint32
	MaxConnNumber     uint32
	CurrentConnNumber uint32
	Lock              sync.RWMutex
}

type UserMap struct {
	UserMap map[string]*net.Conn
	Lock    sync.RWMutex
}

func (this *STree) PrintSelfInfo() {
	this.SelfInfo.Lock.RLock()
	fmt.Println("\n")
	fmt.Println("\t", time.Now().String(), " SelfInfo : ")
	fmt.Println("UserNumber = ", this.SelfInfo.UserNumber)
	fmt.Println("MaxConnNumber = ", this.SelfInfo.MaxConnNumber)
	fmt.Println("CurrentConnNumber = ", this.SelfInfo.CurrentConnNumber, "\n")
	this.SelfInfo.Lock.RUnlock()
}

func (this *STree) PrintUserMap() {
	fmt.Println("\n")
	fmt.Println("\t", time.Now().String(), " UserMap : ")
	this.UserMap.Lock.RLock()
	fmt.Println(this.UserMap.UserMap)
	this.UserMap.Lock.RUnlock()
	fmt.Println("\n")
}

func (this *STree) PrintUserMapCount() {
	userCount := 0
	this.UserMap.Lock.Lock()
	for _, _ = range this.UserMap.UserMap {
		userCount++
	}
	this.UserMap.Lock.Unlock()
	fmt.Println("\n", time.Now().String(), " UserMapCount: ", userCount, "\n")
}

func (this *UserMap) Put(uuid string, conn *net.Conn) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	this.UserMap[uuid] = conn
	return nil
}

func (this *UserMap) Get(uuid string) *net.Conn {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	return this.UserMap[uuid]
}

func (this *UserMap) Delete(uuid string) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	delete(this.UserMap, uuid)
	return nil
}

var sTree STree

func GetSTree() *STree {
	return &sTree
}
func NewSTree(seconds int) (*STree, error) {
	if seconds <= 0 {
		err := errors.New("component: NewSTree() failed, queue length/seconds should be positive")
		fmt.Println(err.Error())
		return nil, err
	}
	stree := &sTree
	stree.Seconds = seconds
	return stree, nil
}

func (this *STree) Init() {
	if this.ServiceMap.ServiceMap == nil {
		this.ServiceMap.ServiceMap = make(map[*net.Conn]ServiceInfo, 0)
	}
	if this.UserMap.UserMap == nil {
		this.UserMap.UserMap = make(map[string]*net.Conn, 0)
	}
}

func (this *STree) Start() {
	if this.Type == "Root" {
	} else {
		go this.SendHeartBeat()
	}
}

func (this *STree) SendHeartBeat() {
	for {
		packet := structure.Packet{}
		packet.Number = define.HEART_BEAT
		packet.Uuid = define.P_UID
		packet.DataLen = 12
		packet.Data = make([]byte, packet.DataLen)

		this.SelfInfo.Lock.RLock()
		userNumber := this.SelfInfo.UserNumber
		maxConn := this.SelfInfo.MaxConnNumber
		currConn := this.SelfInfo.CurrentConnNumber
		this.SelfInfo.Lock.RUnlock()

		binary.BigEndian.PutUint32(packet.Data[0:4], uint32(userNumber))
		binary.BigEndian.PutUint32(packet.Data[4:8], uint32(maxConn))
		binary.BigEndian.PutUint32(packet.Data[8:12], uint32(currConn))
		GetParentHandler().SendToParent(&packet)
		time.Sleep(time.Duration(this.Seconds * 1000 * 1000 * 1000))
	}
}

func (this *STree) GetMinLoadConn() *net.Conn {
	this.ServiceMap.Lock.RLock()
	minLoad := float32(1.1)
	var conn *net.Conn
	for key, val := range this.ServiceMap.ServiceMap {
		load := float32(val.CurrentConnNumber) / float32(val.MaxConnNumber)
		if load < minLoad {
			minLoad = load
			conn = key
		}
	}
	this.ServiceMap.Lock.RUnlock()
	return conn
}

func (this *STree) UpdateSelfInfo() {
	this.SelfInfo.Lock.Lock()
	defer this.SelfInfo.Lock.Unlock()
	this.ServiceMap.Lock.RLock()
	defer this.ServiceMap.Lock.RUnlock()

	this.SelfInfo.UserNumber = 0
	this.SelfInfo.MaxConnNumber = 0
	this.SelfInfo.CurrentConnNumber = 0
	for _, v := range this.ServiceMap.ServiceMap {
		this.SelfInfo.UserNumber += v.UserNumber
		this.SelfInfo.MaxConnNumber += v.MaxConnNumber
		this.SelfInfo.CurrentConnNumber += v.CurrentConnNumber
	}
}

func (this *STree) addOnlineInfo(packet *structure.Packet, conn *net.Conn) {
	info := ServiceInfo{}
	this.ServiceMap.Lock.Lock()
	info.UserNumber = this.ServiceMap.ServiceMap[conn].UserNumber + 1
	info.CurrentConnNumber = this.ServiceMap.ServiceMap[conn].CurrentConnNumber + 1
	info.MaxConnNumber = this.ServiceMap.ServiceMap[conn].MaxConnNumber
	this.ServiceMap.ServiceMap[conn] = info
	this.ServiceMap.Lock.Unlock()

	this.UserMap.Put(packet.Uuid, conn)
	this.UpdateSelfInfo()
}

func (this *STree) addOfflineInfo(packet *structure.Packet, conn *net.Conn) {
	info := ServiceInfo{}
	this.ServiceMap.Lock.Lock()
	//	this.ServiceMap.ServiceMap[conn].UserNumber--
	//	this.ServiceMap.ServiceMap[conn].CurrentConnNumber--
	info.UserNumber = this.ServiceMap.ServiceMap[conn].UserNumber - 1
	info.CurrentConnNumber = this.ServiceMap.ServiceMap[conn].CurrentConnNumber - 1
	info.MaxConnNumber = this.ServiceMap.ServiceMap[conn].MaxConnNumber
	this.ServiceMap.ServiceMap[conn] = info
	this.ServiceMap.Lock.Unlock()

	this.UserMap.Delete(packet.Uuid)
	this.UpdateSelfInfo()
}
