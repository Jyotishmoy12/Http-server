package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chunkReader simulates reading from a network connection in small, fixed-size chunks.
type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call.
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestFromReader(t *testing.T) {
	t.Run("Good GET Request line with small chunks", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good GET Request line with path and tiny chunks", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good POST Request with large chunk", func(t *testing.T) {
		data := "POST /submit HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		reader := &chunkReader{
			data:            data,
			numBytesPerRead: len(data),
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/submit", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})
}