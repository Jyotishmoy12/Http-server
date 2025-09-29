package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv" // New import for converting strings to integers

	"HttpFromTcp/internal/headers"
)

// Add the new StateBody
const (
	StateRequestLine = iota
	StateHeaders
	StateBody
	StateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte // Add the Body field
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state:   StateRequestLine,
		Headers: headers.NewHeaders(),
	}
	var buffer []byte
	tmp := make([]byte, 1024)
	for req.state != StateDone {
		n, err := reader.Read(tmp)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read from reader: %w", err)
		}
		if n > 0 {
			buffer = append(buffer, tmp[:n]...)
		}
		bytesConsumed, parseErr := req.parse(buffer)
		if parseErr != nil {
			return nil, parseErr
		}
		buffer = buffer[bytesConsumed:]
		if err == io.EOF {
			if req.state != StateDone {
				return nil, errors.New("incomplete request: stream ended before request was fully parsed")
			}
			break
		}
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != StateDone {
		if len(data[totalBytesParsed:]) == 0 {
			break
		}
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

// parseSingle now includes logic for parsing the body.
func (r *Request) parseSingle(data []byte) (int, error) {
	// Step 1: Parse the request line.
	if r.state == StateRequestLine {
		rl, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesConsumed == 0 {
			return 0, nil
		}
		r.RequestLine = rl
		r.state = StateHeaders // Transition to parsing headers.
		return bytesConsumed, nil
	}

	// Step 2: Parse the headers.
	if r.state == StateHeaders {
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = StateBody // When headers are done, transition to parsing the body.
		}
		return n, nil
	}

	// Step 3: Parse the body.
	if r.state == StateBody {
		// Get the Content-Length to know how many bytes to read.
		contentLengthStr := r.Headers.Get("Content-Length")
		if contentLengthStr == "" {
			// If no Content-Length, assume no body and we're done.
			r.state = StateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length header: %w", err)
		}

		if contentLength == 0 {
			r.state = StateDone
			return 0, nil
		}

		// Figure out how many bytes to consume from the current data chunk.
		bytesNeeded := contentLength - len(r.Body)
		bytesToConsume := len(data)
		if bytesToConsume > bytesNeeded {
			bytesToConsume = bytesNeeded
		}

		// Append the consumed bytes to the body.
		r.Body = append(r.Body, data[:bytesToConsume]...)

		// Check if we've read the entire body.
		if len(r.Body) == contentLength {
			r.state = StateDone
		}

		// Check if we've read too much (shouldn't happen with the logic above, but good practice).
		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("body is longer than Content-Length")
		}

		return bytesToConsume, nil
	}

	return 0, nil
}
// parseRequestLine remains unchanged
func parseRequestLine(data []byte) (RequestLine, int, error) {
	crlfIndex := bytes.Index(data, []byte("\r\n"))
	if crlfIndex == -1 {
		return RequestLine{}, 0, nil
	}
	lineBytes := data[:crlfIndex]
	bytesConsumed := crlfIndex + 2
	parts := bytes.Split(lineBytes, []byte(" "))
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: expected 3 parts, got %d", len(parts))
	}
	method, target, versionRaw := parts[0], parts[1], parts[2]
	for _, char := range method {
		if char < 'A' || char > 'Z' {
			return RequestLine{}, 0, fmt.Errorf("invalid method '%s': must be all uppercase", string(method))
		}
	}
	if string(versionRaw) != "HTTP/1.1" {
		return RequestLine{}, 0, fmt.Errorf("invalid http version '%s': only HTTP/1.1 is supported", string(versionRaw))
	}
	rl := RequestLine{
		Method:        string(method),
		RequestTarget: string(target),
		HttpVersion:   "1.1",
	}
	return rl, bytesConsumed, nil
}