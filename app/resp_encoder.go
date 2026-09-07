package main

import (
	"fmt"
)

func EncodeSimpleString(s string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", s))
}

func EncodeBulkString(s string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func EncodeNullBulkString() []byte {
	return []byte(fmt.Sprintf("$-1\r\n"))
}

func EncodeInteger(num int) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", num))
}

func EncodeStringArray(str_arr []string) []byte {
	encode_str := fmt.Sprintf("*%d\r\n", len(str_arr))

	for _, str := range str_arr {
		encode_str += string(EncodeBulkString(str))
	}

	return []byte(encode_str)
}

func EncodeError(msg error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", msg))
}
