package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const MESSAGE_LEN = 44

//At least 44 bytes
type Message struct {
	Version  uint8
	MsgType  uint8
	PushType uint8
	Token    uint8
	Number   uint16
	DataLen  uint16
	UserUuid string
	DestUuid string
	DeviceId uint32
	Data     []byte
}

func NewMessage(version, msgType, pushType, token uint8, number, dataLen uint16,
	userUuid, destUuid string, deviceId uint32, data []byte) (*Message, error) {
	if int(dataLen) != len(data) || len(userUuid) != 32 || len(destUuid) != 32 {
		err := errors.New("Message Create Error")
		return nil, err
	}

	message := &Message{
		Version:  version,
		MsgType:  msgType,
		PushType: pushType,
		Token:    token,
		Number:   number,
		DataLen:  dataLen,
		UserUuid: userUuid,
		DestUuid: destUuid,
		DeviceId: deviceId,
		Data:     data,
	}

	return message, nil
}

func (message *Message) Encode() (buffer []byte, err error) {
	buffer = append(buffer, byte(message.Version))
	buffer = append(buffer, byte(message.MsgType))
	buffer = append(buffer, byte(message.PushType))
	buffer = append(buffer, byte(message.Token))

	tempBuff := make([]byte, 16)
	binary.BigEndian.PutUint16(tempBuff, message.Number)
	buffer = append(buffer, tempBuff[:2]...)
	binary.BigEndian.PutUint16(tempBuff, message.DataLen)
	buffer = append(buffer, tempBuff[:2]...)
	if tempBuff, err = StringToUuid(message.UserUuid); err != nil {
		err = fmt.Errorf("Message Encode Error : %v", err)
		return buffer, err
	}
	buffer = append(buffer, tempBuff[:16]...)
	if tempBuff, err = StringToUuid(message.DestUuid); err != nil {
		err = fmt.Errorf("Message Encode Error : %v", err)
		return buffer, err
	}
	buffer = append(buffer, tempBuff[:16]...)

	binary.BigEndian.PutUint32(tempBuff, message.DeviceId)
	buffer = append(buffer, tempBuff[:4]...)

	buffer = append(buffer, message.Data...)
	return buffer, err
}
func (message *Message) Decode(buffer []byte) (err error) {
	if len(buffer) < 44 {
		err := errors.New("Message Decode Error: the []byte is shorter than 44")
		return err
	}
	index := 0
	message.Version = uint8(buffer[index])
	index++
	message.MsgType = uint8(buffer[index])
	index++
	message.PushType = uint8(buffer[index])
	index++
	message.Token = uint8(buffer[index])
	index++
	message.Number = binary.BigEndian.Uint16(buffer[index : index+2])
	index += 2
	message.DataLen = binary.BigEndian.Uint16(buffer[index : index+2])
	index += 2

	if message.UserUuid, err = UuidToString(buffer[index : index+16]); err != nil {
		err = fmt.Errorf("Message Decode Error: %v", err)
		return err
	}
	index += 16
	if message.DestUuid, err = UuidToString(buffer[index : index+16]); err != nil {
		err = fmt.Errorf("Message Decode Error: %v", err)
		return err
	}
	index += 16
	message.DeviceId = binary.BigEndian.Uint32(buffer[index : index+4])
	index += 4

	message.Data = buffer[index:]

	return err
}

type BRMsg struct {
	UserUuid string
	DeviceId uint32
}

func (this *BRMsg) Encode() (buffer []byte, err error) {
	buffer, err = StringToUuid(this.UserUuid)
	if err != nil {
		err = fmt.Errorf("BRMsg Encode Error : %v", err)
		return buffer, err
	}
	tempBuff := make([]byte, 4)
	binary.BigEndian.PutUint32(tempBuff, this.DeviceId)
	buffer = append(buffer, tempBuff...)
	return buffer, err
}

func (this *BRMsg) Decode(buffer []byte) (err error) {
	if len(buffer) < 20 {
		err = errors.New("BRMSg Decode Error: the []byte length is shorter than 20")
		return err
	}
	this.UserUuid, err = UuidToString(buffer[:16])
	if err != nil {
		err = fmt.Errorf("BRMsg Decode Error: %v", err)
		return err
	}
	this.DeviceId = binary.BigEndian.Uint32(buffer[16:20])
	return err
}
