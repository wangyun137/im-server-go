package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const QCHEADBASE_LEN = 12

type QCHeadBase struct {
	Len     int64
	Version uint8
	Type    uint8
	Number  int16
}

func (base *QCHeadBase) Encode() (msgBytes []byte, err error) {
	msgBytes = make([]byte, QCHEADBASE_LEN)
	index := 0
	binary.BigEndian.PutUint64(msgBytes[index:index+8], uint64(base.Len))
	index += 8
	msgBytes[index] = byte(base.Version)
	index += 1
	msgBytes[index] = byte(base.Type)
	index += 1
	binary.BigEndian.PutUint16(msgBytes[index:index+2], uint16(base.Number))
	index += 2
	return msgBytes, err
}

func (base *QCHeadBase) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < QCHEADBASE_LEN {
		err = errors.New("QCHeadBase Decode:the argument can't be shorter 12 bytes")
		return err
	}
	msgLen := int64(len(msgBytes))
	index := 0
	base.Len = int64(binary.BigEndian.Uint64(msgBytes[index : index+8]))
	if msgLen < base.Len {
		err = errors.New("QCHeadBase Decode:the []byte length is shorter than expected")
	}
	index += 8
	base.Version = uint8(msgBytes[index])
	index += 1
	base.Type = int8(msgBytes[index])
	index += 1
	base.Number = int16(binary.BigEndian.Uint16(msgBytes[index : index+2]))
	index += 2
	return err
}
