package model

import (
	core "connect-server/protocol"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"td-server/protocol"
)

const (
	UUID_LENGTH    = 16
	TOKEN_LENGTH   = 16
	TIMEOUT_LENGTH = 8

	ALL_LENGTH = 2*UUID_LENGTH + 2*TOKEN_LENGTH + TIMEOUT_LENGTH + protocol.HEADLENGTH
)

type Token struct {
	protocol.TsHead
	FromUuid  string
	FromToken string
	DestUuid  string
	DestToken string
	TimeOut   int64
}

func (token *Token) Encode() (buffer []byte, err error) {
	buffer, err = token.TsHead.Encode()
	if err != nil {
		err = fmt.Errorf("Token Encode Error: %v", err)
		return buffer, err
	}

	uuidBuff := make([]byte, UUID_LENGTH)
	uuidBuff, err = core.StringToUuid(token.FromUuid)
	if err != nil {
		err = fmt.Errorf("Token Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, uuidBuff...)

	tokenBuff := make([]byte, TOKEN_LENGTH)
	tokenBuff, err = core.StringToUuid(token.FromToken)
	if err != nil {
		err = fmt.Errorf("Token Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, tokenBuff...)

	uuidBuff, err = core.StringToUuid(token.DestUuid)
	if err != nil {
		err = fmt.Errorf("Token Encode Error: %v", err)
		return buffer, err
	}

	tokenBuff, err = core.StringToUuid(token.DestToken)
	if err != nil {
		err = fmt.Errorf("Token Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, tokenBuff...)

	binary.BigEndian.PutUint64(uuidBuff, uint64(token.TimeOut))
	buffer = append(buffer, uuidBuff[:TOKEN_LENGTH]...)

	return buffer, nil
}

func (token *Token) Decode(buffer []byte) (err error) {
	if len(buffer) < ALL_LENGTH {
		err = errors.New("Token Decode Error: the []byte lenght is shorter than 84")
		return err
	}

	index := 0
	err = token.TsHead.Decode(buffer[:protocol.HEADLENGTH])
	if err != nil {
		err = fmt.Errorf("Token Decode Error : %v", err)
		return err
	}
	index += protocol.HEADLENGTH

	token.FromUuid, err = core.UuidToString(buffer[index : index+UUID_LENGTH])
	if err != nil {
		err = fmt.Errorf("Token Decode Error : %v", err)
		return err
	}
	index += UUID_LENGTH

	token.FromToken, err = core.UuidToString(buffer[index : index+TOKEN_LENGTH])
	if err != nil {
		err = fmt.Errorf("Token Decode Error : %v", err)
		return err
	}
	index += TOKEN_LENGTH

	token.DestUuid, err = core.UuidToString(buffer[index : index+UUID_LENGTH])
	if err != nil {
		err = fmt.Errorf("Token Decode Error : %v", err)
		return err
	}
	index += UUID_LENGTH

	token.DestToken, er = core.UuidToString(buffer[index : index+TOKEN_LENGTH])
	if err != nil {
		err = fmt.Errorf("Token Decode Error : %v", err)
		return err
	}
	index += TOKEN_LENGTH

	token.TimeOut = int64(binary.BigEndian.Uint64(buffer[index : index+TIMEOUT_LENGTH]))
	index += TIMEOUT_LENGTH

	return nil
}

type Tokener struct {
	FromToken string
	DestToken string
	TimeOut   int64
}

func (tokener *Tokener) Encode() (buffer []byte, err error) {
	buffer, err = core.StringToUuid(tokener.FromToken)
	if err != nil {
		err = fmt.Errorf("Tokener Encode Error: %v", err)
		return buffer, err
	}

	tempBuff := make([]byte, 16)
	tempBuff, err = core.StringToUuid(tokener.DestToken)
	if err != nil {
		err = fmt.Errorf("Tokener Encode Error: %v", err)
		return buffer, err
	}
	buffer = append(buffer, tempBuff...)

	binary.BigEndian.PutUint64(tempBuff, uint64(tokener.TimeOut))
	buffer = append(buffer, tempBuff[:8]...)
}

func (tokener *Tokener) Decode(buffer []byte) (err error) {
	if len(buffer) < 40 {
		err = errors.New("Tokener Decode Errror: the []byte length is shorter than 40")
		return err
	}

	index := 0
	tokener.FromToken, err = core.UuidToString(buffer[index : index+TOKEN_LENGTH])
	if err != nil {
		err = fmt.Errorf("Tokener Decode Errror: %v", err)
		return err
	}
	index += TOKEN_LENGTH

	tokener.DestToken, err = core.UuidToString(buffer[index : index+TOKEN_LENGTH])
	if err != nil {
		err = fmt.Errorf("Tokener Decode Errror: %v", err)
		return err
	}
	index += TOKEN_LENGTH

	tokener.TimeOut = int64(binary.BigEndian.Uint64(buffer[index : index+8]))
	return nil
}

type Tokens struct {
	Token map[string]*Tokener
	Lock  sync.Mutex
}

func NewTokens() *Tokens {
	return &Tokens{Token: make(map[string]*Tokener, 0)}
}

func (tokens *Tokens) Add(fromUuid, destUuid string, token *Tokener) (err error) {
	tokens.Lock.Lock()
	defer tokens.Lock.Unlock()
	if _, ok := tokens.Token[fromUuid+destUuid]; ok {
		err = errors.New("Tokens Add Error: exist " + fromUuid + destUuid)
		return err
	}
	tokens.Token[fromUuid+destUuid] = token
	return nil
}

func (tokens *Tokens) Delete(fromUuid, destUuid string) (err error) {
	tokens.Lock.Lock()
	defer tokens.Lock.Unlock()
	if _, ok := tokens.Token[fromUuid+destUuid]; ok {
		delete(tokens.Token, fromUuid+destUuid)
		return nil
	}
	err = errors.New("Tokens Delete Error: not exist " + fromUuid + destUuid)
	return err
}

func (tokens *Tokens) Query(fromUuid, destUuid string) (*Tokener, error) {
	tokens.Lock.Lock()
	defer tokens.Lock.Lock()
	if tokener, ok := tokens.Token[fromUuid+destUuid]; ok {
		return tokener, nil
	}
	err = errors.New("Tokens Query Error: not exist " + fromUuid + destUuid)
	return nil, err
}
