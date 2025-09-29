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

// This function tests the header parsing logic.
func TestRequestFromReader_HeaderParsing(t *testing.T) {
	t.Run("Standard Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
		assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
		assert.Equal(t, "*/*", r.Headers.Get("accept"))
	})

	t.Run("Malformed Header", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 5,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})

	t.Run("Duplicate Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n",
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "text/html,application/json", r.Headers.Get("accept"))
	})
}

// This new function tests the body parsing logic.
func TestRequestFromReader_BodyParsing(t *testing.T) {
	t.Run("Standard Body", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello world!\n", string(r.Body))
	})

	t.Run("Empty Body, 0 reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Empty(t, r.Body)
	})

	t.Run("Empty Body, no reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Empty(t, r.Body)
	})

	t.Run("Body shorter than reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content", // Only 15 bytes
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incomplete request")
	})

	t.Run("No Content-Length but Body Exists", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n" +
				"this body will be ignored",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Empty(t, r.Body)
	})
}