package model

import (
	"connect-server/protocol"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	VERSION_LENGTH  = 1
	TYPE_LENGTH     = 1
	UUID_LENGTH     = 16
	IP_LENGTH       = 4
	PORT_LENGTH     = 2
	DEVICEID_LENGTH = 4

	REQUEST_LENGTH = VERSION_LENGTH + TYPE_LENGTH + UUID_LENGTH*2 + DEVICEID_LENGTH*2
)

//Request至少42
type Request struct {
	Version      uint8
	Type         uint8
	FromUuid     string
	DestUuid     string
	FromDeviceId uint32
	DestDeviceId uint32
	Data         []byte
}

func (this *Request) Encode() (buff []byte, err error) {
	dataLength := len(this.Data)
	buff = make([]byte, REQUEST_LENGTH+dataLength)

	index := 0
	buff[index] = this.Version
	index += VERSION_LENGTH

	buff[index] = this.Type
	index += TYPE_LENGTH

	fromUuidBuff, err := protocol.StringToUuid(this.FromUuid)
	if err != nil {
		return nil, fmt.Errorf("Request Encode Error: %v", err)
	}
	copy(buff[index:index+UUID_LENGTH], fromUuidBuff)
	index += UUID_LENGTH

	destUuidBuff, err := protocol.StringToUuid(this.DestUuid)
	if err != nil {
		return nil, fmt.Errorf("Request Encode Error: %v", err)
	}
	copy(buff[index:index+UUID_LENGTH], destUuidBuff)
	index += UUID_LENGTH

	binary.BigEndian.PutUint32(buff[index:index+DEVICEID_LENGTH], this.FromDeviceId)
	index += DEVICEID_LENGTH

	binary.BigEndian.PutUint32(buff[index:index+DEVICEID_LENGTH], this.DestDeviceId)
	index += DEVICEID_LENGTH

	copy(buff[index:index+dataLength], this.Data[0:dataLength])
	index += dataLength

	if index != REQUEST_LENGTH+dataLength {
		return nil, errors.New("Request Encode Error:encode length is wrong ")
	}

	return buff, nil
}

func (this *Request) Decode(buff []byte) error {
	if len(buff) < REQUEST_LENGTH {
		return errors.New("Request Decode Error : the []byte length is shorter than 42")
	}
	index := 0
	dataLength := len(buff)

	this.Version = buff[index]
	index += VERSION_LENGTH

	this.Type = buff[index]
	index += TYPE_LENGTH

	var err error

	if this.FromUuid, err = protocol.UuidToString(buff[index : index+UUID_LENGTH]); err != nil {
		return fmt.Errorf("Request Decode Error: %v", err)
	}
	index += UUID_LENGTH

	if this.DestUuid, err = protocol.UuidToString(buff[index : index+UUID_LENGTH]); err != nil {
		return fmt.Errorf("Request Decode Error: %v", err)
	}
	index += UUID_LENGTH

	this.FromDeviceId = binary.BigEndian.Uint32(buff[index : index+DEVICEID_LENGTH])
	index += DEVICEID_LENGTH

	this.DestDeviceId = binary.BigEndian.Uint32(buff[index : index+DEVICEID_LENGTH])
	index += DEVICEID_LENGTH

	this.Data = buff[index:dataLength]
	index += dataLength - index
	if index != dataLength {
		return errors.New("Request Decode Error: dataLength is wrong")
	}

	return nil
}

type Register struct {
	Request
	IP   []byte
	Port uint16
}
