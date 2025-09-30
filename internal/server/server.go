package server

import (
	"HttpFromTcp/internal/response"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool // used to gracefully shautdown the server
}

func Serve(port int) (*Server, error) {
	// create the address string for the listener
	addr := fmt.Sprintf(":%d", port)

	// create a tcp listener on the specified port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", port, err)
	}

	// create a new server instance
	server := &Server{
		listener: listener,
	}
	log.Printf("Server started on port %d", port)
	// start the main accept loop in a separate goroutine
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	// set the isclosed flag to true to signal the listen loop to stop
	s.isClosed.Store(true)
	// close the underlying TCP listener
	return s.listener.Close()
}

func (s *Server) listen() {
	// start an infinite loop to continously accept new connections
	for {
		// Accept() blocks until a new connection is made then returns it
		conn, err := s.listener.Accept()
		if err != nil {
			// if the server has been closed, isClosed will be true
			// we can safely ignore the error and exit the loop
			if s.isClosed.Load() {
				log.Println("Server closed shutting down listener")
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue // continue to next iteraion on other errors
		}
		// handle each new connection in its own go routing
		go s.handle(conn)
	}
}

// handle manages a single client connection.
func (s *Server) handle(conn net.Conn) {
	// Ensure the connection is closed when this function exits.
	defer conn.Close()
	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	// get the default headers (with a body of length 0 for now)
	defaultHeaders := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, defaultHeaders); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
}
