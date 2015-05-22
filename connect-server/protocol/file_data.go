package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type FileData struct {
	FromUuid string
	DestUuid string
	FileName string
	OffSet   int64
	Data     []byte
}

func (this *FileData) Encode() (buffer []byte, err error) {
	fromBuff := make([]byte, 16)
	fromBuff, err = StringToUuid(this.FromUuid)
	if err != nil {
		err = fmt.Errorf("FileData Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, fromBuff...)

	destBuff, err := StringToUuid(this.DestUuid)
	if err != nil {
		err = fmt.Errorf("FileData Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, destBuff...)

	length := uint8(len([]byte(this.FileName)))
	buffer = append(buffer, byte(length))
	buffer = append(buffer, []byte(this.FileName))
	binary.BigEndian.PutUint64(destBuff, uint8(this.OffSet))
	buffer = append(buffer, destBuff[:8]...)
	buffer = append(buffer, this.Data...)

	return buffer, nil
}

func (this *FileData) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < 41 {
		err = errors.New("FileData Decode Error: the []bytes < 41")
		return err
	}

	var index int64 = 0
	this.FromUuid, err = UuidToString(buf[index : index+16])
	if err != nil {
		err = fmt.Errorf("FileData Decode Error: %v", err)
		return err
	}
	index += 16

	this.DestUuid, err = UuidToString(buf[index : index+16])
	if err != nil {
		err = fmt.Errorf("FileData Decode Error: %v", err)
		return err
	}
	index += 16

	length := uint8(buffer[index])
	index += 1
	this.FileName = string(buffer[index : index+int64(length)])
	index += int64(length)
	this.OffSet = int64(binary.BigEndian.Uint64(buf[index : index+8]))
	index += 8

	if len(msgBytes) > 41 {
		this.Data = make([]byte, len(msgBytes)-index)
		copy(this.Data, msgBytes[index:])
	}

	return nil
}
