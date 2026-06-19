// Package idgen provides a tiny UUIDv4 generator backed by crypto/rand.
// We avoid github.com/google/uuid in the initial scaffold so go build is
// hermetic; the wire format is identical so a swap later is invisible.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

// NewUUID returns a v4-shaped 36-char hex UUID. Not RFC-4122 strict on
// every reserved bit, but stable and unique enough for ids in tests and
// real loads (the DB still owns id generation in production via gen_random_uuid()).
func NewUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	// Set version (4) and variant (10) bits per RFC 4122.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	dst := make([]byte, 36)
	hex.Encode(dst[0:8], b[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], b[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], b[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], b[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], b[10:16])
	return string(dst)
}
