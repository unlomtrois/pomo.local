package store

import (
	"crypto/rand"
	"encoding/hex"
)

// newHash returns a random 40-hex-char session hash (like a git object name,
// but assigned at creation rather than content-derived — sessions mutate).
func newHash() string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
