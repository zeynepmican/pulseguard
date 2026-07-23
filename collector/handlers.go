package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostEvents(w, r)
	case http.MethodGet:
		handleGetEvents(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetEvents(w http.ResponseWriter, r *http.Request) {
	var events []Event

	// Veritabanından en son eklenen 50 olayı tarihe göre tersten (en yeni en üstte) sıralayarak çek
	result := DB.Order("timestamp desc").Limit(50).Find(&events)
	if result.Error != nil {
		http.Error(w, "Failed to retrieve events from database", http.StatusInternalServerError)
		return
	}

	// Geri dönüş tipinin JSON olduğunu belirtiyoruz ki tarayıcılar güzel göstersin
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Veritabanından çektiğimiz diziyi JSON formatına çevirip ekrana (istemciye) basıyoruz
	json.NewEncoder(w).Encode(events)
}

func handlePostEvents(w http.ResponseWriter, r *http.Request) {
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

	var events []Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		fmt.Printf("JSON PARSING ERROR: %v\n", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// GORM ile veritabanına kaydet
	result := DB.Create(&events)
	if result.Error != nil {
		fmt.Printf("DATABASE ERROR: %v\n", result.Error)
		http.Error(w, "Failed to save events to database", http.StatusInternalServerError)
		return
	}

	fmt.Printf("\n✅ SECURE PAYLOAD SAVED TO DB | Agent: %s | Events: %d\n", agentID, len(events))
	for _, event := range events {
		icon := "ℹ️"
		if event.Level == "WARNING" {
			icon = "⚠️"
		} else if event.Level == "ERROR" {
			icon = "🚨"
		}
		fmt.Printf("%s [%-7s] [%s] %-4s (%.2f): %s\n",
			icon, event.Level, event.Timestamp.Format("15:04:05"),
			event.CheckType, event.Value, event.Message)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payload saved securely"))
}
