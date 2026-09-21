package main

import (
	"strconv"
)

var (
	nullBulkString = []byte("$-1\r\n")
	nullArray      = []byte("*-1\r\n")
)

func EncodeSimpleString(s string) []byte {
	buf := make([]byte, 0, len(s)+3)
	buf = append(buf, '+')
	buf = append(buf, s...)
	buf = append(buf, '\r', '\n')
	return buf
}

func EncodeBulkString(s string) []byte {
	buf := make([]byte, 0, len(s)+16)
	buf = append(buf, '$')
	buf = strconv.AppendInt(buf, int64(len(s)), 10)
	buf = append(buf, '\r', '\n')
	buf = append(buf, s...)
	buf = append(buf, '\r', '\n')
	return buf
}

func EncodeNullBulkString() []byte {
	return nullBulkString
}

func EncodeInteger(num int) []byte {
	buf := make([]byte, 0, 24)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(num), 10)
	buf = append(buf, '\r', '\n')
	return buf
}

func EncodeNullArray() []byte {
	return nullArray
}

func EncodeError(msg error) []byte {
	errStr := msg.Error()
	buf := make([]byte, 0, len(errStr)+3)
	buf = append(buf, '-')
	buf = append(buf, errStr...)
	buf = append(buf, '\r', '\n')
	return buf
}

type EncodeArray struct {
	elements [][]byte
}

func NewEncodeArray() *EncodeArray {
	return &EncodeArray{
		elements: make([][]byte, 0),
	}
}

func (a *EncodeArray) Append(element []byte) {
	a.elements = append(a.elements, element)
}

func (a *EncodeArray) Bytes() []byte {
	// Pre-calculate total size to do a single allocation
	headerLen := 16 // enough for "*<len>\r\n"
	totalLen := headerLen
	for _, element := range a.elements {
		totalLen += len(element)
	}

	out := make([]byte, 0, totalLen)
	out = append(out, '*')
	out = strconv.AppendInt(out, int64(len(a.elements)), 10)
	out = append(out, '\r', '\n')

	for _, element := range a.elements {
		out = append(out, element...)
	}

	return out
}
