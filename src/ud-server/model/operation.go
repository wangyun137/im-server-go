package model

import (
	"connect-server/protocol"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	LEN_LENGTH  = 2
	BASE_LENGTH = LEN_LENGTH + TYPE_LENGTH
)

type Operation struct {
	Len  uint16
	Type uint8
	List []*List
}

func NewOperation(Type uint8, List ...*List) *Operation {
	return &Operation{
		Len:  uint16(BASE_LENGTH + len(List)*(UUID_LENGTH+DEVICEID_LENGTH)),
		Type: Type,
		List: List,
	}
}

type List struct {
	Uuid     string
	DeviceId uint32
}

func NewList(Uuid string, DeviceId uint32) *List {
	return &List{
		Uuid:     Uuid,
		DeviceId: DeviceId,
	}
}

func (this *Operation) Encode() ([]byte, error) {
	user_count := len(this.List)
	length := BASE_LENGTH + user_count*UUID_LENGTH + user_count*DEVICEID_LENGTH
	buff := make([]byte, length)
	index := 0

	binary.BigEndian.PutUint16(buff[index:index+LEN_LENGTH], this.Len)
	index += LEN_LENGTH

	buff[index] = this.Type
	index += TYPE_LENGTH

	for _, v := range this.List {
		user_uuid_buff, err := protocol.StringToUuid(v.Uuid)
		if err != nil {
			err = fmt.Errorf("Operation Encode Error: %v", err)
			return nil, err
		}

		copy(buff[index:index+UUID_LENGTH], user_uuid_buff[0:])
		index += UUID_LENGTH

		binary.BigEndian.PutUint32(buff[index:index+DEVICEID_LENGTH], v.DeviceId)
		index += DEVICEID_LENGTH
	}

	if index != length {
		return nil, errors.New("Operation Encode Error: index != length")
	}

	return buff, nil
}

func (this *Operation) Decode(buff []byte) error {
	buff_length := len(buff)
	user_count := (buff_length - BASE_LENGTH) / (UUID_LENGTH + DEVICEID_LENGTH)

	this.List = make([]*List, user_count)

	index := 0
	this.Len = binary.BigEndian.Uint16(buff[index : index+LEN_LENGTH])
	index += LEN_LENGTH

	this.Type = buff[index]
	index += TYPE_LENGTH

	for i, _ := range this.List {
		Uuid, err := protocol.UuidToString(buff[index : index+UUID_LENGTH])
		if err != nil {
			return fmt.Errorf("Operation Decode Error:%v", err)
		}
		index += UUID_LENGTH

		DeviceId := binary.BigEndian.Uint32(buff[index : index+DEVICEID_LENGTH])
		index += DEVICEID_LENGTH

		this.List[i] = NewList(Uuid, DeviceId)
	}

	if index != buff_length {
		return errors.New("Operation Decode Error: index != length")
	}

	return nil
}
