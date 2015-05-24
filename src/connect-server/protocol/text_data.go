package protocol

import (
	"errors"
	"fmt"
)

type TextData struct {
	DestUuid string
	Data     []byte
}

func (this *TextData) Encode() (buff []byte, err error) {
	buff = make([]byte, len(this.Data)+16)
	index := 0
	tmpBuf, err := StringToUuid(this.DestUuid)
	if err != nil {
		return buff, err
	}
	buff = append(buff, tmpBuf...)
	index += 16

	if len(this.Data) > 0 {
		buff = append(buff, this.Data...)
	}

	return buff, nil
}

func (this *TextData) Decode(buff []byte) (err error) {
	if len(buff) < 16 {
		err = errors.New("TextData Decode:the argument can't be shorter than 16 bytes")
		return err
	}
	this.DestUuid, err = UuidToString(buff[:16])
	if err != nil {
		err = fmt.Errorf("TextData Decode: %v", err)
		return err
	}
	this.Data = make([]byte, len(buff[16:]))
	n := copy(this.Data, buff[16:])
	if n != len(buff[16:]) {
		err = errors.New("TextData Decode: can't copy the full data to TextDataq.Data")
		return err
	}
	return err
}
