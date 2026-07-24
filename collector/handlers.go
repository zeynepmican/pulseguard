package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gorm.io/gorm/clause"
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

// Filtered get function.
func handleGetEvents(w http.ResponseWriter, r *http.Request) {
	var events []Event

	// Temel GORM sorgumuzu başlatıyoruz (Henüz veritabanına gitmedik, sadece hazırlıyoruz)
	query := DB.Model(&Event{})

	// Filter: AgentID
	if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	// Filter Alarm Level
	if level := r.URL.Query().Get("level"); level != "" {
		query = query.Where("level = ?", level)
	}

	// Filter Check Type
	if checkType := r.URL.Query().Get("check_type"); checkType != "" {
		query = query.Where("check_type = ?", checkType)
	}

	// Use the query, get 50 events.
	result := query.Order("timestamp desc").Limit(50).Find(&events)
	if result.Error != nil {
		http.Error(w, "Failed to retrieve events from database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
	// Adding AgentID information to the event
	for i := range events {
		events[i].AgentID = agentID
	}

	// If this ID exists in the database do nothing
	result := DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&events)

	if result.Error != nil {
		fmt.Printf("DATABASE ERROR: %v\n", result.Error)
		http.Error(w, "Failed to save events to database", http.StatusInternalServerError)
		return
	}

	fmt.Printf("\nSECURE PAYLOAD SAVED TO DB | Agent: %s | Events: %d\n", agentID, len(events))
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

func handleGetHosts(w http.ResponseWriter, r *http.Request) {
	// Sadece GET isteklerine izin ver
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var agents []string

	// Sihirli GORM Sorgusu:
	// "Event tablosuna bak, 'agent_id' sütunundaki benzersiz isimleri bul ve 'agents' dizisine doldur."
	result := DB.Model(&Event{}).Distinct("agent_id").Pluck("agent_id", &agents)

	if result.Error != nil {
		http.Error(w, "Failed to retrieve hosts from database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(agents)
}
