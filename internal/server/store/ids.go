package store

import (
	"crypto/rand"
	"encoding/hex"
)

// newID erzeugt eine zufällige, dialektunabhängige ID (UUIDv4-ähnlich als Hex).
func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewID ist die exportierte Variante für andere Pakete (z.B. API-Handler).
func NewID() string { return newID() }
