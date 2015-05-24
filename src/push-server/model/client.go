package model

import (
	"errors"
	"math/rand"
	"push-server/protocol"
	"sync"
	"time"
)

type DeviceIds map[uint32]Tokens
type ClientUuids map[string]DeviceIds
type ClientNumbers map[uint16]ClientUuids

type Clients struct {
	//map[Message.Number]map[Meesage.Uuid]map[Message.DeviceId]map[Message.Token]*Client
	Clients ClientNumbers //map[uint16]map[string]map[uint32]map[uint8]*Client
	Lock    sync.RWMutex
}

func NewClients() *Clients {
	return &Clients{Clients: make(ClientNumbers, 0)}
}

//返回token
func (this *Clients) Add(Number uint16, Uuid string, DeviceId uint32, client *Client) uint8 {
	this.Lock.Lock()

	client.Status = false
	client.Time = time.Now()

	if this.Clients[Number] == nil {
		this.Clients[Number] = make(ClientUuids, 0)
	}

	if this.Clients[Number][Uuid] == nil {
		this.Clients[Number][Uuid] = make(DeviceIds, 0)
	}

	if this.Clients[Number][Uuid][DeviceId] == nil {
		this.Clients[Number][Uuid][DeviceId] = make(Tokens, 0)
	}

	var token uint8
	if v, ok := this.Clients[Number][Uuid][DeviceId]; ok {
		if len(v) >= 255 {
			this.Lock.Unlock()
			return token
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for {
			token = uint8(r.Intn(define.TOKEN_MAX))
			if token == 0 {
				continue
			}
			if _, ok := v[token]; !ok {
				v[token] = client
				break
			}

		}
	}

	this.Lock.Unlock()
	return token
}

func (this *Clients) Remove(Number uint16, Uuid string, DeviceId uint32, Token uint8) error {

	this.Lock.Lock()

	if v, ok := this.Clients[Number][Uuid][DeviceId]; ok {
		if _, ok := v[Token]; ok {
			if v[Token] != nil && v[Token].Conn != nil {
				(*v[Token].Conn).Close()
			}
			delete(v, Token)
			if len(v) == 0 {
				delete(this.Clients[Number][Uuid], DeviceId)
				if len(this.Clients[Number][Uuid]) == 0 {
					delete(this.Clients[Number], Uuid)
					if len(this.Clients[Number]) == 0 {
						delete(this.Clients, Number)
					}
				}
			}
			this.Lock.Unlock()
			return nil
		}
	}

	this.Lock.Unlock()
	return errors.New("Clients Remove Error: do not have the client")
}

func (this *Clients) DeleteUuid(Number uint16, Uuid string) error {
	this.Lock.Lock()

	if _, ok := this.Clients[Number][Uuid]; ok {
		delete(this.Clients[Number], Uuid)
		if len(this.Clients[Number] == 0) {
			delete(this.Clients, Number)
		}
		this.Lock.Unlock()
		return nil
	}
	this.Lock.Unlock()
	return errors.New("Clients DeleteUuid Error: do not have the Uuid")
}

func (this *Clients) DeleteDevice(Number uint16, Uuid string, DeviceId uint32) error {
	this.Lock.Lock()

	if _, ok := this.Clients[Number][Uuid][DeviceId]; ok {
		delete(this.Clients[Number][Uuid], DeviceId)
		if len(this.Clients[Number][Uuid]) == 0 {
			delete(this.Clients[Number], Uuid)
			if len(this.Clients[Number] == 0) {
				delete(this.Clients, Number)
			}
		}
		this.Lock.Unlock()
		return nil
	}

	this.Lock.Unlock()
	return errors.New("Clients DeleteDevice Error : do not have the DeviceId")
}

func (this *Clients) UpdateStatus(Number uint16, Uuid string, DeviceId uint32, token uint8, Status bool) error {
	this.Lock.Lock()

	if _, ok := this.Clients[Number][Uuid][DeviceId][token]; ok {
		this.Clients[Number][Uuid][DeviceId][token].Status = Status
		this.Clients[Number][Uuid][DeviceID][token].Time = time.Now()
		this.Lock.Unlock()
		return nil
	}

	this.Lock.Unlock()
	return errors.New("Clients UpdateStatus Error: do not have the client")
}

func (this *Clients) QueryForNumber(Number uint16) []*Client {
	this.Lock.Lock()

	clients := make([]*Client, 0)

	for key, number_value := range this.Clients {
		if key&Number > 0 {
			for _, uuids_value := range number_value { //uuids map[string]map[string]map[int]*Client
				for _, devices_value := range uuids_value { //devices map[string]map[int]*Client
					for _, token_value := range devices_value { //token map[int]*Client
						clients = append(clients, token_value)
					}
				}
			}
		}
	}
	this.Lock.Unlock()
	return clients
}

func (this *Clients) QueryForUuid(Number uint16, Uuid string) []*Client {
	this.Lock.Lock()

	clients := make([]*Client, 0)
	for key, numbers := range this.Clients {
		if key&Number > 0 {
			if uuids, ok := numbers[Uuid]; ok {
				for _, devices := range uuids {
					for _, token := range devices {
						clients = append(clients, token)
					}
				}
			}
		}
	}

	this.Lock.Unlock()
	return clients
}

func (this *Clients) QueryForDeviceId(Number uint16, Uuid string, DeviceId uint32) []*Client {
	this.Lock.Lock()

	clients := make([]*Client, 0)
	for key, _ := range this.Clients {
		if key&Number > 0 {
			if deviceIds, ok := this.Clients[key][Uuid][DeviceId]; ok {
				for _, value := range deviceIds {
					clients = append(clients, value)
				}
			}
		}
	}

	this.Lock.Unlock()
	return client
}
