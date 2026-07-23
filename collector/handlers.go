package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

func verifySignature(payload []byte, incomingSignature string) bool {
	mac := hmac.New(sha256.New, []byte(appConfig.Security.SecretKey))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(incomingSignature))
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported. ", http.StatusMethodNotAllowed)
		return
	}

	signature := r.Header.Get("X-Signature")
	agentID := r.Header.Get("X-Agent-ID")

	if signature == "" {
		http.Error(w, "Unauthorized Access: Signature not found", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if !verifySignature(body, signature) {
		fmt.Printf("FAKE PAYLOAD BLOCKED! Agent: %s\n", agentID)
		http.Error(w, "Invalid Signature", http.StatusForbidden)
		return
	}

	fmt.Printf("SECURE PAYLOAD RECEIVED | Agent: %s | Size: %d bytes\n", agentID, len(body))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payload received securely"))
}
