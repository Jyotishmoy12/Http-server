package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders_Parse(t *testing.T) {

	t.Run("Duplicate header key should append value", func(t *testing.T) {
		headers := NewHeaders()
		headers["accept-language"] = "en-US"

		data:= [] byte("Accept-Language: en-GB\r\n")
		n, done, err:= headers.Parse(data)
		require.NoError(t, err)
		assert.False(t, done)
		assert.NotZero(t, n)

		assert.Equal(t, "en-US,en-GB", headers["accept-language"])
	})
	t.Run("Valid single header with case-insensitivity", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Content-Type:    application/json    \r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "application/json", headers["content-type"])
		assert.Equal(t, 39, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["existing"] = "true"

		data1 := []byte("Host: example.com\r\nUser-Agent: test\r\n")
		n1, done1, err1 := headers.Parse(data1)
		require.NoError(t, err1)
		assert.Equal(t, 19, n1)
		assert.False(t, done1)

		remainingData := data1[n1:]
		n2, done2, err2 := headers.Parse(remainingData)
		require.NoError(t, err2)
		assert.Equal(t, 18, n2)
		assert.False(t, done2)

		assert.Equal(t, "true", headers["existing"])
		assert.Equal(t, "example.com", headers["host"])
		assert.Equal(t, "test", headers["user-agent"])
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host : localhost:42069\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "space before colon")
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid character in header key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Incomplete header needs more data", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})
}
