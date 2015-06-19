package protocol

import (
	"encoding/binary"
	"errors"
)

const HEADBASE_LEN = 28

type HeadBase struct {
	Len         int64
	Version     uint8
	Type        uint8
	Number      int16
	SendTime    int64
	DiscardTiem int64
}

func (head *HeadBase) Encode() (msgBytes []byte) {
	msgBytes = make([]byte, HEADBASE_LEN)
	index := 0
	binary.BigEndian.PutUint64(msgBytes[index:index+8], uint64(head.Len)
	index += 8 
	msgBytes[index] = byte(head.Version)
	index += 1
	msgBytes[index] = byte(head.Type)
	index += 1
	binary.BigEndian.PutUint16(msgBytes[index:index+2], uint16(head.Number))
	index += 2
	binary.BigEndian.PutUint64(msgBytes[index:index+8], uint64(head.SendTime))
	index += 8
	binary.BigEndian.PutUnt64(msgBytes[index:index+8], uint64(head.DiscardTiem))
	index += 8

	return msgBytes
}


func (head *HeadBase) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < HEADBASE_LEN {
		err = errors.New("Connect-Server HeadBase Error: the argument can't be shorter than 28 bytes")
		return err
	}
	msgLen := int64(len(msgBytes))
	index := 0
	head.Len = int64(binary.BigEndian.Uint64(msgBytes[index:index+8]))
	if msgLen < head.Len {
		err = errors.New("Connect-Server HeadBase Error:[]byte length is shorter than expected")
		return err
	}
	index += 8
	head.Version = uint8(msgBytes[index])
	index += 1
	head.Type = uint8(msgBytes[index])
	index += 1
	head.Number = int16(binary.BigEndian.Uint16(msgBytes[index:index+2]))
	index += 2
	head.SendTime = int64(binary.BigEndian.Uint64(msgBytes[index:index+8]))
	index += 8
	head.DiscardTiem = int64(binary.BigEndian.Uint64(msgBytes[index:index+8]))
	index += 8
	return err
}



