package main

import (
	"errors"
	"fmt"
)

// Common Redis protocol and command errors
var (
	ErrSyntax          = errors.New("ERR syntax error")
	ErrNotInteger      = errors.New("ERR value is not an integer or out of range")
	ErrWrongType       = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	ErrUnsupportedType = errors.New("unsupported object type")

	// Stream errors
	ErrStreamIDSmaller = errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	ErrStreamIDZero    = errors.New("ERR The ID specified in XADD must be greater than 0-0")
	ErrInvalidStreamID = errors.New("ERR Invalid stream ID specified as stream command argument")

	// RESP decoder errors
	ErrNoNumberFound = errors.New("No number found")
	ErrEmptyStream   = errors.New("empty stream: no data to parse")
	ErrWrongDelim    = errors.New("wrong delimiter found!")
	ErrEmptyBuffer   = errors.New("empty buffer")
)

func ErrUnknownCommand(cmd string) error {
	return fmt.Errorf("ERR unknown command '%s'", cmd)
}

func ErrWrongArgs(cmd string) error {
	return fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd)
}
