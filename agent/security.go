package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateSignature creates an HMAC-SHA256 hash of the payload using the secret key.
func GenerateSignature(payload []byte, secretKey string) string {
	// SHA256 algoritması ve Gizli Anahtarımız ile yeni bir HMAC kilidi oluştur
	h := hmac.New(sha256.New, []byte(secretKey))

	// Veriyi (JSON paketini) kilidin içine yaz
	h.Write(payload)

	// Oluşan karmaşık byte dizisini okunabilir bir onaltılık (hex) metne çevir
	return hex.EncodeToString(h.Sum(nil))
}
