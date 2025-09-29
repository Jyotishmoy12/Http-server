package headers

import (
	"bytes"
	"fmt"
	"strings"
)

// Headers is a map to store header key-value pairs.
type Headers map[string]string

// NewHeaders creates and returns an initialized Headers map.
func NewHeaders() Headers {
	return make(Headers)
}

// isValidTchar checks if a byte is a valid "tchar" as defined by RFC 9110.
func isValidTchar(b byte) bool {
	// Check for ALPHA
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
		return true
	}
	// Check for DIGIT
	if b >= '0' && b <= '9' {
		return true
	}
	// Check for special characters
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}

// Parse parses a single header line from a byte slice.
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIndex := bytes.Index(data, []byte("\r\n"))
	if crlfIndex == -1 {
		return 0, false, nil
	}

	if crlfIndex == 0 {
		return 2, true, nil
	}

	lineBytes := data[:crlfIndex]
	bytesConsumed := crlfIndex + 2

	colonIndex := bytes.Index(lineBytes, []byte(":"))
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid header line: missing colon")
	}

	keyBytes := lineBytes[:colonIndex]
	if bytes.HasSuffix(keyBytes, []byte(" ")) {
		return 0, false, fmt.Errorf("invalid header line: space before colon")
	}

	keyString := string(bytes.TrimSpace(keyBytes))
	if len(keyString) == 0 {
		return 0, false, fmt.Errorf("invalid header line: empty key")
	}

	// NEW: Validate all characters in the key.
	for i := 0; i < len(keyString); i++ {
		if !isValidTchar(keyString[i]) {
			return 0, false, fmt.Errorf("invalid character '%c' in header key", keyString[i])
		}
	}

	valueBytes := bytes.TrimSpace(lineBytes[colonIndex+1:])
	value := string(valueBytes)
	lowerKey := strings.ToLower(keyString)

	// check if the header key already exists
	if existingValue, ok := h[lowerKey]; ok {
		// if it exists append the new value separated by a comma
		h[lowerKey] = existingValue + "," + value
	} else {
		// otherwise just add the new header
		h[lowerKey] = value
	}

	return bytesConsumed, false, nil
}
