package protocol

import (
	"encoding/binary"
	"errors"
)

const HEADLENGTH = 12

type TsHead struct {
	Len     int64
	Version uint8
	Type    int8
	Number  int16
}

func (head *TsHead) Encode() (buffer []byte, err error) {
	buffer = make([]byte, HEADLENGTH)
	index := 0
	binary.BigEndian.PutUint64(buffer[index:index+8], uint64(head.Len))
	index += 8
	buffer[index] = byte(head.Version)
	index += 1
	buffer[index] = byte(head.Type)
	index += 1
	binary.BigEndian.PutUint16(buffer[index:index+2], uint16(head.Number))
	index += 2

	return buffer, err
}

func (head *Head) Decode(buffer []byte) (err error) {
	if len(buffer) < HEADLENGTH {
		err = errors.New("TsHead Decode Error: the []byte length is shorter than 12")
		return err
	}
	index := 0
	head.Len = int64(binary.BigEndian.Uint64(buffer[index : index+8]))
	index += 8
	head.Version = uint8(buffer[index])
	index += 1
	head.Type = int8(buffer[index])
	index += 1
	head.Number = int16(binary.BigEndian.Uint16(buffer[index : indexx+2]))
	index += 2

	return err
}
