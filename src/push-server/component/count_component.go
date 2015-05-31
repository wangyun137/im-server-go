package component

import (
	"fmt"
	"sync"
	"time"
)

type QueueCount struct {
	Number int
	sync.RWMutex
}

func (this *QueueCount) Add() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.Number++
	return this.Number
}
func (this *QueueCount) Dec() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.Number--
	return this.Number
}
func (this *QueueCount) Print(queueName string) {
	this.RWMutex.RLock()
	this.RWMutex.RUnlock()
	fmt.Println("\n", time.Now().String(), queueName+" total number: ", this.Number, "\n")
}

var dispQueueCount QueueCount

func GetDispTotalQueueCount() *QueueCount {
	return &dispQueueCount
}

var dispNormalQueueCount QueueCount

func GetDispNormalQueueCount() *QueueCount {
	return &dispNormalQueueCount
}

var dispEmergentQueueCount QueueCount

func GetDispEmergentQueueCount() *QueueCount {
	return &dispEmergentQueueCount
}

var dispOutQueueCount QueueCount

func GetDispOutTotalQueueCount() *QueueCount {
	return &dispOutQueueCount
}

var dispOutNormalQueueCount QueueCount

func GetDispOutNormalQueueCount() *QueueCount {
	return &dispOutNormalQueueCount
}

var dispOutEmergentQueueCount QueueCount

func GetDispOutEmergentQueueCount() *QueueCount {
	return &dispOutEmergentQueueCount
}

var queryQueueCount QueueCount

func GetQueryQueueCount() *QueueCount {
	return &queryQueueCount
}

var queryRespQueueCount QueueCount

func GetQueryRespQueueCount() *QueueCount {
	return &queryRespQueueCount
}

var registerQueueCount QueueCount

func GetRegisterQueueCount() *QueueCount {
	return &registerQueueCount
}
