package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type TextMsg struct {
	QCHeadBase
	TextData
}

func (text TextMsg) Encode() (msgBytes []byte, err error) {
	msgBytes, err = text.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("TextMsg Encode Error: %v", err)
		return msgBytes, err
	}
	buff, err = text.TextData.Encode()
	if err != nil {
		err = fmt.Errorf("TextMsg Encode Error: %v", err)
		return msgBytes, err
	}
	msgBytes = append(msgBytes, buf...)
}

func (text TextMsg) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < QCHEADBASE_LEN+16 {
		err = errors.New("TextMsg Decode() Error: []byte is shorter than 28 bytes")
		return err
	}
	err = text.QCHeadBase.Decode(msgBytes)
	if err != nil {
		err = fmt.Errorf("TextMsg Decode() Error: %v", err)
		return err
	}
	err = text.TextData.Decode(msgBytes[QCHEADBASE_LEN:])
	if err != nil {
		err = fmt.Errorf("TextMsg Decode() Error: %v", err)
		return err
	}
	return err
}
