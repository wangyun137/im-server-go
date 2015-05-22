package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type FileResponse struct {
	DestUuid   string
	IP         string //4bytes
	Port       uint16
	Token      string // 16bytes
	ConnMethod uint8  //tcp or udp
}

func (response FileResponse) Encode() (buffer []byte, err error) {
	buffer = make([]byte, 39)
	index := 0
	destBuff, err := StringToUuid(response.DestUuid)
	if err != nil {
		err = fmt.Errorf("FileResponse Encode Error: %v", err)
		return err
	}
	buffer = append(buffer, destBuff...)
	index += 16
	ip := net.ParseIP(response.IP)
	copy(buffer[index:index+4], ip.To4())

	index += 4
	binary.BigEndian.PutUint16(buf[index:index+2], response.Port)
	index += 2
	tokenBuff, err = StringToUuid(response.Token)
	if err != nil {
		err = fmt.Errorf("FileResponse Encode Error : %v", err)
		return err
	}
	copy(buffer[index:index+16], tokenBuff)
	index += 16
	buffer[index] = byte(response.ConnMethod)
	return buffer, err
}

func (response FileResponse) Decode(buffer []byte) (err error) {
	if len(buffer) != 39 {
		err = errors.New("FileResponse Decode Error: the []byte length is shorter than 39")
		return err
	}
	index := 0
	response.DestUuid, err = UuidToString(buffer[index : index+16])
	if err != nil {
		err = fmt.Errorf("FileResponse Decode Error: %v", err)
		return err
	}
	index += 16
	ip := net.IP(buffer[index : index+4])
	response.IP = ip
	index += 4
	response.Port = binary.BigEndian.Uint16(buffer[index : index+2])
	index += 2
	response.Token, err = UuidToString(buffer[index : index+16])
	if err != nil {
		err = fmt.Errorf("FileResponse Decode Error: %v", err)
		return err
	}
	index += 16
	response.ConnMethod = buffer[index]
	index += 1

	return err
}

type FileMsgResponse struct {
	QCHeadBase
	FileResponse
}

func (response FileMsgResponse) Encode() (buffer []byte, err error) {
	buffer, err = response.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsgResponse Encode Error: %v", err)
		return err
	}
	fileBuff, err := response.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsgResponse Encode Error: %v", err)
		return err
	}
	buffer = append(buffer, fileBuff...)
	return buffer, err
}

func (response FileMsgResponse) Decode(buffer []byte) (err error) {
	if len(buffer) != 51 {
		err = errors.New("FileMsgResponse Decode Error: the []byte length is shorter than 51")
		return err
	}

	err = response.QCHeadBase.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("FileMsgResponse Decode Error: %v", err)
		return err
	}
	err = response.FileResponse.Decode(buffer[QCHEADBASE_LEN:])
	if err != nil {
		err = fmt.Errorf("FileMsgResponse Decode Error: %v", err)
		return err
	}
	return err
}

type FileResponse2 struct {
	DestUuid string
	State    int8
}

func (response2 *FileResponse2) Encode() (buffer []byte, err error) {
	buffer = make([]byte, 17)
	destUuid, err := StringToUuid(response2.DestUuid)
	if err != nil {
		err = fmt.Errorf("FileResponse2 Encode Error: %v", v)
		return buffer, err
	}
	buffer = append(buffer, destUuid)
	buffer[16] = byte(response2.State)
	return
}

func (response2 *FileResponse2) Decode(buffer []byte) (err error) {
	if len(buf) != 17 {
		err = errors.New("FileResponse2 Decode Error:the []byte lenght is shorter than 17")
		return err
	}
	response2.DestUuid, err = UuidToString(buffer[:16])
	if err != nil {
		err = fmt.Errorf("FileResponse2 Decode Error: %v", err)
		return err
	}
	response2.State = int8(buffer[16])
	return err
}

type FileMsgResponse2 struct {
	QCHeadBase
	FileResponse2
}

func (response2 *FileMsgResponse2) Encode() (buffer []byte, err error) {
	buffer, err = response2.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsgResponse2 Encode Error: %v", err)
		return buffer, err
	}
	responseBuff, err := response2.FileResponse2.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsgResponse2 Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, responseBuff)

	return buffer, err
}

func (response2 *FileMsgResponse2) Decode(buffer []byte) (err error) {
	if len(buffer) != 29 {
		err = errors.New("FileMsgResponse2 Decode Error: the byte[] length is shorter 29")
		return err
	}
	err = response2.QCHeadBase.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("FileMsgResponse2 Decode Error:%v", err)
		return err
	}
	err = response2.FileResponse2.Decode(buffer[QCHEADBASE_LEN:])
	if err != nil {
		err = fmt.Errorf("FileMsgResponse2 Decode Error:%v", err)
		return err
	}
	return err
}
