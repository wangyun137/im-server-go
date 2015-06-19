package model

import (
	core "connect-server/protocol"
	"errors"
	"fmt"
	"td-server/protocol"
)

type Related struct {
	protocol.TsHead
	FromUuid string
	DestUuid string
}

func (related *Related) Encode() (buffer []byte, err error) {

	buffer, err = related.TsHead.Encode()
	if err != nil {
		err = fmt.Errorf("Related Encode Error: %v", err)
		return buffer, err
	}

	uuidBuff := make([]byte, 16)
	uuidBuff, err = core.StringToUuid(related.FromUuid)
	if err != nil {
		err = fmt.Errorf("Related Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, uuidBuff...)

	uuidBuff, err = core.StringToUuid(related.DestUuid)
	if err != nil {
		err = fmt.Errorf("Related Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, uuidBuff...)

	return buffer, nil
}

func (related *Related) Decode(buffer []byte) (err error) {
	if len(buffer) < 44 {
		err = errors.New("Related Decode Error: the []byte length is shorter than 44")
		return err
	}
	index := 0
	err = related.TsHead.Decode(buffer[index : index+12])
	if err != nil {
		err = fmt.Errorf("related Decode Error: %v", err)
		return err
	}
	index += 12

	related.FromUuid, err = core.UuidToString(buffer[index : index+16])
	if err != nil {
		err = fmt.Errorf("related Decode Error: %v", err)
		return err
	}
	index += 16

	related.DestUuid, err = core.UuidToString(buffer[index : index+16])
	if err != nil {
		err = fmt.Errorf("related Decode Error: %v", err)
		return err
	}
	index += 16

	return nil
}

type RelatedV struct {
	FromUuid string
	DestUuid string
}

func (this *RelatedV) Encode() (buff []byte, err error) {
	tmpbuff := make([]byte, 16)

	tmpbuff, err = core.StringToUuid(this.FromUuid)
	if err != nil {
		return nil, err
	}
	buff = append(buff, tmpbuff...)

	tmpbuff, err = core.StringToUuid(this.DestUuid)
	if err != nil {
		return nil, err
	}
	buff = append(buff, tmpbuff...)

	return buff, nil
}

func (this *RelatedV) Decode(buff []byte) (err error) {
	index := 0
	this.FromUuid, err = core.UuidToString(buff[index : index+16])
	if err != nil {
		return err
	}
	index += 16

	this.DestUuid, err = core.UuidToString(buff[index : index+16])
	if err != nil {
		return err
	}
	index += 16

	return nil
}
