package protocol

import (
	"errors"
	"fmt"
)

type AudioMsg struct {
	QCHeadBase
	AudioData
}

func (audio AudioMsg) Encode() (msgBytes []byte, err error) {
	msgBytes, err = audio.QCHeadBase.Encode()
	if err != nil {
		err = fmt.Errorf("AudioMsg Encode Error: %v", err)
		return msgBytes, err
	}
	buff, err := audio.AudioData.Encode()
	if err != nil {
		err = fmt.Errorf("AudioMsg Encode Error: %v", err)
		return msgBytes, err
	}
	msgBytes = append(msgBytes, buff...)
	return msgBytes, err
}

func (audio AudioMsg) Decode(msgBytes []byte) (err error) {
	if len(msgBytes) < QCHEADBASE_LEN+32 {
		err = errors.New("AudioMsg Decode:the []byte length is shoter than 44 bytes")
		return err
	}
	err = audio.QCHeadBase.Decode(msgBytes)
	if err != nil {
		err = fmt.Errorf("AudioMsg Decode Error : %v", err)
		return err
	}
	err = audio.AudioData.Decode(msgBytes[QCHEADBASE_LEN:])
	if err != nil {
		err = fmt.Errorf("AudioMsg Decode Error: %v", err)
		return err
	}
	return err
}
