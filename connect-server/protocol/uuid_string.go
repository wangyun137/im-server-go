package protocol

import (
	"errors"
	"fmt"
	"strconv"
)

/*
  uuid必须为16个字节
*/
func UuidToString(uuid []byte) (uuidStr string, err error) {
	if len(uuid) != 16 {
		err = errors.New("The Uuid must be 16 bytes")
		return err
	}
	for i := 0; i < 16; i++ {
		higher := uuid[i] >> 4 //the higher bits value
		lower := uuid[i] & 0xf //the lower bits value
		higherStr := strconv.FormatInt(higher, 16)
		lowerStr := strconv.FormatInt(lower, 16)
		uuidStr += higherStr + lowerStr
	}
	return uuidStr
}

/*
  string长度必须为32
*/
func StringToUuid(uuidStr string) (uuid []byte, err error) {
	if len(uuidStr) != 32 {
		err = errors.New("The string which want to convert to uuid must be 32 bytes")
		return err
	}
	uuid = make([]byte, 16)
	for i := 0; i < 32; i += 2 {
		higher, err := strconv.ParseUint(string(uuid[i]), 16, 8)
		if err != nil {
			err = fmt.Errorf("Error in StringToUuid of higher: %v", err)
			return uuid, err
		}
		lower, err := strconv.ParseUint(string(uuid[i]), 16, 8)
		if err != nil {
			err = fmt.Errorf("Error in StringToUuid of lower: %v", err)
			return uuid, err
		}
		uuid[i/2] = (uint8(higher) << 4) | (uint8(lv) & 0xf)
	}
}
