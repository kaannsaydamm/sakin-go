package utils

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
	"unsafe"
)

// --- Zero-Allocation String-Byte Conversion ---
// WARNING: Use with caution. Only for read-only access.

// StringToBytes converts string to byte slice without allocation.
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts byte slice to string without allocation.
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// --- High Performance Buffer Pool ---

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 4096) // 4KB buffer
	},
}

// GetBuffer retrieves a buffer from the pool.
func GetBuffer() []byte {
	return bufferPool.Get().([]byte)
}

// PutBuffer returns a buffer to the pool.
func PutBuffer(b []byte) {
	b = b[:0]
	bufferPool.Put(b)
}

// --- ID Generation ---

// GenerateID generates a random 16-byte hex ID.
// Optimized to reduce allocations.
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// --- Time Utilities ---

// NowUTC returns current time in UTC, truncated to milliseconds for consistency.
func NowUTC() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}
