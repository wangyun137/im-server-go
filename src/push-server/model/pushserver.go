package model

import (
	"connect-server/utils/config"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"push-server/buffpool"
	"push-server/component"
	"push-server/connector"
	"push-server/utils"
	snode_component "snode-server/component"
)

type PushServer struct {
	GoMacProcs       int
	PushServerIP     string
	PushServerPort   uint16
	DispQueueLen     int
	ConnQueueLen     int
	EmergentQueueLen int
	Level            int
	ArrLen           int
	MaxConnNumber    uint32
	CacheQueueLen    int
	CacheResponseLen int
	PushID           string
}

func (this *PushServer) ReadConf(confPath string) error {
	config := config.NewConfig()
	config.Read(confPath)
	NCPU, err := config.GetInt("GOMAXPROCS")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.GoMaxProcs = NCPU

	pushIP, err := config.GetString("PushServerIP")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.PushServIP = pushIP

	pushPort, err := config.GetInt("PushServerPort")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.PushServPort = uint16(pushPort)

	dispQueueLen, err := config.GetInt("DispatcherQueueLen")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.DispQueueLen = dispQueueLen
	emergeQueueLen, err := config.GetInt("EmergentQueueLen")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.EmergentQueueLen = emergeQueueLen
	level, err := config.GetInt("Level")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.Level = level

	connLen, err := config.GetInt("ConnQueueLen")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.ConnQueueLen = connLen

	arrLen, err := config.GetInt("ArrLen")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.ArrLen = arrLen

	maxConn, err := config.GetInt("MaxConnNumber")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.MaxConnNumber = uint32(maxConn)

	cacheLen, err := config.GetInt("CacheQueueLen")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.CacheQueueLen = cacheLen
	cacheRespLen, err := config.GetInt("CacheResponseLen")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.CacheResponseLen = cacheRespLen

	pushID, err := config.GetString("PushID")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.PushID = pushID
	return nil
}

func (this *PushServer) Init() (*connector.Connecter, error) {
	buffPool := buffpool.GetBuffPool()
	if buffPool == nil {
		return nil, errors.New("Error when init BuffPool")
	}
	fmt.Println("BuffPool init ok")
	//初始化一个长度为100的全局QueryInfo类型通道
	queryComponent := component.NewQueryComponent()
	if queryComponent == nil {
		return nil, errors.New("Error when create QueryComponent")
	}
	queryComponent.Init()
	fmt.Println("QueryComponent init ok")
	//DispQueueLen = 1000，EmergentQueueLen = 50，Level = 1
	//用上述三个值初始化一个PriorityQueue，DispatcherComponent只嵌入了PriorityQueue
	dispatcherComponent, err := component.NewDispatcherComponent(this.DispQueueLen, this.EmergentQueueLen, this.Level)
	if dispatcherComponent == nil || err != nil {
		return nil, errors.New("Error when create DispatcherComponent")
	}
	dispatcherComponent.Init()
	fmt.Println("DispatcherComponent init ok")
	//DispQueueLen = 1000，EmergentQueueLen = 50，Level = 1
	//用上述三个值初始化一个PriorityQueue，DispatcherOutComponent只嵌入了PiorityQueue
	dispatcherOutComponent, err := component.NewDispatcherOutComponent(this.DispQueueLen, this.EmergentQueueLen, this.Level)
	if dispatcherOutComponent == nil || err != nil {
		return nil, errors.New("Error when create DispatcherOutComponent")
	}
	dispatcherOutComponent.Init()
	fmt.Println("DispatcherOutComponent init OK")
	//ResponseComponent结构体无字段，初始化一个Message类型的通道，长度为100
	responseComponent := component.NewResponseComponent()
	if responseComponent == nil {
		return nil, errors.New("Error when create ResponseComponent")
	}
	responseComponent.Init()
	fmt.Println("ResponseComponent init ok")

	//用10.0.0.7:7005生成一个RegisterComponent，并将queueLen和arrLen设为1000,1
	ip := binary.BigEndian.Uint32(net.ParseIP(this.PushServIP).To4())
	registerComponent := component.NewRegisterComponent(this.ConnQueueLen, this.ArrLen, ip, this.PushServPort, this.PushID)
	if registerComponent == nil {
		return nil, errors.New("Error when create RegisterComponent")
	}
	registerComponent.Init()
	fmt.Println("RegisterComponent init ok")

	//run with snode-stree-scache
	//这里未启用Snode
	if utils.GetRunWithSnode().Flag {
		node := snode_component.GetNodeComponent()
		node.Init()
		node.SetMaxConnNumber(this.MaxConnNumber)
		fmt.Println("snode.Init() OK")

		component.GetSTreeMsgHandler().Init()
		fmt.Println("SNode-STreeMsgHandler() Init OK")

		cache := component.NewCacheComponent2(this.CacheQueueLen, this.CacheResponseLen)
		cache.Init2()
		fmt.Println("CacheComponent2.Init2() OK")
	} else {
		//run without snode-stree-scache
		//生成一个CacheComponent的单例，并初始化其中的Cache为长度为0的映射
		cacheComponent := component.NewCacheComponent()
		cacheComponent.Init()
		fmt.Println("CacheComponent.Init() OK")
	}
	//用10.0.0.7:7005初始化一个Connector
	connect := connector.NewConnecter()
	if connect == nil {
		fmt.Println("Error when create connecter")
	}
	connect.Init(this.PushServIP, int(this.PushServPort))
	fmt.Println("connecter init ok")

	return conner, nil
}

func (this *PushServer) Start(connect *connector.Connecter) error {

	queryComponent := component.GetQueryComponent()
	if queryComponent == nil {
		return errors.New("Can not get QueryComponent sigleton")
	}
	queryComponent.Start()
	fmt.Println("QueryComponent started")

	dispatcherComponent := component.GetDispatcherComponent()
	if dispatcherComponent == nil {
		return errors.New("Can not get Dispatcher sigleton")
	}
	dispatcherComponent.Start()
	fmt.Println("DispatcherComponent started")

	dispatcherOutComponent := component.GetDispatcherOutComponent()
	if dispOut == nil {
		return errors.New("Can not get DispatcherOut sigleton")
	}
	dispatcherOutComponent.Start()
	fmt.Println("DispatcherOutComponent started")

	responseComponent := component.GetResponseComponent()
	if responseComponent == nil {
		return errors.New("Can not get ResponseComponent sigleton")
	}
	responseComponent.Start()
	fmt.Println("ResponseComponent started")

	if utils.GetRunWithSnode().Flag {
		node := snode_component.GetNodeComponent()
		node.Start()
		fmt.Println("NodeComponent started")

		component.GetSTreeMsgHandler().Start()
		fmt.Println("SNode-STreeMsgHandler Startd")

		//run with snode-stree-scache
		cache := component.GetCacheComponent2()
		if cache == nil {
			return errors.New("Can not get CacheComponent2 sigleton")
		}
		cache.Start2()
		fmt.Println("CacheComponent2 started")
	} else {
		//run without snode-stree-scache
		cacheComponent := component.GetCacheComponent()
		if cacheComponent == nil {
			return errors.New("Can not get CacheComponent sigleton")
		}
		cacheComponent.Start()
		fmt.Println("CacheComponent started")
	}

	registerComponent := component.GetRegisterComponent()
	if registerComponent == nil {
		return errors.New("Can not get RegisterComponent sigleton")
	}
	registerComponent.Start()
	fmt.Println("RegisterComponent started")

	err := connect.Start()
	if err != nil {
		return errors.New("Can not start connecter, " + err.Error())
	}
	fmt.Println("connecter Started")
	return nil
}
