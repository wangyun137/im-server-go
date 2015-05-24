package model

import (
	"errors"
	"push-server/protocol"
	"sync"
)

type MessagePacket struct {
	Message *protocl.Message
	Flag    bool
}

/*
*impletation of Priority Queue. When Priority Queue is empty, read from it will be blocked
*Using NPushBack() and EPushBack() to insert events, (only) using Pop() to get a event
*the Mutex is used to protect the QueueSize: NormalQueueSize and EmergentQueueSize
*
*Since the QueueSize is not strictly with the elements number in channel, the QueueSize may be < 0 some times,
*But finally, it will be zero(Evently, the QueueSize is sync to the elements number of channel
 */

type PriorityQueue struct {
	NormalQueue       chan *MsgPacket
	NormalQueueLen    int
	NormalQueueSize   int
	EmergentQueue     chan *MsgPacket
	EmergentQueueLen  int
	EmergentQueueSize int
	Level             int
	ECount            int
	sync.Mutex
}

func NewPriorityQueue(nQueueLen, eQueueLen, level int) (*PriorityQueue, error) {
	if nQueueLen <= 0 || eQueueLen <= 0 {
		err := errors.New("NewPriorityQueue() failed, queue length must be positive")
		return nil, err
	}

	queue := &PriorityQueue{
		NormalQueueLen:   nQueueLen,
		EmergentQueueLen: eQueueLen,
		Level:            level,
		NormalQueue:      make(chan *MsgPacket, nQueueLen),
		EmergentQueue:    make(chan *MsgPacket, eQueueLen),
	}
	return queue, nil
}

func (this *PriorityQueue) NPushBack(msg *protocol.Message, flag bool) error {
	if msg == nil {
		err := errors.New("Priority.NPushBack() failed, got a nil pointer")
		return err
	}
	packet := &MsgPacket{
		Message: msg,
		Flag:    flag,
	}
	this.NormalQueue <- packet

	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.NormalQueueSize++
	return nil
}

func (this *PriorityQueue) EPushBack(msg *protocol.Message, flag bool) error {
	if msg == nil {
		err := errors.New("Priority.EPushBack() failed, got a nil pointer")
		return err
	}
	packet := &MsgPacket{
		Message: msg,
		Flag:    flag,
	}
	this.EmergentQueue <- packet
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.EmergentQueueSize++
	return nil
}

func (this *PriorityQueue) nPop() *MessagePacket {
	if this.NormalQueueSize <= 0 {
		return nil
	}
	packet := <-this.NormalQueue
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.NormalQueueSize--
	return packet
}

func (this *PriorityQueue) ePop() *MessagePacket {
	if this.EmergentQueueSize <= 0 {
		return nil
	}
	packet := <-this.EmergentQueue
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.EmergentQueueSize--
	return packet
}

func (this *PriorityQueue) Pop() *MessagePacket {
	packet := this.PrioPop()

	if packet == nil {
		select {
		case packet = <-this.EmergentQueue:
			this.Mutex.Lock()
			this.EmergentQueueSize--
			this.Mutex.Unlock()
		case packet = <-this.NormalQueue:
			this.Mutex.Lock()
			this.NormalQueueSize--
			this.Mutex.Unlock()
		}
	}
	return packet
}

func (this *PriorityQueue) PrioPop() *MessagePacket {
	var packet *MessagePacket
	if this.Level <= 0 || this.Level > 3 {
		//default case, handle all emergent events then handle normal events
		packet = this.ePop()
		if packet == nil {
			packet = this.nPop()
		}
	} else {
		//only in this function, the ECount will be modified, so it is not necessary to protect it by a Lock
		if this.ECount < this.Level {
			packet = this.ePop()
			this.ECount++
			if packet == nil {
				packet = this.nPop()
			}
		} else {
			this.ECount = 0
			packet = this.nPop()
			if packet == nil {
				packet = this.ePop()
			}
		}
	}
	return packet
}
