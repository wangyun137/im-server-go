package model

import (
	"errors"
	"net"
	"sync"
)

type Relation struct {
	FromUuid  string
	DestUuid  string
	FromToken string
	DestToken string
	Status    int
	TimeOut   int64
	FromConn  *net.Conn
	DestConn  *net.Conn
}

type Relations struct {
	Relation map[string]*Relation
	Lock     sync.Mutex
}

func NewRelations() *Relations {
	return &Relations{Relation: make(map[string]*Relation)}
}

func (this *Relations) Add(relation *Relation) (err error) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if _, ok := this.Relation[relation.FromUuid+relation.DestUuid]; ok {
		err = errors.New("Relations Add Error: Exists " + relation.FromUuid + relation.DestUuid)
		return err
	}
	this.Relation[relation.FromUuid+relation.DestUuid] = relation
	return nil
}

func (this *Relations) AddFrom(FromUuid, DestUuid, FromToken string, FromConn *net.Conn) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if k, ok := this.Relation[FromUuid+DestUuid]; ok {
		k.FromUuid = FromUuid
		k.FromToken = FromToken
		k.FromConn = FromConn
		this.Relation[FromUuid+DestUuid] = k
	} else {
		this.Relation[FromUuid+DestUuid] = &Relation{FromUuid: FromUuid, FromToken: FromToken, FromConn: FromConn}
	}
}

func (this *Relations) AddDest(FromUuid, DestUuid, DestToken string, DestConn *net.Conn) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if k, ok := this.Relation[DestUuid+FromUuid]; ok {
		k.DestUuid = FromUuid
		k.DestToken = DestToken
		k.DestConn = DestConn
		this.Relation[DestUuid+FromUuid] = k
	} else {
		this.Relation[DestUuid+FromUuid] = &Relation{DestUuid: FromUuid, DestToken: DestToken, DestConn: DestConn}
	}
}

func (this *Relations) UpdateStatus(uuid string, status int) (err error) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if k, ok := this.Relation[uuid]; ok {
		k.Status = status
		this.Relation[uuid] = k
		return nil
	}
	err = errors.New("Relations UpdateStatus Error: Not exists " + uuid)
	return err
}

func (this *Relations) UpdateTimeOut(uuid string, timeout int64) (err error) {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	if k, ok := this.Relation[uuid]; ok {
		k.TimeOut = timeout
		this.Relation[uuid] = k
		return nil
	}
	err = errors.New("Relations UpdateTimeOut Error: Not Exist " + uuid)
	return err
}

func (this *Relations) Delete(uuid string) (err error) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if _, ok := this.Relation[uuid]; ok {
		delete(this.Relation, uuid)
		return nil
	}
	err = errors.New("Relations Delete Error : Not Exist " + uuid)
	return err
}

func (this *Relations) Query(uuid string) (relation *Relation, err error) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if k, ok := this.Relation[uuid]; ok {
		return k, nil
	}
	err = errors.New("not exists information")
	return nil, err
}
