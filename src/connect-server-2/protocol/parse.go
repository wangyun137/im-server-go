package protocol

import (
	"errors"
)

func EncodeLength(length int32) (buffer []byte, err error) {
	if length < 0 {
		err = errors.New("Parse Encode Error: the length is invalid")
		return buffer, err
	}
	var value int32 = length
	var tmpValue uint8 = 0
	for {
		if (value / 128) > 0 {
			tmpValue = uint8(value % 128)
			buffer = append(buffer, byte(tmpValue|128))
		} else {
			tmpValue = uint8(value % 128)
			buffer = append(buffer, byte(tmpValue|0))
			break
		}

		value = value / 128
	}

	return buffer, nil
}

func DecodeLength(buffer []byte) (size int, length int32, err error) {
	if len(buffer) < 0 {
		err = errors.New("Parse Decode Error:the buffer is invalid")
		return 0, length, err
	}
	i := 0
	index := 0
	for index := 0; index < 4; index++ {
		if buffer[index] >= 128 {
			if i == 0 {
				length += int32(buffer[index] - 128)
				i++
				continue
			}
			if i == 1 {
				length += int32(int32(buffer[index]-128) * 128)
				i++
				continue
			}
			if i == 2 {
				length += int32(int32(buffer[index]-128) * (128 * 128))
				i++
				continue
			}
		} else if buffer[index] < 128 {
			if i == 0 {
				length += int32(buffer[index])
				break
			}
			if i == 1 {
				length += int32(int32(buffer[index]) * 128)
				break
			}
			if i == 2 {
				length += int32(int32(buffer[index]) * (128 * 128))
				break
			}
			if i == 3 {
				length += int32(int32(buffer[index]) * (128 * 128 * 128))
				break
			}
		}
	}

	if index == 4 {
		err = errors.New("Parse Decode Error: the buffer length is invalid")
		return 0, length, err
	}

	return i + 1, length, nil
}
