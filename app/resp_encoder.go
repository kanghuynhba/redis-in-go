package main

import "fmt"

func EncodeSimpleString(s string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", s))
}

func EncodeBulkString(s string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func EncodeNullBulkString() []byte {
	return []byte(fmt.Sprintf("$-1\r\n"))
}

func EncodeError(msg error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", msg))
}
