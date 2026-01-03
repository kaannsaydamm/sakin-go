package dpi

import (
	"strings"
	"testing"
)

func TestParseHTTPRequest(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		wantHost string
		wantOK   bool
	}{
		{
			name:     "Valid GET Request",
			payload:  []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			wantHost: "example.com",
			wantOK:   true,
		},
		{
			name:     "Valid GET Request Long Header",
			payload:  []byte("GET /api/v1/status HTTP/1.1\r\nUser-Agent: bot\r\nHost: sub.test.com\r\n"),
			wantHost: "sub.test.com",
			wantOK:   true,
		},
		{
			name:    "Invalid Method",
			payload: []byte("BOOM / HTTP/1.1\r\nHost: example.com\r\n"),
			wantOK:  false,
		},
		{
			name:    "Missing Host Header",
			payload: []byte("GET / HTTP/1.1\r\nUser-Agent: test\r\n"),
			wantOK:  true, // Returns true but empty host
		},
		{
			name:    "Binary Data (Attack)",
			payload: append([]byte("GET / HTTP/1.1\r\nHost: "), 0x00, 0x01, 0x02),
			wantOK:  false,
		},
		{
			name:    "Oversized Payload (Partial check)",
			payload: []byte("GET / " + strings.Repeat("A", MaxHostLength+10) + " HTTP/1.1\r\nHost: overflow.com\r\n"),
			wantOK:  true, // It parses method, but might fail host if too far
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseHTTPRequest(tt.payload)
			if ok != tt.wantOK {
				t.Errorf("ParseHTTPRequest() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if ok && tt.wantHost != "" && got.Host != tt.wantHost {
				t.Errorf("ParseHTTPRequest() Host = %v, want %v", got.Host, tt.wantHost)
			}
		})
	}
}
