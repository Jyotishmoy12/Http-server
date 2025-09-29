package main

import (
	"HttpFromTcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)
	// sign chan is a buffered channel of size 1
	sigChan := make(chan os.Signal, 1)
	// Notify registers the signal s to the channel sigChan
	// This is a non-blocking call so no goroutine is created
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server graceful shutdown")
}
