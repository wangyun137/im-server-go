package buffpool

import (
	"fmt"
	"net"
	"push-server/model"
	"time"
)

type BuffPool struct {
	Services   *model.Services
	TmpClients *model.TmpClients
	Clients    *model.Clients
}

func init() {
	pool = GetBuffPool()
}

func NewBuffPool() *BuffPool {
	return &BuffPool{
		Services:   model.NewServices(),
		TmpClients: model.NewTmpClients(),
		Clients:    model.NewClients(),
	}
}

var pool *BuffPool

func GetBuffPool() *BuffPool {
	if pool == nil {
		pool = NewBuffPool()
	}
	return pool
}

//返回token
func (this *BuffPool) AddTmpClient(Number uint16, Uuid string, Device string, client *model.Client) uint8 {
	return this.TmpClients.Add(Number, Uuid, Device, client)
}

func (this *BuffPool) AddTmpClient2(Number uint16, Uuid string, Device string, DeviceID uint32, oldToken uint8, client *model.Client) uint8 {
	this.Clients.Remove(Number, Uuid, DeviceID, oldToken)
	return this.TmpClients.Add(Number, Uuid, Device, client)
}

func (this *BuffPool) MoveTmpClient(Number uint16, Uuid string, Device string, DeviceID uint32, oldToken uint8, client *model.Client) uint8 {
	this.TmpClients.Remove2(Number, Uuid, Device, oldToken)
	return this.Clients.Add(Number, Uuid, DeviceID, client)
}

func (this *BuffPool) RemoveTmpClient(Number uint16, Uuid string, Device string, Token uint8) error {
	return this.TmpClients.Remove(Number, Uuid, Device, Token)
}

func (this *BuffPool) DeleteTmpClient(Number uint16, Uuid string) error {
	return this.TmpClients.DeleteUuid(Number, Uuid)
}

func (this *BuffPool) DeleteTmpDevice(Number uint16, Uuid string, Device string) error {
	return this.TmpClients.DeleteDevice(Number, Uuid, Device)
}

func (this *BuffPool) GetTmpClient(Number uint16, Uuid string, Device string, Token uint8) *model.Client {
	return this.TmpClients.Query(Number, Uuid, Device, Token)
}

func (this *BuffPool) UpdateTmpClientStatus(Number uint16, Uuid string, Device string, Token uint8, Status bool) error {
	return this.TmpClients.UpdateStatus(Number, Uuid, Device, Token, Status)
}

func (this *BuffPool) QueryTmpForNumber(Number uint16) []*model.Client {
	return this.TmpClients.QueryForNumber(Number)
}

func (this *BuffPool) QueryTmpForUuid(Number uint16, Uuid string) []*model.Client {
	return this.TmpClients.QueryForUuid(Number, Uuid)
}

func (this *BuffPool) QueryTmpForDevice(Number uint16, Uuid string, Device string) []*model.Client {
	return this.TmpClients.QueryForDevice(Number, Uuid, Device)
}

func (this *BuffPool) AddClient(Number uint16, Uuid string, Device uint32, client *model.Client) uint8 {
	return this.Clients.Add(Number, Uuid, Device, client)
}

func (this *BuffPool) RemoveClient(Number uint16, Uuid string, Device uint32, Token uint8) error {
	return this.Clients.Remove(Number, Uuid, Device, Token)
}

func (this *BuffPool) DeleteClient(Number uint16, Uuid string) error {
	return this.Clients.DeleteUuid(Number, Uuid)
}

func (this *BuffPool) DeleteDevice(Number uint16, Uuid string, Device uint32) error {
	return this.Clients.DeleteDevice(Number, Uuid, Device)
}

func (this *BuffPool) GetClient(Number uint16, Uuid string, Device uint32, Token uint8) *model.Client {
	return this.Clients.Query(Number, Uuid, Device, Token)
}

func (this *BuffPool) UpdateClientStatus(Number uint16, Uuid string, Device uint32, Token uint8, Status bool) error {
	//	fmt.Println("\n", time.Now().String(), " UpdateClientStatus\n")
	return this.Clients.UpdateStatus(Number, Uuid, Device, Token, Status)
}

func (this *BuffPool) QueryForNumber(Number uint16) []*model.Client {
	return this.Clients.QueryForNumber(Number)
}

func (this *BuffPool) QueryForUuid(Number uint16, Uuid string) []*model.Client {
	return this.Clients.QueryForUuid(Number, Uuid)
}

func (this *BuffPool) QueryForDeviceId(Number uint16, Uuid string, Device uint32) []*model.Client {
	return this.Clients.QueryForDeviceId(Number, Uuid, Device)
}

func (this *BuffPool) AddService(Number uint16, Uuid string, Conn *net.Conn) error {
	return this.Services.Add(Number, Uuid, Conn)
}

func (this *BuffPool) DeleteService(Number uint16, Uuid string) error {
	return this.Services.Remove(Number, Uuid)
}

func (this *BuffPool) QueryService(Number uint16, Uuid string) *net.Conn {
	return this.Services.Query(Number, Uuid)
}

func (this *BuffPool) GetService(Number uint16) *net.Conn {
	return this.Services.Get(Number)
}

func (this *BuffPool) GetAllForNumber(Number uint16) model.ServiceUuids {
	return this.Services.GetAllForNumber(Number)
}

func (this *BuffPool) GetAllConn() []*net.Conn {
	return this.Services.GetAllConn()
}

func (this *BuffPool) GetAll() map[string]*net.Conn {
	return this.Services.GetAll()
}

func (this *BuffPool) GetAllServices() map[uint16][]string {
	return this.Services.GetAllServices()
}

func (this *BuffPool) PrintServiceConnections() {
	fmt.Println("BuffPool.ServiceConnections: ", this.Services.Services)
}

func (this *BuffPool) PrintClientConnections() {
	fmt.Println("BuffPool.ClientConnections: ", this.Clients.Clients)
}

func (this *BuffPool) PrintTmpClientConnections() {
	fmt.Println("BuffPool.TmpClientConnections: ", this.TmpClients.TmpClients)
}

func (this *BuffPool) PrintConnNumber() {
	iConn, iTmpConn := 0, 0
	this.Clients.Lock.RLock()
	for _, clientUuids := range this.Clients.Clients {
		for _, devIds := range clientUuids {
			for _, tokens := range devIds {
				for _, _ = range tokens {
					iConn++
				}
			}
		}
	}
	this.Clients.Lock.RUnlock()

	this.TmpClients.Lock.RLock()
	for _, tmpClientUuids := range this.TmpClients.TmpClients {
		for _, devNo := range tmpClientUuids {
			for _, tokens := range devNo {
				for _, _ = range tokens {
					iTmpConn++
				}
			}
		}
	}
	this.TmpClients.Lock.RUnlock()

	fmt.Println("\n", time.Now().String(), " Total Connections: ", iConn, " ; Total TmpConnections: ", iTmpConn, "\n")
}
