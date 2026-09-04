package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)

		if err != nil {
			break
		}

		p := NewRESPDecoder(buf[:n])

		p.decode()

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
