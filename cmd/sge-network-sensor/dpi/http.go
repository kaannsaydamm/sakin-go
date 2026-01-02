package dpi

import (
	"bytes"
	"unicode/utf8"
)

// MaxPayloadSize limits the amount of data we inspect to prevent DoS
const MaxPayloadSize = 8192

// MaxHostLength limits extracted Host header length
const MaxHostLength = 255

var (
	httpMethods = [][]byte{
		[]byte("GET "),
		[]byte("POST "),
		[]byte("PUT "),
		[]byte("DELETE "),
		[]byte("HEAD "),
		[]byte("OPTIONS "),
		[]byte("PATCH "),
	}
	headerHost = []byte("\r\nHost: ")
)

// HTTPRequest info
type HTTPRequest struct {
	Method string
	Host   string
	URI    string
}

// ParseHTTPRequest extracts HTTP details from payload if present.
// Includes safety checks against malformed/malicious input.
func ParseHTTPRequest(payload []byte) (*HTTPRequest, bool) {
	// Safety: limit payload size to prevent CPU exhaustion
	if len(payload) == 0 {
		return nil, false
	}
	if len(payload) > MaxPayloadSize {
		payload = payload[:MaxPayloadSize]
	}

	// Safety: check for null bytes (binary data, not HTTP)
	if bytes.IndexByte(payload[:min(256, len(payload))], 0) != -1 {
		return nil, false
	}

	// 1. Check method
	var method string
	for _, m := range httpMethods {
		if bytes.HasPrefix(payload, m) {
			method = string(m[:len(m)-1]) // remove space
			break
		}
	}
	if method == "" {
		return nil, false
	}

	// 2. Extract Host header
	hostStart := bytes.Index(payload, headerHost)
	var host string
	if hostStart != -1 {
		start := hostStart + len(headerHost)
		end := bytes.IndexByte(payload[start:], '\r')
		if end != -1 && end <= MaxHostLength {
			hostBytes := payload[start : start+end]
			// Validate: must be valid UTF-8, no control chars
			if utf8.Valid(hostBytes) && !containsControlChars(hostBytes) {
				host = string(hostBytes)
			}
		}
	}

	return &HTTPRequest{
		Method: method,
		Host:   host,
	}, true
}

// containsControlChars checks for ASCII control characters (except HTAB)
func containsControlChars(b []byte) bool {
	for _, c := range b {
		if c < 32 && c != 9 { // allow tab (9)
			return true
		}
		if c == 127 { // DEL
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
