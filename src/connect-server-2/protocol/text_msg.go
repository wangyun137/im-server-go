package protocol

import (
	"errors"
	"fmt"
)

type TextMsg struct {
	HeadBase
	TextData
}

func (text TextMsg) Encode() ([]byte, error) {
	msgBytes := text.HeadBase.Encode()
	buffer, err := text.TextData.Encode()
	if err != nil {
		err = fmt.Errorf("Connect-Server TextMsg Encode Error: %v", err)
		return msgBytes, err
	}
	msgBytes = append(msgBytes, buffer...)
	return msgBytes, err
}

func (text TextMsg) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < HEADBASE_LEN+16 {
		err = errors.New("Connect-Server TextMsg Decode Error: []byte is shorter than 44 bytes")
		return err
	}
	err = text.HeadBase.Decode(msgBytes[:HEADBASE_LEN])
	if err != nil {
		err = fmt.Errorf("Connect-Server TextMsg Decode Error: %v", err)
		return err
	}
	err = text.TextData.Decode(msgBytes[HEADBASE_LEN:])
	if err != nil {
		err = fmt.Errorf("Connect-Server TextMsg Decode Error: %v", err)
		return err
	}
	return err
}
