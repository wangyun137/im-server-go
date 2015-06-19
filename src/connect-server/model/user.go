package model

import (
	"errors"
	"net"
	"sync"
)

type User struct {
	Uuid       string
	IP         []byte
	Port       int32
	Conn       *net.Conn
	ConnRWLock sync.Mutex
}

func (user *User) Write(buffer []byte) (int, error) {
	user.ConnRWLock.Lock()
	defer user.ConnRWLock.Unlock()
	n, err := (*user.Conn).Write(buffer)
	return n, err
}

type UserCache struct {
	Lock      sync.Mutex
	UserCache map[string]*User
}

func NewUserCache() *UserCache {
	return &UserCache{UserCache: make(map[string]*User, 0)}
}

func (this *UserCache) Add(user *User) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	this.UserCache[user.Uuid] = user
}

func (this *UserCache) Remove(uuid string) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if _, ok := this.UserCache[uuid]; ok {
		delete(this.UserCache, uuid)
		return nil
	}

	return errors.New("UserCache remove error: do not have " + uuid)
}

func (this *UserCache) Update(user *User) error {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if v, ok := this.UserCache[user.Uuid]; ok {
		v.IP = user.IP
		v.Port = user.Port
		return nil
	}
	return errors.New("UserCache update error: " + user.Uuid)
}

func (this *UserCache) Query(uuid string) *User {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if user, ok := this.UserCache[uuid]; ok {
		return user
	}
	return nil
}
