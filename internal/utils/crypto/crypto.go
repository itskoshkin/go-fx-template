package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRandomString returns a cryptographically secure random hex string of the given byte length.
// The resulting string is twice the byte length (e.g. 32 bytes = 64 hex characters).
func GenerateRandomString(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random string: %w", err)
	}

	return hex.EncodeToString(b), nil
}
