package protocol

import (
	"fmt"
)

type FileMsg struct {
	QCHeadBase
	FileData
}

func (this *FileMsg) Encode() (buffer []byte, err error) {
	buffer, err = this.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsg Encode Error: %v", err)
		return buffer, err
	}

	dataBuff, err := this.FileData.Encode()
	if err != nil {
		err = fmt.Errorf("FileMsg Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, dataBuff...)
	return buffer, nil
}

func (this *FileMsg) Decode(buffer []byte) (err error) {
	err = this.QCHeadBase.Decode(buffer)
	if err != nil {
		err = fmt.Errorf("FileMsg Decode Error: %v", err)
		return err
	}
	err = this.FileData.Decode(buffer[12:])
	if err != nil {
		err = fmt.Errorf("FileMsg Decode Error: %v", err)
		return err
	}

	return nil
}
