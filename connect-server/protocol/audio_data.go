package protocol

import (
	"errors"
	"fmt"
)

type AudioData struct {
	FromUuid string
	DestUuid string
	Data     []byte
}

func (this *AudioData) Encode() (buff []byte, err error) {
	tmpBuf := make([]byte, 16)
	tmpBuf, err = StringToUuid(this.FromUuid)
	if err != nil {
		err = fmt.Errorf("AudioData Encode Error: %V", err)
		return buff, err
	}
	buff = append(buff, tmpBuf...)

	tmpBuf, err = StringToUuid(this.DestUuid)
	if err != nil {
		return buff, err
	}

	buff = append(buf, tmpBuf...)

	buff = append(buff, this.Data...)

	return buff, nil
}

func (this *AudioData) Decode(buff []byte) (err error) {
	if len(buff) < 32 {
		err = errors.New("AudioData Decode: the args can't be shorter than 32 bytes")
		return err
	}

	var index int64 = 0
	this.FromUuid, err = UuidToString(buff[index : index+16])
	if err != nil {
		return err
	}
	index += 16

	this.DestUuid, err = UuidToString(buff[index : index+16])
	if err != nil {
		return err
	}

	index += 16
	if len(buff) > 32 {
		dataLen := int64(len(buff) - int64(index))
		this.Data = make([]byte, dataLen)
		n := copy(this.Data, buf[index:len(buff)])
		if n != dataLen {
			err = errors.New("AudioData Decode:can't copy the full data to AudioData.Data")
			return err
		}
	}

	return nil

}
