package server

import (
	"bytes" // We need a buffer to capture the handler's output
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"HttpFromTcp/internal/request"
	"HttpFromTcp/internal/response"
)

// HandlerError represents an error that includes an HTTP status code.
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

// Error makes HandlerError satisfy the standard error interface.
func (e *HandlerError) Error() string {
	return e.Message
}

// Handler is a function type that defines our application logic.
// It takes a writer to build the response body and the parsed request.
type Handler func(w *bytes.Buffer, req *request.Request) *HandlerError

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
	handler  Handler // The server now holds a reference to the handler.
}

// Serve now accepts a handler function to process requests.
func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	server := &Server{
		listener: listener,
		handler:  handler, // Store the provided handler.
	}
	log.Printf("Listening on :%d...", port)
	go server.listen()
	return server, nil
}

// writeErrorResponse is a helper to keep error handling DRY.
func (s *Server) writeErrorResponse(conn net.Conn, err *HandlerError) {
	// Write the status line for the error (e.g., 400 or 500).
	if writeErr := response.WriteStatusLine(conn, err.StatusCode); writeErr != nil {
		log.Printf("Error writing error status line: %v", writeErr)
		return
	}
	// Create headers with the error message's length.
	headers := response.GetDefaultHeaders(len(err.Message))
	if writeErr := response.WriteHeaders(conn, headers); writeErr != nil {
		log.Printf("Error writing error headers: %v", writeErr)
		return
	}
	// Write the error message as the body.
	if _, writeErr := conn.Write([]byte(err.Message)); writeErr != nil {
		log.Printf("Error writing error body: %v", writeErr)
	}
}

// handle now orchestrates the request parsing and handler execution.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Step 1: Parse the incoming request.
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		// For a parsing error, we send a 400 Bad Request.
		s.writeErrorResponse(conn, &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Bad Request\n",
		})
		return
	}

	// Step 2: Create a buffer to capture the response body from the handler.
	responseBody := new(bytes.Buffer)

	// Step 3: Call the application's handler logic.
	handlerErr := s.handler(responseBody, req)

	// Step 4: Check if the handler returned an error.
	if handlerErr != nil {
		// If so, write the specific error response.
		s.writeErrorResponse(conn, handlerErr)
		return
	}

	// Step 5: If the handler succeeded, write the successful (200 OK) response.
	// Write the status line.
	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	// Create headers with the correct length from the buffer.
	headers := response.GetDefaultHeaders(responseBody.Len())
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	// Write the body that the handler generated.
	if _, err := conn.Write(responseBody.Bytes()); err != nil {
		log.Printf("Error writing response body: %v", err)
	}
}

// listen and Close methods remain unchanged.
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				log.Println("Server closed, shutting down listen loop.")
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}
func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}