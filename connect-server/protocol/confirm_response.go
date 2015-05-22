package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type ConfirmResponse struct {
	QCHeadBase
	Status int8
}

func (response *ConfirmResponse) Encode() (buffer []byte, err error) {
	headBuff, err := response.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("ConfirmResponse Encode Error: %v", headBuff)
		return headBuff, err
	}
	buffer = append(buffer, headBuff...)
	buffer = append(buffer, byte(response.Status))
	return buffer, nil
}

func (response *ConfirmResponse) Decode(buffer []byte) (err error) {
	if len(buffer) < 13 {
		err = errors.New("ConfirmResponse Decode Error: the []byte length is shorter than 13")
		return err
	}
	index := 0
	err = response.QCHeadBase.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("ConfirmResponse Decode Error: %v", err)
		return err
	}
	index += 12
	response.Status = int8(buffer[index])
	return nil
}
