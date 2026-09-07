package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Handler func(args []string) []byte

var ErrSyntax = errors.New("ERR syntax error")

func ErrUnknownCommand(cmd string) error {
	return fmt.Errorf("ERR unknown command '%s'", cmd)
}

func ErrWrongArgs(cmd string) error {
	return fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd)

}

func handlePing(args []string) []byte {
	return EncodeSimpleString("PONG")
}

func handleEcho(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrUnknownCommand("ECHO"))
	}
	return EncodeBulkString(args[0])
}

func handleSet(args []string, s *Store) []byte {
	if len(args) < 2 {
		return EncodeError(ErrWrongArgs("SET"))
	}

	key, value := args[0], args[1]

	switch len(args) {
	case 2:
		s.Set(key, value)
	case 4:
		time_option := strings.ToUpper(args[2])
		duration, _ := strconv.Atoi(args[3])

		ttl := time.Duration(duration)

		if time_option == "EX" {
			ttl = ttl * time.Second
		} else if time_option == "PX" {
			ttl = ttl * time.Millisecond
		}

		s.SetWithExpiry(key, value, ttl)
	}

	return EncodeSimpleString("OK")
}

func handleGet(args []string, s *Store) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("GET"))
	}

	key := args[0]

	value, exists := s.Get(key)

	if !exists {
		return EncodeNullBulkString()
	}

	return EncodeBulkString(value)
}

func BuildRegistry(s *Store) map[string]Handler {
	return map[string]Handler{
		"PING": handlePing,
		"ECHO": handleEcho,
		"SET": func(args []string) []byte {
			return handleSet(args, s)
		},
		"GET": func(args []string) []byte {
			return handleGet(args, s)
		},
	}
}
