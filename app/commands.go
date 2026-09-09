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

type CommandHandler struct {
	store *Store
}

func NewCommandHandler(store *Store) *CommandHandler {
	return &CommandHandler{store: store}
}

// base commands

func (h *CommandHandler) Ping(args []string) []byte {
	return EncodeSimpleString("PONG")
}

func (h *CommandHandler) Echo(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("ECHO"))
	}
	return EncodeBulkString(args[0])
}

func (h *CommandHandler) Set(args []string) []byte {
	if len(args) < 2 {
		return EncodeError(ErrWrongArgs("SET"))
	}

	key, value := args[0], args[1]

	switch len(args) {
	case 2:
		h.store.Set(key, value)
	case 4:
		time_option := strings.ToUpper(args[2])
		duration, _ := strconv.Atoi(args[3])

		ttl := time.Duration(duration)

		if time_option == "EX" {
			ttl = ttl * time.Second
		} else if time_option == "PX" {
			ttl = ttl * time.Millisecond
		}

		h.store.SetWithExpiry(key, value, ttl)
	}

	return EncodeSimpleString("OK")
}

func (h *CommandHandler) Get(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("GET"))
	}

	key := args[0]

	value, exists, err := h.store.Get(key)

	if err != nil {
		return EncodeError(err)
	}

	if !exists {
		return EncodeNullBulkString()
	}
	return EncodeBulkString(value)
}

// list commands

func (h *CommandHandler) RPush(args []string) []byte {
	if len(args) < 2 {
		return EncodeError(ErrWrongArgs("RPUSH"))
	}

	key := args[0]

	idx, err := h.store.RPush(key, args[1:]...)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(idx)
}

func (h *CommandHandler) LRange(args []string) []byte {
	if len(args) != 3 {
		return EncodeError(ErrWrongArgs("LRANGE"))
	}

	key := args[0]

	start, _ := strconv.Atoi(args[1])
	stop, _ := strconv.Atoi(args[2])

	values, err := h.store.LRange(key, start, stop)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeStringArray(values)
}

func (h *CommandHandler) LPush(args []string) []byte {
	if len(args) < 2 {
		return EncodeError(ErrWrongArgs("LPUSH"))
	}

	key := args[0]

	idx, err := h.store.LPush(key, args[1:]...)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(idx)
}

func (h *CommandHandler) LLen(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("LLEN"))
	}

	key := args[0]

	length, err := h.store.LLen(key)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(length)
}

func (h *CommandHandler) LPop(args []string) []byte {
	if len(args) < 1 || len(args) > 2 {
		return EncodeError(ErrWrongArgs("LPOP"))
	}

	key := args[0]

	if len(args) == 1 {
		val, err := h.store.LPop(key, 1)

		if err != nil {
			return EncodeError(err)
		}

		return EncodeBulkString(val[0])
	}
	del_keys, err := strconv.Atoi(args[1])

	if err != nil {
		return EncodeError(ErrWrongArgs("LPOP"))
	}

	val, err := h.store.LPop(key, del_keys)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeStringArray(val)
}

// stream commands

func (h *CommandHandler) Type(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("TYPE"))
	}

	key := args[0]

	keyType := h.store.Type(key)

	return EncodeSimpleString(keyType)
}

// transaction commands

func (h *CommandHandler) Incr(args []string) []byte {
	if len(args) != 1 {
		return EncodeError(ErrWrongArgs("INCR"))
	}

	key := args[0]

	num, err := h.store.Incr(key)

	if err != nil {
		return EncodeError(err)
	}

	return EncodeInteger(num)
}

// command register table

func BuildRegistry(s *Store) map[string]Handler {
	h := NewCommandHandler(s)

	return map[string]Handler{
		"PING":   h.Ping,
		"ECHO":   h.Echo,
		"SET":    h.Set,
		"GET":    h.Get,
		"RPUSH":  h.RPush,
		"LPUSH":  h.LPush,
		"LRANGE": h.LRange,
		"LLEN":   h.LLen,
		"LPOP":   h.LPop,
		"TYPE":   h.Type,
		"INCR":   h.Incr,
	}
}
