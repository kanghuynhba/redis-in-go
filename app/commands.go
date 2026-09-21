package main

import (
	"strconv"
	"strings"
	"time"
)

type Handler func(args []string) []byte

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
		timeOption := strings.ToUpper(args[2])
		duration, _ := strconv.Atoi(args[3])

		ttl := time.Duration(duration)

		if timeOption == "EX" {
			ttl = ttl * time.Second
		} else if timeOption == "PX" {
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

	encodeArray := NewEncodeArray()

	for _, value := range values {
		encodeArray.Append(EncodeBulkString(value))
	}

	return encodeArray.Bytes()
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
	delKeys, err := strconv.Atoi(args[1])

	if err != nil {
		return EncodeError(ErrWrongArgs("LPOP"))
	}

	values, err := h.store.LPop(key, delKeys)

	if err != nil {
		return EncodeError(err)
	}

	encodeArray := NewEncodeArray()

	for _, value := range values {
		encodeArray.Append(EncodeBulkString(value))
	}

	return encodeArray.Bytes()
}

func (h *CommandHandler) BLPop(args []string) []byte {
	if len(args) != 2 {
		return EncodeError(ErrWrongArgs("BLPOP"))
	}

	key := args[0]
	timeoutSec, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return EncodeError(ErrNotInteger)
	}

	timeout := time.Duration(timeoutSec * float64(time.Second))
	val, timedOut, err := h.store.BLPop(key, timeout)
	if err != nil {
		return EncodeError(err)
	}
	if timedOut {
		return EncodeNullArray()
	}

	encodeArray := NewEncodeArray()

	encodeArray.Append(EncodeBulkString(key))
	encodeArray.Append(EncodeBulkString(val))

	return encodeArray.Bytes()
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

func (h *CommandHandler) XAdd(args []string) []byte {
	if len(args) <= 2 {
		return EncodeError(ErrWrongArgs("XADD"))
	}

	key := args[0]
	ID := args[1]

	ID, err := h.store.XAdd(key, ID, args[2:])

	if err != nil {
		return EncodeError(err)
	}

	return EncodeBulkString(ID)
}

func (h *CommandHandler) XRANGE(args []string) []byte {
	if len(args) != 3 {
		return EncodeError(ErrWrongArgs("XRANGE"))
	}

	key := args[0]
	start := args[1]
	end := args[2]

	entries, err := h.store.XRange(key, start, end)

	if err != nil {
		return EncodeError(err)
	}

	encodeArray := NewEncodeArray()

	for _, entry := range entries {
		encodeArray.Append(entry.EncodeRESP())
	}

	return encodeArray.Bytes()
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
		"BLPOP":  h.BLPop,
		"TYPE":   h.Type,
		"XADD":   h.XAdd,
		"XRANGE": h.XRANGE,
		"INCR":   h.Incr,
	}
}
