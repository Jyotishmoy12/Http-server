package main

import (
	"HttpFromTcp/internal/request"
	"HttpFromTcp/internal/response"
	"HttpFromTcp/internal/server"
	"bytes"
	"fmt"
	"log"
)

// appHandler contains our specific routing and business logic.
func appHandler(w *bytes.Buffer, req *request.Request) *server.HandlerError {
	// Route based on the request target (path).
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		// Return a 400 Bad Request error.
		return &server.HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		// Return a 500 Internal Server Error.
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    "Woopsie, my bad\n",
		}
	default:
		// For all other paths, write a success message to the response body.
		fmt.Fprint(w, "All good, frfr\n")
		// Return nil to indicate success.
		return nil
	}
}

func main() {
	// Pass our application handler to the server.
	s, err := server.Serve(42069, appHandler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer s.Close()

	// Keep the server running until manually stopped.
	select {}
}