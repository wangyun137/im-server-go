package model

import (
	"errors"
	"net"
	"sync"
)

type User struct {
	IP           net.IP
	Port         uint16
	BindUuid     string
	BindDeviceId uint32
	Conn         *net.UDPAddr
	Lock         sync.RWMutex
}

func NewUser(BindUuid string, BindDeviceId uint32, IP net.IP, Port uint16, Conn *net.UDPAddr) *User {
	return &User{
		BindUuid:     BindUuid,
		BindDeviceId: BindDeviceId,
		IP:           IP,
		Port:         Port,
		Conn:         Conn,
	}
}

type Users struct {
	Users map[string]map[uint32]*User
	Lock  sync.RWMutex
}

var U *Users

func NewUsers() *Users {
	return &Users{
		Users: make(map[string]map[uint32]*User),
	}
}

func GetUsers() *Users {
	if U == nil {
		U = NewUsers()
	}

	return U
}

func (this *Users) Add(uuid string, deviceId uint32, destUuid string, destDeviceId uint32, IP net.IP, Port uint16, Conn *net.UDPAddr) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if _, ok := this.Users[uuid]; !ok {
		this.Users[uuid] = make(map[uint32]*User)
	}

	if _, ok := this.Users[uuid][deviceId]; !ok {
		this.Users[uuid][deviceId] = NewUser(destUuid, destDeviceId, IP, Port, Conn)
		return nil
	}

	return errors.New("Model Add Error : uuid and device is exist")
}

func (this *Users) Update(uuid string, deviceId uint32, destUuid string, destDeviceId uint32, IP net.IP, Port uint16, Conn *net.UDPAddr) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if _, ok := this.Users[uuid][deviceId]; ok {
		this.Users[uuid][deviceId] = NewUser(destUuid, destDeviceId, IP, Port, Conn)
		return nil
	}

	return errors.New("Model Update Error : not found uuid")
}

func (this *Users) AddConn(uuid string, deviceId uint32, Conn *net.UDPAddr) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if user, ok := this.Users[uuid][deviceId]; ok {
		user.Conn = Conn
		return nil
	}

	return errors.New("Model AddConn Error: uuid and device is exists")
}

func (this *Users) Remove(uuid string, deviceId uint32) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if _, ok := this.Users[uuid][deviceId]; ok {
		delete(this.Users[uuid], deviceId)
		if len(this.Users[uuid]) == 0 {
			delete(this.Users, uuid)
		}

		return nil
	}

	return errors.New("Model Remove Error: uuid and device is exists")
}

func (this *Users) Get(uuid string, deviceId uint32) *User {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if user, ok := this.Users[uuid][deviceId]; ok {
		return user
	}

	return nil
}
