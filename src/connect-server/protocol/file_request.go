package protocol

import (
	"errors"
	"fmt"
)

type FileRequest struct {
	DestUuid string
}

func (request *FileRequest) Encode() (buffer []byte, err error) {
	buffer, err = StringToUuid(request.DestUuid)
	if err != nil {
		err = fmt.Errorf("FileRequest Encode Error: %v", err)
		return buffer, err
	}
	return buffer, err
}

func (request *FileRequest) Decode(buffer []byte) (err error) {
	if len(buffer) != 16 {
		err = errors.New("FileRequest Encode Error: the []byte length is shorter than 16")
		return err
	}
	index := 0
	request.DestUuid, err = UuidToString(buffer[index : index+16])
	if err != nil {
		err = fmt.Errorf("FileRequest Decode Error: %v", err)
		return err
	}
	return err
}

type FileMsgRequest struct {
	QCHeadBase
	FileRequest
}

func (request *FileMsgRequest) Encode() (buffer []byte, err error) {
	buffer, err = request.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsgRequest Encode Error: %v", err)
		return buffer, err
	}
	fileBuff, err := request.FileRequest.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsgRequest Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, fileBuff...)
	return buffer, err
}

func (request *FileMsgRequest) Decode(buffer []byte) (err error) {
	if len(buffer) != 28 {
		err = errors.New("FileMsgRequest Decode Error: the []byte length is shorter than 28")
		return err
	}
	err = request.QCHeadBase.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("FileMsgRequest Decode Error: %v", err)
		return err
	}
	err = request.FileRequest.Decode(buffer[QCHEADBASE_LEN:])
	if err != nil {
		err = fmt.Errorf("FileMsgRequest Decode Error: %v", err)
		return err
	}
	return err
}
