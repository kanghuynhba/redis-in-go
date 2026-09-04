package main

import (
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	register := BuildRegistry()

	for {
		n, err := conn.Read(buf)

		if err != nil {
			break
		}

		parse_result, err := RESPParser(buf[:n])

		command := strings.ToUpper(parse_result[0])
		args := parse_result[1:]

		handler, exists := register[command]

		if !exists {
			conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
			continue
		}

		response := handler(args)

		conn.Write(response)
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")

	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		return
	}

	defer l.Close()

	for {

		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Cannot connect to client")
			continue
		}

		// Handle each connection concurrently in a separate goroutine
		go handleConnection(conn)
	}
}
