package dpi

import (
	"encoding/binary"
	"unicode/utf8"
)

// TLS parsing safety limits
const (
	MaxTLSPayloadSize = 16384 // 16KB max for TLS record inspection
	MaxSNILength      = 253   // DNS name max length
	MinTLSHelloSize   = 43    // Minimum valid ClientHello size
)

// TLSClientHello represents minimal extracted TLS information.
type TLSClientHello struct {
	ServerName string
	Version    uint16
}

// ParseTLSClientHello attempts to extract SNI from a TCP payload.
// Includes safety bounds checking to prevent crashes from malformed data.
func ParseTLSClientHello(payload []byte) (*TLSClientHello, bool) {
	// Safety: limit payload size
	if len(payload) < MinTLSHelloSize {
		return nil, false
	}
	if len(payload) > MaxTLSPayloadSize {
		payload = payload[:MaxTLSPayloadSize]
	}

	// TLS Record Header
	// Content Type: Handshake (22)
	if payload[0] != 0x16 {
		return nil, false
	}

	// Version: TLS 1.0 (0x0301), 1.1 (0x0302), 1.2 (0x0303), 1.3 (0x0303 or similar legacy)
	// We just check major version 3
	if payload[1] != 0x03 {
		return nil, false
	}

	// Handshake Msg Type: Client Hello (1)
	// Skip Record Header (5 bytes) -> Handshake Header
	if payload[5] != 0x01 {
		return nil, false
	}

	// Skip Handshake Header (4 bytes: Type(1) + Length(3))
	// Client Version (2 bytes) + Random (32 bytes)
	offset := 5 + 4 + 2 + 32

	if offset >= len(payload) {
		return nil, false
	}

	// Session ID Length (1 byte)
	sessionIDLen := int(payload[offset])
	offset += 1 + sessionIDLen

	if offset+2 >= len(payload) {
		return nil, false
	}

	// Cipher Suites Length (2 bytes)
	cipherSuitesLen := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2 + cipherSuitesLen

	if offset+1 >= len(payload) {
		return nil, false
	}

	// Compression Methods Length (1 byte)
	compressionLen := int(payload[offset])
	offset += 1 + compressionLen

	if offset+2 >= len(payload) {
		return nil, false
	}

	// Extensions Length (2 bytes)
	extensionsLen := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2

	extensionsEnd := offset + extensionsLen
	if extensionsEnd > len(payload) {
		return nil, false
	}

	// Iterating extensions
	for offset+4 <= extensionsEnd {
		extType := binary.BigEndian.Uint16(payload[offset : offset+2])
		extLen := int(binary.BigEndian.Uint16(payload[offset+2 : offset+4]))
		offset += 4

		// Extension Type 0 is Server Name
		if extType == 0x0000 {
			if offset+extLen > extensionsEnd {
				return nil, false
			}

			// SNI List Length (2 bytes)
			if extLen < 2 {
				return nil, false
			}
			// listLen := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
			offset += 2 // skip list len

			// Server Name Type (1 byte) - 0 is host_name
			if offset >= extensionsEnd {
				return nil, false
			}
			nameType := payload[offset]
			offset++

			if nameType != 0 {
				return nil, false
			}

			// Server Name Length (2 bytes)
			if offset+2 > extensionsEnd {
				return nil, false
			}
			nameLen := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
			offset += 2

			// Safety: validate name length
			if nameLen == 0 || nameLen > MaxSNILength {
				return nil, false
			}

			if offset+nameLen > extensionsEnd {
				return nil, false
			}

			sniBytes := payload[offset : offset+nameLen]
			// Safety: validate UTF-8 and no control characters
			if !utf8.Valid(sniBytes) {
				return nil, false
			}
			for _, b := range sniBytes {
				if b < 32 || b == 127 {
					return nil, false
				}
			}

			return &TLSClientHello{ServerName: string(sniBytes)}, true
		}

		offset += extLen
	}

	return nil, false
}
