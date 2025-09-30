package response

import (
	"HttpFromTcp/internal/headers"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// reasonPhrases maps status codes to their standard reason phrases

var reasonPhrases = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

// WriteStatusLin writes the first line of the HTTP response
//(e.g., "HTTP/1.1 200 OK").

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	// get the reason phrase for the status code from the map
	reasonPhrase := reasonPhrases[statusCode]
	// if the code isnot in the map the phrase will be an empty string

	// format the status line string
	line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)

	// write the status line to the writer
	_, err := w.Write([]byte(line))
	return err
}

func GetDefaultHeaders(constentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(constentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for key, value := range h {
		// // Canonicalize the key for proper HTTP formatting (e.g., "content-type" -> "Content-Type").
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		// format the header line
		line := fmt.Sprintf("%s: %s\r\n", canonicalKey, value)
		// write the line
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
	}
	// write the final blank line that separates the headers from the body
	_, err := w.Write([]byte("\r\n"))
	return err
}
