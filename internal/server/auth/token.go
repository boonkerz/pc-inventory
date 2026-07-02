package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GenerateToken erzeugt ein kryptografisch sicheres, URL-sicheres Token (32 Byte Entropie).
// Verwendet für Session-, Agent- und Enrollment-Tokens.
func GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashToken bildet den SHA-256-Hash eines Tokens (Hex). Token werden niemals
// im Klartext gespeichert; Lookups erfolgen über den Hash.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
