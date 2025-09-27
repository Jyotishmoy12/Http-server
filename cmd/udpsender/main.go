package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// resolve the servers udp address
	destAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Error resolving server address: %v", err)
	}
	fmt.Printf("Resolved server address: %v\n", destAddr)

	conn, err := net.DialUDP("udp", nil, destAddr)
	if err != nil {
		log.Fatalf("Error dialing server: %v", err)
	}

	defer conn.Close()

	fmt.Printf("Ready to yeet UDP packets to %v\n", conn.RemoteAddr())

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')

		if err != nil {
			log.Printf("Error reading input: %v\n", err)
			break
		}
		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Printf("Failed to send udp packet: %v\n", err)
			break
		}

	}

}
