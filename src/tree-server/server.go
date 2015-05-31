package main

import (
	//	"errors"
	"connect-server/utils/config"
	"fmt"
	"runtime"
	"time"
	"tree-server/component"
	"tree-server/statistics"
)

type server struct {
	Conf    config.Config
	Type    string
	RunMode string
}

func main() {
	serv := server{}
	err := serv.ReadConf("conf/stree.conf")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	GoMaxProcs, _ := serv.Conf.GetInt("GOMAXPROCS")

	runtime.GOMAXPROCS(GoMaxProcs)

	err = serv.Init()
	if err != nil {
		fmt.Println(err)
		return
	}

	serv.Start()

	go func() {
		for {
			component.GetSTree().PrintUserMapCount()
			component.GetSTree().PrintSelfInfo()
			statistics.GetToDaughterQueueCount().Print("ToDaughterQueue")
			statistics.GetToParentQueueCount().Print("ToParentQueue")
			statistics.GetForwardCount().Print()
			time.Sleep(5 * time.Second)
		}
	}()
	sig := make(chan bool, 0)
	<-sig
}

//从stree.conf中读取配置属性
func (this *server) ReadConf(confPath string) error {
	this.Conf = *config.NewConfig()
	err := this.Conf.Read(confPath)
	if err != nil {
		return err
	}
	return nil
}

func (this *server) Init() error {
	//Type分为Root和Branch,如果为Branch，MasterIP、SlaveIP，ParentPort是必须的
	tp, err := this.Conf.GetString("Type")
	if err != nil {
		return err
	}
	mod, err := this.Conf.GetString("RunMode")
	if err != nil {
		return err
	}
	this.Type = tp
	this.RunMode = mod

	if this.Type == "Root" {
		err = this.InitCacheComponent()
		if err != nil {
			return err
		}
		err = this.InitClientHandler()
		if err != nil {
			return err
		}
	} else {
		err = this.InitParentHandler()
		if err != nil {
			return err
		}
	}
	err = this.InitDaughterHandler()
	if err != nil {
		return err
	}
	err = this.InitSTree()
	if err != nil {
		return err
	}
	return nil
}

func (this *server) InitDaughterHandler() error {
	//IP:10.0.0.11
	ip, err := this.Conf.GetString("IP")
	if err != nil {
		return err
	}
	//ServicePort:11001
	servicePort, err := this.Conf.GetInt("ServicePort")
	if err != nil {
		return err
	}
	//daughterQueueLen:1000
	queueLen, err := this.Conf.GetInt("daughterQueueLen")
	if err != nil {
		return err
	}
	fmt.Println("daughterQueueLen = ", queueLen)
	dh, err := component.NewDaughterHandler(queueLen)
	if err != nil {
		return err
	}
	dh.Init()
	dh.IP = ip
	dh.ServicePort = uint16(servicePort)
	fmt.Println("InitDaughterHandler() OK")
	return nil
}

//Type为Root时会调用
func (this *server) InitClientHandler() error {
	//IP:10.0.0.11
	ip, err := this.Conf.GetString("IP")
	if err != nil {
		return err
	}
	//ClientPort:11000
	port, err := this.Conf.GetInt("ClientPort")
	if err != nil {
		return err
	}
	//clientQueueLen:1000
	cLen, err := this.Conf.GetInt("clientQueueLen")
	if err != nil {
		return err
	}
	//allocRespQueueLen:1000
	aLen, err := this.Conf.GetInt("allocRespQueueLen")
	if err != nil {
		return err
	}

	clientHandler, err := component.NewClientHandler(cLen, aLen)
	if err != nil {
		return err
	}
	clientHandler.IP = ip
	clientHandler.ClientPort = uint16(port)
	clientHandler.Init()
	fmt.Println("InitClientHandler() OK")
	return nil
}

//Type为Root时会调用
func (this *server) InitCacheComponent() error {
	//CacheIP:10.0.0.11
	ip, err := this.Conf.GetString("CacheIP")
	if err != nil {
		return err
	}
	//CachePort:11003
	port, err := this.Conf.GetInt("CachePort")
	if err != nil {
		return err
	}
	//CacheQueueLen:1000
	qLen, err := this.Conf.GetInt("CacheQueueLen")
	if err != nil {
		return err
	}
	//respQueueLen:1000
	respLen, err := this.Conf.GetInt("respQueueLen")
	if err != nil {
		return err
	}

	cache, err := component.NewCacheComponent(qLen, respLen)
	if err != nil {
		return err
	}
	cache.CacheIP = ip
	cache.CachePort = uint16(port)
	cache.Init()
	fmt.Println("InitCacheComponent() OK")
	return nil
}

//Type为Branch时会调用
func (this *server) InitParentHandler() error {
	//MasterIP:10.0.0.11
	mip, err := this.Conf.GetString("MasterIP")
	if err != nil {
		return err
	}
	//SlaveIP:10.0.0.11
	sip, err := this.Conf.GetString("SlaveIP")
	if err != nil {
		return err
	}
	//ParentPort:11002
	port, err := this.Conf.GetInt("ParentPort")
	if err != nil {
		return err
	}
	//PushServerPort:9005
	pushPort, err := this.Conf.GetInt("PushServerPort")
	if err != nil {
		return err
	}
	//sendQueueLen:1000
	qLen, err := this.Conf.GetInt("sendQueueLen")
	if err != nil {
		return err
	}

	pHandler, err := component.NewParentHandler(qLen)
	if err != nil {
		return err
	}
	pHandler.MasterIP = mip
	pHandler.SlaveIP = sip
	pHandler.ParentPort = uint16(port)
	pHandler.PushServerPort = uint16(pushPort)

	pHandler.Init()
	fmt.Println("InitParentHandler() OK")
	return nil
}

func (this *server) InitSTree() error {
	//RunMode = master
	runMode, err := this.Conf.GetString("RunMode")
	if err != nil {
		return err
	}
	//Type = Branch
	tp, err := this.Conf.GetString("Type")
	if err != nil {
		return err
	}
	//Period = 120
	seconds, err := this.Conf.GetInt("Peroid")
	if err != nil {
		return err
	}

	stree, err := component.NewSTree(seconds)
	if err != nil {
		return err
	}
	stree.RunMode = runMode
	stree.Type = tp
	stree.Seconds = seconds

	stree.Init()
	fmt.Println("InitSTree() OK")
	return nil
}

func (this *server) Start() {
	if this.Type == "Root" {
		component.GetCacheComponent().Start()
		component.GetClientHandler().Start()
	} else {
		component.GetParentHandler().Start()
	}
	component.GetDaughterHandler().Start()
	component.GetSTree().Start()
}
