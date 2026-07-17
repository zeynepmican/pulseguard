package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateSignature creates an HMAC-SHA256 hash of the payload using the secret key.
func GenerateSignature(payload []byte, secretKey string) string {
	// SHA256 algorithm creates a new HMAC hash using the provided secret key. The key is converted to a byte slice.
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write the payload to the HMAC hash. This is where the actual hashing happens.
	h.Write(payload)

	// The final HMAC hash is computed and returned as a hexadecimal string.
	return hex.EncodeToString(h.Sum(nil))
}
