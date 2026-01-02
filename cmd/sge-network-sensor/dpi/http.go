package dpi

import (
	"bytes"
)

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
func ParseHTTPRequest(payload []byte) (*HTTPRequest, bool) {
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
	// Basic implementation - scanning for "\r\nHost: "
	hostStart := bytes.Index(payload, headerHost)
	var host string
	if hostStart != -1 {
		start := hostStart + len(headerHost)
		end := bytes.IndexByte(payload[start:], '\r')
		if end != -1 {
			host = string(payload[start : start+end])
		}
	}

	return &HTTPRequest{
		Method: method,
		Host:   host,
	}, true
}
