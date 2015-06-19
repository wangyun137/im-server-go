package protocol

import (
	"errors"
	"fmt"
)

type TextData struct {
	DestUuid string
	Data     []byte
}

func (text *TextData) Encode() (buffer []byte, err error) {
	buffer = make([]byte, len(text.Data)+16)
	index := 0
	tmpBuff, err := StringToUuid(text.DestUuid)
	if err != nil {
		err = fmt.Errorf("Connec-Server TextData Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, tmpBuff...)
	index += 16

	if len(text.Data) > 0 {
		buffer = append(buffer, text.Data...)
	}

	return buffer, nil
}

func (text *TextData) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < 16 {
		err = errors.New("Connect-Server TextData Decode Error: argument can't be shorter than 16 bytes")
		return err
	}
	text.DestUuid, err = UuidToString(msgBytes[:16])
	if err != nil {
		err = fmt.Errorf("Connec-Server TextData Decode Error: %v", err)
		return err
	}
	text.Data = make([]byte, len(msgBytes[16:]))
	n := copy(text.Data, msgBytes[16:])
	if n != len(msgBytes[16:]) {
		err = errors.New("Connect-Server TextData Decode Error: can't copy full data to TextData.Data")
		return err
	}
	return err
}
