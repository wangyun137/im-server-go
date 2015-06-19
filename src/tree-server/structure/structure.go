package structure

import (
	"encoding/binary"
	"errors"
	"fmt"
	"tree-server/define"
	"tree-server/utils"
)

type Packet struct {
	Head
	Data []byte
}

type Head struct {
	Number  uint8
	Uuid    string
	DataLen uint16
}

func (this *Head) Encode() ([]byte, error) {
	//PACKET_HEAD_LEN为19
	buf := make([]byte, define.PACKET_HEAD_LEN)
	buf[0] = byte(this.Number)
	//uuid 16位
	uuid, err := utils.StringToUuid(this.Uuid)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	//UUID_BYTE_LEN为16
	copy(buf[1:1+define.UUID_BYTE_LEN], uuid)
	binary.BigEndian.PutUint16(buf[1+define.UUID_BYTE_LEN:3+define.UUID_BYTE_LEN], this.DataLen)
	return buf, nil
}

func (this *Head) Decode(buf []byte) error {
	if len(buf) < define.PACKET_HEAD_LEN { //19
		err := errors.New("structrue: Head.Decode() failed, buf is too short")
		fmt.Println(err.Error())
		return err
	}
	this.Number = uint8(buf[0])
	uuid, err := utils.UuidToString(buf[1 : 1+define.UUID_BYTE_LEN])
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.Uuid = uuid
	this.DataLen = binary.BigEndian.Uint16(buf[1+define.UUID_BYTE_LEN : 3+define.UUID_BYTE_LEN])
	return nil
}

func (this *Packet) Encode() ([]byte, error) {
	buf := make([]byte, define.PACKET_HEAD_LEN+len(this.Data))
	buf[0] = byte(this.Number)
	uuid, err := utils.StringToUuid(this.Uuid)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	copy(buf[1:1+define.UUID_BYTE_LEN], uuid)
	binary.BigEndian.PutUint16(buf[1+define.UUID_BYTE_LEN:3+define.UUID_BYTE_LEN], this.DataLen)
	copy(buf[3+define.UUID_BYTE_LEN:], this.Data)
	return buf, nil
}

func (this *Packet) Decode(buf []byte) error {
	if len(buf) < define.PACKET_HEAD_LEN {
		err := errors.New("structrue: Packet.Decode() failed, buf is too short")
		fmt.Println(err.Error())
		return err
	}
	this.Number = uint8(buf[0])
	uuid, err := utils.UuidToString(buf[1 : 1+define.UUID_BYTE_LEN])
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	this.Uuid = uuid
	this.DataLen = binary.BigEndian.Uint16(buf[1+define.UUID_BYTE_LEN : 3+define.UUID_BYTE_LEN])
	this.Data = buf[3+define.UUID_BYTE_LEN:]
	return nil
}
