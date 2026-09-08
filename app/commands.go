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
var ErrNotInteger = errors.New("ERR value is not an integer or out of range")

func ErrUnknownCommand(cmd string) error {
	return fmt.Errorf("ERR unknown command '%s'", cmd)
}

func ErrWrongArgs(cmd string) error {
	return fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd)

}

// base command

func handlePing(args []string) []byte {
	return EncodeSimpleString("PONG")
}

func handleEcho(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("ECHO"))
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

	value, exists, err := s.Get(key)

	if err != nil {
		return EncodeError(err)
	}

	if !exists {
		return EncodeNullBulkString()
	}
	return EncodeBulkString(value)
}

// lists command

func handleRPush(args []string, s *Store) []byte {
	if len(args) < 2 {
		return EncodeError(ErrWrongArgs("RPUSH"))
	}

	key := args[0]

	idx, err := s.RPush(key, args[1:]...)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(idx)
}

func handleLRange(args []string, s *Store) []byte {
	if len(args) != 3 {
		return EncodeError(ErrWrongArgs("LRANGE"))
	}

	key := args[0]

	start, _ := strconv.Atoi(args[1])
	stop, _ := strconv.Atoi(args[2])

	values, err := s.LRange(key, start, stop)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeStringArray(values)
}

func handleLPush(args []string, s *Store) []byte {
	if len(args) < 2 {
		return EncodeError(ErrWrongArgs("RPUSH"))
	}

	key := args[0]

	idx, err := s.RPush(key, args[1:]...)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(idx)
}

func handleLLen(args []string, s *Store) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("LLEN"))
	}

	key := args[0]

	len, err := s.LLen(key)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(len)

}

// streams command

func handleType(args []string, s *Store) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("TYPE"))
	}

	key := args[0]

	keyType := s.Type(key)

	return EncodeSimpleString(keyType)

}

// transactions command

func handleIncr(args []string, s *Store) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("INCR"))
	}

	key := args[0]

	num, err := s.Incr(key)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(num)
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
		"RPUSH": func(args []string) []byte {
			return handleRPush(args, s)
		},
		"LRANGE": func(args []string) []byte {
			return handleLRange(args, s)
		},
		"LPUSH": func(args []string) []byte {
			return handleLPush(args, s)
		},
		"LLEN": func(args []string) []byte {
			return handleLLen(args, s)
		},
		"TYPE": func(args []string) []byte {
			return handleType(args, s)
		},
		"INCR": func(args []string) []byte {
			return handleIncr(args, s)
		},
	}
}
