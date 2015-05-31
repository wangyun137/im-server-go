/*
*uuid to string and string to uuid function
*author zhangjingqjing jingqing.zhang@godinsec.com
*
*/
package utils

import (
	"strconv"
	"errors"
	"fmt"
)

func UuidToString(uuid []byte) (suuid string, err error) {
	if len(uuid) != 16 {
		err = errors.New("Error in protocal.UuidToString: invalid argument, the uuid should be 16 bytes")
		fmt.Println(err)
		return
	}
	for i := 0; i < 16; i++ {
		hc := uuid[i] >> 4 //the higher bits value
		lc := uuid[i] & 0xf //the lower bits value
		hcStr := strconv.FormatInt(int64(hc), 16)
		lcStr := strconv.FormatInt(int64(lc), 16)
		suuid += hcStr + lcStr
	}
	return
}

func StringToUuid(suuid string) (uuid []byte, err error) {
	if len(suuid) != 32 {
		err = errors.New("Error in protocal.StringToUuid: invalid argument, the string of uuid should be 32 chars and should not include '-'")
		fmt.Println(err)
		return
	}
	uuid = make([]byte, 16)
	for i := 0; i < 32; i+=2 {
		hv, err := strconv.ParseUint(string(suuid[i]), 16, 8)
		if err != nil {
			err = fmt.Errorf("Error in protocal.StringToUuid: %v", err)
			return uuid, err
		}
		lv, err := strconv.ParseUint(string(suuid[i + 1]), 16, 8)
		if err != nil {
			err = fmt.Errorf("Error in protocal.StringToUuid: %v", err)
			return uuid, err
		}
		uuid[i/2] = (uint8(hv) << 4) | (uint8(lv) & 0xf)
	}
	return
}
