package statistics

import (
	"fmt"
	"time"
	"sync"
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
	fmt.Println("\n", time.Now().String(),  queueName + " total elements: ", this.Number, "\n")
}

type ForwardCount struct {
    FromDaughter int 
    ToDaughter int 
    sync.RWMutex
}

var forwardCount  ForwardCount
func GetForwardCount() *ForwardCount {
    return &forwardCount
}

func (this *ForwardCount) AddFromDaughter() int {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    this.FromDaughter++
    return this.FromDaughter
}

func (this *ForwardCount) AddToDaughter() int {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    this.ToDaughter++
    return this.ToDaughter
}

func (this *ForwardCount) Print() {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    fmt.Println("\n", "Forward Packets from daughter: ", this.FromDaughter, " ; Forward Packets to Daughter: ", this.ToDaughter, "\n")
}



var dispQueueCount QueueCount
func GetDispQueueCount() *QueueCount {
	return &dispQueueCount
}
var dispOutQueueCount QueueCount
func GetDispOutQueueCount() *QueueCount {
	return &dispOutQueueCount
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

//for STree
var toDaughterQueueCount QueueCount
func GetToDaughterQueueCount() *QueueCount {
	return &toDaughterQueueCount
}
var toParentQueueCount QueueCount
func GetToParentQueueCount() *QueueCount {
	return &toParentQueueCount
}
var clientQueueCount QueueCount
func GetClientQueueCount() *QueueCount {
	return &clientQueueCount
}
var allocRespQueueCount QueueCount
func GetAllocRespQueueCount() *QueueCount {
	return &allocRespQueueCount
}
var cacheQueueCount QueueCount
func GetCacheQueueCount() *QueueCount {
	return &cacheQueueCount
}
var cacheRespQueueCount QueueCount
func GetCacheRespQueueCount() *QueueCount {
	return &cacheRespQueueCount
}
//for STree End

type rawAccountStatistics struct {
	Request int
	Response int
	sync.RWMutex
}

var rawStat rawAccountStatistics

func GetRawAccountStatistics() *rawAccountStatistics {
	return &rawStat
}

func (this *rawAccountStatistics) AddRequest() int {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    this.Request++
    return this.Request
}

func (this *rawAccountStatistics) AddResponse() int {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    this.Response++
    return this.Response
}

func (this *rawAccountStatistics) Print() {
    this.RWMutex.RLock()
    defer this.RWMutex.RUnlock()
    fmt.Println("\n", time.Now().String(), " Total Raw Request: ", this.Request, "; Total Raw Response: ", this.Response, "\n")
    return
}

type AccountStatistics struct {
    Request int
    Response int
    sync.RWMutex
}

var accStat AccountStatistics

func GetAccountStatistics() *AccountStatistics {
    return &accStat
}

func (this *AccountStatistics) AddRequest() int {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    this.Request++
    return this.Request
}

func (this *AccountStatistics) AddResponse() int {
    this.RWMutex.Lock()
    defer this.RWMutex.Unlock()
    this.Response++
    return this.Response
}

func (this *AccountStatistics) Print() {
    this.RWMutex.RLock()
    defer this.RWMutex.RUnlock()
    fmt.Println("\n", time.Now().String(), " Total Request: ", this.Request, "; Total Response: ", this.Response, "\n")
    return
}


type ClientAddFailed struct {
	TotalFailed int
	InfoArr []string
	sync.RWMutex
}

type ConnStatistics struct {
	TotalConn int
	sync.RWMutex
}

type LoginStatistics struct {
	TotalLogin int
	sync.RWMutex
}

type LoginFailed struct {
	TotalFailed int
	InfoArr []string
	sync.RWMutex
}

type CallbackStatistics struct {
	TotalCallback int
	FailedCallback int
	SucceedCallback int
	FailedInfo []string
	sync.RWMutex
}
var callbackStatistics CallbackStatistics

func GetCallbackStatistics() *CallbackStatistics {
	cb := &callbackStatistics
	if cb.FailedInfo == nil {
		cb.FailedInfo = make([]string, 0, 100)
	}
	return cb
}

func (this *CallbackStatistics) Add() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalCallback++
	return this.TotalCallback
}

func (this *CallbackStatistics) AddSucceed() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.SucceedCallback++
	return this.SucceedCallback
}

func (this *CallbackStatistics) AddFailed() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.FailedCallback++
	return this.FailedCallback
}

func (this *CallbackStatistics) AddFailedInfo(info string) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.FailedInfo = append(this.FailedInfo, info)
	return
}

func (this *CallbackStatistics) Print() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	fmt.Println("\n", time.Now().String(), " TotalCallback: ", this.TotalCallback, "; FailedCallback: ", this.FailedCallback, "; SucceedCallback: ", this.SucceedCallback, "\n")
	return
}

func (this *CallbackStatistics) PrintInfo() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	for i, v := range this.FailedInfo {
		fmt.Println(i, " : ", v)
	}
}



var clientAddFailed ClientAddFailed

func GetClientAddFailed() *ClientAddFailed {
	caf := &clientAddFailed
	if caf.InfoArr == nil {
		caf.InfoArr = make([]string, 0, 100)
	}
	return caf
}

func (this *ClientAddFailed) Add() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalFailed++
	return this.TotalFailed
}

func (this *ClientAddFailed) AddInfo(info string) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.InfoArr = append(this.InfoArr, info)
	return
}

func (this *ClientAddFailed) Print() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	fmt.Println("\n", time.Now().String(), " Total ClientAddFailed : ", this.TotalFailed, "\n")
	return
}

func (this *ClientAddFailed) PrintInfo() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	for i, v := range this.InfoArr {
		fmt.Println(i, " : ", v)
	}
}

var loginFailed LoginFailed

func GetLoginFailed() *LoginFailed {
	lf := &loginFailed
	if loginFailed.InfoArr == nil {
		lf.InfoArr = make([]string, 0, 100)
	}
	return lf
}

func (this *LoginFailed) AddInfo(info string) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.InfoArr = append(this.InfoArr, info)
	return
}

func (this *LoginFailed) Add() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalFailed++
	return this.TotalFailed
}

func (this *LoginFailed) Print() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	fmt.Println("\n", time.Now().String(), " Total Failed Login : ", this.TotalFailed, "\n")
}

func (this *LoginFailed) PrintInfo() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	for i, v := range this.InfoArr {
		fmt.Println(i, " : ", v)
	}
}

var connStatistics ConnStatistics

func GetConnStatistics() *ConnStatistics {
	return &connStatistics
}

var loginStatistics LoginStatistics

func GetLoginStatistics() *LoginStatistics {
	return &loginStatistics
}

func (this *ConnStatistics) Get() int {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	return this.TotalConn
}

func (this *ConnStatistics) Print() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	fmt.Println("\n", time.Now().String(), " Total Connections: ", this.TotalConn, "\n")
}

func (this *ConnStatistics) Add() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalConn++
	return this.TotalConn
}

func (this *ConnStatistics) Dec() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalConn--
	return this.TotalConn
}

func (this LoginStatistics) Get() int {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	return this.TotalLogin
}

func (this *LoginStatistics) Add() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalLogin++
	return this.TotalLogin
}

func (this *LoginStatistics) Dec() int {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	this.TotalLogin--
	return this.TotalLogin
}

func (this *LoginStatistics) Print() {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	fmt.Println("\n", time.Now().String(), " Total Login Connections: ", this.TotalLogin, "\n")
}
