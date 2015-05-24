package model

import (
	"errors"
	"math/rand"
	"net"
	"push-server/protocol"
	"sync"
	"time"
)

type ServiceUuids map[string]*net.Conn
type ServiceNumbers map[uint16]ServiceUuids

type Services struct {
	Services ServiceNumbers
	Lcok     sync.RWMutex
}

func NewServices() *Services {
	return &Services{Services: make(ServiceNumbers, 0)}
}

func (this *Services) Add(Number uint16, Uuid string, Conn *net.Conn) error {
	this.Lock.Lock()

	if this.Services[Number] == nil {
		this.Services[Number] = make(ServiceUuids, 0)
	}

	if _, ok := this.Services[Number][Uuid]; !ok {
		this.Services[Number][Uuid] = Conn
		this.Lock.Unlock()
		return nil
	}

	this.Lock.Unlock()
	return errors.New("Services Add Error : exist Uuid " + Uuid)
}

func (this *Services) Remove(Number uint16, Uuid string) error {

	this.Lock.Lock()

	if _, ok := this.Services[Number][Uuid]; ok {
		if this.Services[Number][Uuid] != nil {
			(*this.Services[Number][Uuid]).Close()
		}
		delete(this.Services[Number], Uuid)
		if len(this.Services[Number]) == 0 {
			delete(this.Services, Number)
		}
		this.Lock.Unlock()
		return nil
	}

	this.Lock.Unlock()
	return errors.New("Services Remove Error: do not find the Uuid " + Uuid)
}

func (this *Services) Query(Number uint16, Uuid string) *net.Conn {

	this.Lock.Lock()

	if v, ok := this.Services[Number][Uuid]; ok {
		this.Lock.Unlock()
		return v
	}

	this.Lock.Unlock()
	return nil
}

func (this *Services) Get(Number uint16) *net.Conn {

	this.Lock.Lock()

	for _, v := range this.Services[Number] {
		this.Lock.Unlock()
		return v
	}

	this.Lock.Unlock()
	return nil
}

func (this *Services) GetAllForNumber(Number uint16) ServiceUuids {

	this.Lock.Lock()

	if v, ok := this.Services[Number]; ok {
		this.Lock.Unlock()
		return v
	}

	this.Lock.Unlock()
	return nil
}

func (this *Services) GetAllConn() []*net.Conn {
	services := make([]*net.Conn, 0)

	this.Lock.Lock()
	for _, v := range this.Services {
		for _, value := range v {
			services = append(services, value)
		}
	}

	this.Lock.Unlock()
	return services
}

func (this *Services) GetAll() map[string]*net.Conn {
	services := make(map[string]*net.Conn, 0)
	this.Lock.Lock()

	for _, v := range this.Services {
		for k, value := range v {
			services[k] = value
		}
	}

	this.Lock.Unlock()
	return services
}

func (this *Services) GetAllServices() map[uint16][]string {
	services := make(map[uint16][]string, 0)
	this.Lock.RLock()
	for num, vmap := range this.Services {
		for uuid, _ := range vmap {
			services[num] = append(services[num], uuid)
		}
	}
	this.Lock.RUnlock()
	return services
}
