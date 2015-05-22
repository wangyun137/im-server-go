package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Confirm struct {
	QCHeadBase
	UserUuid string
	Port     uint16
}

func (confirm *Confirm) Encode() (buffer []byte, err error) {
	headBuff := confirm.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("Confirm Encode Error: %v", headBuff)
		return buffer, err
	}
	buffer = append(buffer, headBuff...)

	userBuff := make([]byte, 16)
	userBuff, err = StringToUuid(confirm.UserUuid)
	if err != nil {
		err = fmt.Errorf("Confirm Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, userBuff[:16]...)
	binary.BigEndian.PutUint16(userBuff[:2], uint16(confirm.Port))
	buffer = append(buffer, userBuff[:2])

	return buffer, nil
}

func (confirm *Confirm) Decode(buffer []byte) error {
	if len(buffer) < 30 {
		err := errors.New("Confirm Decode Error: the []byte length is shorter than 30")
		return err
	}
	index := 0
	err := confirm.QCHeadBase.Decode(buffer)
	if err != nil {
		return err
	}
	index += 12

	confirm.UserUuid, err = UuidToString(buffer[index : index+16])
	if err != nil {
		return err
	}
	index += 16
	confirm.Port = uint16(binary.BigEndian.Uint16(buffer[index : index+2]))

	return nil
}
