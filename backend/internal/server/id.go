package server

import (
	"crypto/rand"
	"encoding/hex"
)

// newPlayerID generates a short random ID for a newly connected client.
// Good enough for a dev-scale server; swap for a more robust ID scheme
// (e.g. with persistent accounts) later if needed.
func newPlayerID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
