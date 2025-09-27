//string parser for requestLine:
// if the startline is a part of GET request then it called as requestLine
// GET /coffee HTTP/1.1

package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	StateRequestLine = iota // initial state, value is 0
	StateDone               // parsing is complete value is 1
)

type Request struct {
	RequestLine RequestLine
	state       int // tracks the current state of the parser
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state: StateRequestLine, // start with the request line
	}
	// buffer accumulate data across multiple reads from the stream
	var buffer []byte
	// tmp is a temporary space to read the next chunk into

	tmp := make([]byte, 1024)

	// loop until the parser's state is done

	for req.state != StateDone {
		// 1: Read the next chunk of data from the reader
		n, err := reader.Read(tmp)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read from reader:%w", err)
		}
		// 2: append the newly read chunk to our main buffer
		buffer = append(buffer, tmp[:n]...)

		// 3: pass the accumulated buffer to the state machine to parse what it can
		bytesConsumed, parseErr := req.parse(buffer)
		if parseErr != nil {
			return nil, parseErr
		}
		// 4: shift the buffer by removing the bytes that were consumed
		buffer = buffer[bytesConsumed:]

		// 5: if we have reached the end of the stream EOF
		// and our buffer is emoty.
		// we cant read any more data from the stream
		if err == io.EOF && len(buffer) == 0 {
			break
		}
	}
	// After the loop, if the parser isn't in the "Done" state, it means the
	// stream ended before we could parse a complete request line.
	if req.state != StateDone {
		return nil, errors.New("incomplete request: stream ended before request line was parsed")
	}

	return req, nil
}
func (r *Request) parse(data []byte) (int, error) {
	if r.state == StateRequestLine {
		// try to parse the request line from the current data
		rl, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		// if bytesConsumed is 0 it means we need more data
		if bytesConsumed == 0 {
			return 0, nil
		}
		// success! update the request and change the state
		r.RequestLine = rl
		r.state = StateDone
		return bytesConsumed, nil
	}
	return 0, nil
}

// parseRequestLine now takes a byte slice and returns the number of bytes consumed.
// If it needs more data, it returns 0 bytes consumed and no error.
func parseRequestLine(data []byte) (RequestLine, int, error) {
	// find the end of the line: \r\n
	crlfIndex := bytes.Index(data, []byte("\r\n"))
	if crlfIndex == -1 {
		// No "\r\n" found, so the line is incomplete. Signal that we need more data.
		return RequestLine{}, 0, nil
	}
	// we found the end of the line, Extract it
	lineBytes := string(data[:crlfIndex])
	bytesConsumed := crlfIndex + 2 // +2 for the \r\n characters

	// spilt into 3 parts
	parts := strings.Split(string(lineBytes), " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: expected 3 parts, got %d", len(parts))
	}
	method, target, versionRaw := parts[0], parts[1], parts[2]
	// 2. Verify that the "method" part only contains capital alphabetic characters.
	for _, char := range method {
		if char < 'A' || char > 'Z' {
			return RequestLine{}, 0, fmt.Errorf("invalid method '%s': must be all uppercase alphabetic characters", method)
		}
	}

	// 3. Verify that the http version part is HTTP/1.1.
	if versionRaw != "HTTP/1.1" {
		return RequestLine{}, 0, fmt.Errorf("invalid http version '%s': only HTTP/1.1 is supported", versionRaw)
	}
	rl := RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   "1.1", // Extract "1.1" from "HTTP/1.1"
	}
	return rl, bytesConsumed, nil
}
