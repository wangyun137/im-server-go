package component

import (
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"push-server/buffpool"
	"push-server/protocol"
	"sync"
	"time"
)

//const HEAD_RESET = 0xffff >> 2

type CacheComponent struct {
	Cache     map[string]CacheDev
	CacheLock sync.RWMutex
	RespQueue chan *protocol.Message
}

var cacheComponentSigleton CacheComponent

func NewCacheComponent() *CacheComponent {
	return &cacheComponentSigleton
}

func GetCacheComponent() *CacheComponent {
	return &cacheComponentSigleton
}

func (this *CacheComponent) Init() {
	this.CacheLock.Lock()
	if this.Cache == nil {
		this.Cache = make(map[string]CacheDev, 0)
	}
	this.CacheLock.Unlock()
	return
}

func (this *CacheComponent) Start() {
	return
}

type CacheDev struct {
	MDev map[uint32]*CacheList
}

type CacheList struct {
	List  *list.List
	Total uint32
}

type CacheEntry struct {
	protocol.Message
	ID uint32
}

func (this *CacheComponent) Save(msg protocol.Message) error {
	this.CacheLock.Lock()
	defer this.CacheLock.Unlock()
	if this.Cache == nil {
		this.Cache = make(map[string]CacheDev, 0)
	}

	if _, ok := this.Cache[msg.DestUuid]; !ok {
		val := CacheDev{
			MDev: make(map[uint32]*CacheList, 0),
		}
		this.Cache[msg.DestUuid] = val
	}

	if _, ok := this.Cache[msg.DestUuid].MDev[msg.DeviceId]; !ok {
		val := &CacheList{
			List:  list.New(),
			Total: 0,
		}
		this.Cache[msg.DestUuid].MDev[msg.DeviceId] = val
	}

	this.Cache[msg.DestUuid].MDev[msg.DeviceId].Total++
	id := this.Cache[msg.DestUuid].MDev[msg.DeviceId].Total
	msg.DataLen = 4 + msg.DataLen
	buf := make([]byte, msg.DataLen)
	binary.BigEndian.PutUint32(buf[0:4], id)
	if len(msg.Data) > 0 {
		copy(buf[4:msg.DataLen], msg.Data)
	}
	msg.Data = buf

	msg.MsgType = protocol.MT_PUSH_WITH_RESPONSE
	msg.PushType = protocol.PT_PUSHSERVER
	msg.Number &= HEAD_RESET

	e := this.Cache[msg.DestUuid].MDev[msg.DeviceId].List.PushBack(CacheEntry{msg, id})
	if e == nil {
		err := errors.New("CacheComponent.Save() Error: can not save msg to list")
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (this *CacheComponent) SendCached(destUuid string, deviceID uint32) error {

	this.CacheLock.RLock()
	defer this.CacheLock.RUnlock()

	if !this.CheckCached(destUuid, deviceID) {
		fmt.Println("\n", time.Now().String(), " CacheComponent.SendCached(), no cached messages ", "\n")
		return nil
	}

	dispatcherOut := GetDispatcherOutComponent()
	buffPool := buffpool.GetBuffPool()

	for e := this.Cache[destUuid].MDev[deviceID].List.Front(); e != nil; e = e.Next() {
		val, ok := e.Value.(CacheEntry)
		if !ok {
			continue
		}
		err := dispatcherOut.sendToClient(&val.Message, buffPool)
		if err != nil {
			fmt.Println("CacheComponent.SendCached() sendToClient Error: " + err.Error())
			return err
		}
	}
	return nil
}

func (this *CacheComponent) CheckCached(destUuid string, deviceID uint32) bool {
	if _, ok1 := this.Cache[destUuid]; !ok1 {
		fmt.Println("\n", time.Now().String(), " exit CacheComponent.CheckCached", "\n")
		return false
	}
	if _, ok2 := this.Cache[destUuid].MDev[deviceID]; !ok2 {
		fmt.Println("\n", time.Now().String(), " exit CacheComponent.CheckCached", "\n")
		return false
	}

	if this.Cache[destUuid].MDev[deviceID] == nil || this.Cache[destUuid].MDev[deviceID].List == nil {
		fmt.Println("\n", time.Now().String(), " exit CacheComponent.CheckCached", "\n")
		return false
	}
	return true
}

func (this *CacheComponent) ReduceCached(destUuid string, deviceID uint32) {
	if !this.CheckCached(destUuid, deviceID) {
		fmt.Println("\n", time.Now().String(), " exit ReduceCached", "\n")
		return
	}
	//如果以deviceId为键的链表为空，删除以deviceId为键的链表
	if this.Cache[destUuid].MDev[deviceID].List.Len() == 0 {
		delete(this.Cache[destUuid].MDev, deviceID)
	}
	//如果以DestId为键的MDex映射为空，删除这个映射
	if len(this.Cache[destUuid].MDev) == 0 {
		delete(this.Cache, destUuid)
	}
}

func (this *CacheComponent) RemoveCached(resp *protocol.Message) error {
	if resp.MsgType != protocol.MT_PUSH_WITH_RESPONSE || resp.Number != protocol.N_ACCOUNT_SYS {
		err := errors.New("CacheComponent RemoveCached Error: Wrong response from client")
		fmt.Println(err.Error())
		return err
	}
	userUuid := resp.UserUuid
	deviceID := resp.DeviceId
	id := binary.BigEndian.Uint32(resp.Data[0:4])

	this.CacheLock.Lock()
	defer this.CacheLock.Unlock()

	if !this.CheckCached(userUuid, deviceID) {
		err := errors.New("CacheComponent RemoveCached Error:  CachedEntry already deleted")
		fmt.Println(err.Error())
		return nil
	}
	//遍历链表，断言链表中每一个节点的Value为CacheEntry类型，如果是删除这个节点
	for e := this.Cache[userUuid].MDev[deviceID].List.Front(); e != nil; e = e.Next() {
		val, ok := e.Value.(CacheEntry)
		if !ok {
			continue
		}
		if id != val.ID {
			continue
		}
		this.Cache[userUuid].MDev[deviceID].List.Remove(e)
		break
	}

	this.ReduceCached(userUuid, deviceID)
	return nil
}
