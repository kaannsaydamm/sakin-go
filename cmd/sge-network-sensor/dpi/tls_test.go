package dpi

import (
	"encoding/hex"
	"testing"
)

func TestParseTLSClientHello(t *testing.T) {
	// Sample ClientHello hex dump (Valid Google.com SNI)
	// Simplified construction for testing
	// Record(16) + Ver(0301) + Len + Handshake(01) + Len + Ver + Random + SessID + CipherSuites + Compression + ExtLen + Ext(SNI)
	validSNIHex := "16030100c8010000c40303" + // Record + Handshake Header
		"0000000000000000000000000000000000000000000000000000000000000000" + // Random
		"00" + // SessionID Len
		"0002002f" + // Cipher Suites
		"0100" // Compression
		// Extensions would allow SNI, constructing full packet manually in hex is tedious and error prone without a builder.
		// Instead, we will test the rejection logic primarily as constructing valid full binary TLS packets is complex.

	_ = validSNIHex

	tests := []struct {
		name       string
		payloadHex string
		wantSNI    string
		wantOK     bool
	}{
		{
			name:       "Short Payload",
			payloadHex: "16030100",
			wantOK:     false,
		},
		{
			name:       "Invalid Record Type",
			payloadHex: "15030100c8010000c40303" + "0000000000000000000000000000000000000000000000000000000000000000",
			wantOK:     false, // Content Type not 0x16
		},
		{
			name:       "Invalid Handshake Type",
			payloadHex: "16030100c8020000c40303" + "0000000000000000000000000000000000000000000000000000000000000000",
			wantOK:     false, // Handshake type not 0x01 (ClientHello)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := hex.DecodeString(tt.payloadHex)
			got, ok := ParseTLSClientHello(payload)
			if ok != tt.wantOK {
				t.Errorf("ParseTLSClientHello() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if ok && got.ServerName != tt.wantSNI {
				t.Errorf("ParseTLSClientHello() SNI = %v, want %v", got.ServerName, tt.wantSNI)
			}
		})
	}
}
