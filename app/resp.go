package main

import (
	"bytes"
	"errors"
	"strconv"
)

type RESPDecoder struct {
	buf []byte
	pos int
}

func RESPParser(buf []byte) ([]string, error) {
	parser := &RESPDecoder{buf: buf, pos: 0}
	return parser.Decode()
}

func (p *RESPDecoder) readDecimal() (int, error) {
	idx := bytes.Index(p.buf[p.pos:], []byte("\r\n"))

	if idx == -1 {
		return -1, errors.New("No number found")
	}

	numBytes := p.buf[p.pos : p.pos+idx]

	// skip the digits and "\r\n"
	p.pos += (idx + 2)

	return strconv.Atoi(string(numBytes))
}

func (p *RESPDecoder) parseBulkHeader(delim byte) (int, error) {
	if len(p.buf[p.pos:]) == 0 {
		return -1, errors.New("empty stream: no data to parse")
	}

	if p.buf[p.pos] != delim {
		return -1, errors.New("wrong delimiter found!")
	}

	// skip the delim character (e.g '*', '$',...)
	p.pos += 1

	return p.readDecimal()
}

func (p *RESPDecoder) Decode() ([]string, error) {
	if len(p.buf[p.pos:]) == 0 {
		return nil, errors.New("empty buffer")
	}

	count, err := p.parseBulkHeader(byte('*'))

	if err != nil {
		return nil, err
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		length, err := p.parseBulkHeader(byte('$'))

		if err != nil {
			return nil, err
		}

		args[i] = string(p.buf[p.pos : p.pos+length])
		p.pos += (length + 2)
	}

	return args, nil
}
