package main

import (
	"HttpFromTcp/internal/request"
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Accepted connection from:", conn.RemoteAddr())

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}
	// If successful, print the RequestLine in the specified format.
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, value := range req.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}

	fmt.Println("Connection closed:", conn.RemoteAddr())
}

func main() {
	// waits for tcp connection on port 42069
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Println("Listening on:42069....")

	for {
		// blocks until a client connects
		// Accept waits for and returns the next connection to the listener
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		// reads 8 bytes at a time, reconstructs lines and sends them to a channel
		// each clinet is handled in a separate goroutine
		go handleConnection(conn)
	}
}
