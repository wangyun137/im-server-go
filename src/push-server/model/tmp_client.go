package model

import (
	"errors"
	"math/rand"
	"net"
	"push-server/protocol"
	"sync"
	"time"
)

type Client struct {
	Conn   *net.Conn
	Status bool
	Time   time.Time
}

type Tokens map[uint8]*Client
type Devices map[string]Tokens
type TmpClientUuids map[string]Devices
type TmpClientNumbers map[uint16]TmpClientUuids

type TmpClients struct {
	//map[Message.Number]map[Message.USerUuid]map[string(Message.Data)]map[Message.Token]
	TmpClients TmpClientNumbers //map[uint16]map[string]map[string]map[uint8]*Client
	Lock       sync.RWMutex
}

func NewTmpClients() *TmpClients {
	return &TmpClients{TmpClients: make(TmpClientNumbers, 0)}
}

//返回Token
func (this *TmpClients) Add(Number uint16, Uuid string, Device string, client *Client) uint8 {
	this.Lock.Lock()
	client.Status = false
	client.Time = time.Now()

	if this.TmpClients[Number] == nil {
		this.TmpClients[Number] = make(TmpClientUuids, 0)
	}
	if this.TmpClients[Number][Uuid] == nil {
		this.TmpClients[Number][Uuid] = make(Devices, 0)
	}
	if this.TmpClients[Number][Uuid][Device] == nil {
		this.TmpClients[Number][Uuid][Device] = make(Tokens, 0)
	}
	var token uint8
	if value, ok := this.TmpClients[Number][Uuid][Device]; ok {
		if len(value) >= 255 {
			this.Lock.Unlock()
			return token
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for {
			token = uint8(r.Intn(protocol.TOKEN_MAX))
			if token == 0 {
				continue
			}
			if _, ok := value[token]; !ok {
				value[token] = client
				break
			}
		}
	}
	this.Lock.Unlock()
	return token
}

//删除并关闭Conn
func (this *TmpClients) Remove(Number uint16, Uuid string, Device string, Token uint8) error {
	this.Lock.Lock()

	if value, ok := this.TmpClients[Number][Uuid][Device]; ok {
		if _, ok := value[Token]; ok {
			if value[Token] != nil && value[Token].Conn != nil {
				(*value[Token].Conn).Close()
			}
			delete(value, Token)
			if len(value) == 0 {
				delete(this.TmpClients[Number][Uuid], Device)
				if len(this.TmpClients[Number][Uuid]) == 0 {
					delete(this.TmpClients[Number], Uuid)
					if len(this.TmpClients[Number]) == 0 {
						delete(this.TmpClients, Number)
					}
				}
			}
			this.Lock.Unlock()
			return nil
		}
	}
	this.Lock.Unlock()
	return errors.New("TmpClients Remove Error:do not have the client")
}

//删除但不关闭Conn
func (this *TmpClients) Remove2(Number uint16, Uuid string, Device string, Token uint8) error {
	this.Lock.Lock()

	if v, ok := this.TmpClients[Number][Uuid][Device]; ok {
		if _, ok := v[Token]; ok {
			delete(v, Token)
			if len(v) == 0 {
				delete(this.TmpClients[Number][Uuid], Device)
				if len(this.TmpClients[Number][Uuid]) == 0 {
					delete(this.TmpClients[Number], Uuid)
					if len(this.TmpClients[Number]) == 0 {
						delete(this.TmpClients, Number)
					}
				}
			}
			this.Lock.Unlock()
			return nil
		}
	}

	this.Lock.Unlock()
	return errors.New("TmpClients Remove2 Error:do not have the client")
}

func (this *TmpClients) DeleteUuid(Number uint16, Uuid string) error {
	this.Lock.Lock()
	if value, ok := this.TmpClients[Number][Uuid]; ok {
		delete(value, Uuid)
		if len(this.TmpClients[Number]) == 0 {
			delete(this.TmpClients, Number)
		}
		this.Lock.Unlock()
		return nil
	}

	this.Lock.Unlock()
	return errors.New("TmpClients DeleteUuid Error : do not find the uuid")
}

func (this *TmpClients) DeleteDevice(Number uint16, Uuid string, Device string) error {
	this.Lock.Lock()

	if _, ok := this.TmpClients[Number][Uuid][Device]; ok {
		delete(this.TmpClients[Number][Uuid], Device)
		if len(this.TmpClients[Number][Uuid]) == 0 {
			delete(this.TmpClients[Number], Uuid)
			if len(this.TmpClients[Number]) == 0 {
				delete(this.TmpClients, Number)
			}
		}
		this.Lock.Unlock()
		return nil
	}
	this.Lock.Unlock()
	return errors.New("TmpClients DeleteUuid Error : do not find the device")
}

func (this *TmpClients) Query(Number uint16, Uuid string, Device string, Token uint8) *Client {
	this.Lock.Lock()

	for k, v := range this.TmpClients {
		if Number&k != 0 {
			if value, ok := v[Uuid][Device][Token]; ok {
				this.Lock.Unlock()
				return value
			}
		}
	}

	this.Lock.Unlock()
	return nil
}

func (this *TmpClients) UpdateStatus(Number uint16, Uuid string, Device string, Token uint8, Status bool) error {
	this.Lock.Lock()
	if _, ok := this.TmpClients[Number][Uuid][Device][Token]; ok {
		this.TmpClients[Number][Uuid][Device][Token].Status = Status
		this.TmpClients[Number][Uuid][Device][Token].Time = time.Now()
		this.Lock.Unlock()
		return nil
	}
	this.Lock.Unlock()
	return errors.New("TmpClient UpdateStatus Error")
}

func (this *TmpClients) QueryForNumber(Number uint16) []*Client {
	this.Lock.Lock()

	clients := make([]*Client, 0)
	if n_value, ok := this.TmpClients[Number]; ok {
		for _, u_value := range n_value {
			for _, d_value := range u_value {
				for _, t_value := range d_value {
					clients = append(clients, t_value)
				}
			}
		}
	}

	this.Lock.Unlock()
	return clients
}

func (this *TmpClients) QueryForUuid(Number uint16, Uuid string) []*Client {
	this.Lock.Lock()

	clients := make([]*Client, 0)
	for key, n_value := range this.TmpClients {
		if key*Number > 0 {
			if u_value, ok := n_value[Uuid]; ok {
				for _, d_value := range u_value {
					for _, t_value := range d_value {
						clients = append(clients, t_value)
					}
				}
			}
		}
	}

	this.Lock.Unlock()
	return clients
}

func (this *TmpClients) QueryForDevice(Number uint16, Uuid string, Device string) []*Client {
	this.Lock.Lock()

	clients := make([]*Client, 0)
	if d_value, ok := this.TmpClients[Number][Uuid][Device]; ok {
		for _, t_value := range d_value {
			clients = append(clients, t_value)
		}
	}

	this.Lock.Unlock()
	return clients
}
