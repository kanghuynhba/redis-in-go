package main

import "fmt"

type Handler func(args []string) []byte

func handlePing(args []string) []byte {
	return []byte("+PONG\r\n")
}

func handleEcho(args []string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[0]), args[0]))
}

func BuildRegistry() map[string]Handler {
	return map[string]Handler{
		"PING": handlePing,
		"ECHO": handleEcho,
	}
}
