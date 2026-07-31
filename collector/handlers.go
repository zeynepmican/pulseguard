package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm/clause"
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

	// Temel GORM sorgumuzu başlatıyoruz
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

	// YENİ: Dashboard'daki en uzun süre 1 saat olduğu için,
	// sadece son 1 saatin verilerini çekiyoruz ki veritabanı şişmesin.
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	query = query.Where("timestamp >= ?", oneHourAgo)

	// ESKİ: result := query.Order("timestamp desc").Limit(50).Find(&events)
	// YENİ: Limit kısıtlaması tamamen kaldırıldı!
	result := query.Order("timestamp desc").Find(&events)

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

	// -----------------------------------------------------------------
	// 1. AJANIN CONFIG DEĞERLERİNİ VERİTABANINDAN BUL
	// -----------------------------------------------------------------
	var config AgentConfig
	hasConfig := true
	if err := DB.First(&config, "agent_id = ?", agentID).Error; err != nil {
		hasConfig = false
		fmt.Printf("Uyarı: %s ajanı için config bulunamadı. Orijinal Level değerleri kullanılacak.\n", agentID)
	}

	// -----------------------------------------------------------------
	// 2. GELEN EVENTLERİ DÖNGÜYE SOK VE ID'LERİ EKLE
	// -----------------------------------------------------------------
	for i := range events {
		events[i].AgentID = agentID

		// If there is new config setting Collector change levels
		if hasConfig {
			val := events[i].Value
			events[i].Level = "INFO"

			switch events[i].CheckType {
			case "cpu":
				if val >= config.CpuErr {
					events[i].Level = "ERROR"
				} else if val >= config.CpuWarn {
					events[i].Level = "WARNING"
				}
			case "ram":
				if val >= config.RamErr {
					events[i].Level = "ERROR"
				} else if val >= config.RamWarn {
					events[i].Level = "WARNING"
				}
			case "disk":
				if val >= config.DiskErr {
					events[i].Level = "ERROR"
				} else if val >= config.DiskWarn {
					events[i].Level = "WARNING"
				}
			}
		}
	}

	// -----------------------------------------------------------------
	// 3. ÇAKIŞMALARI ÖNLEYEREK VERİTABANINA KAYDET
	// -----------------------------------------------------------------
	result := DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&events)

	if result.Error != nil {
		fmt.Printf("DATABASE ERROR: %v\n", result.Error)
		http.Error(w, "Failed to save events to database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payload saved securely"))
}

/*

// Save posted events to the database
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

	/*
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
*/

// Gets all agents (hosts)
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

// Gets host (agent) details
func handleGetHostDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL'den ajan ID'sini çıkar (Örn: /api/hosts/macbook-pro -> macbook-pro)
	agentID := strings.TrimPrefix(r.URL.Path, "/api/hosts/")
	if agentID == "" || agentID == r.URL.Path {
		http.Error(w, "Agent ID is required", http.StatusBadRequest)
		return
	}

	var detail HostDetail
	detail.AgentID = agentID

	// 1. Toplam olay sayısını bul
	DB.Model(&Event{}).Where("agent_id = ?", agentID).Count(&detail.TotalEvents)

	if detail.TotalEvents == 0 {
		http.Error(w, "Host not found", http.StatusNotFound)
		return
	}

	// 2. En son ne zaman veri gönderdiğini bul
	var lastEvent Event
	DB.Where("agent_id = ?", agentID).Order("timestamp desc").First(&lastEvent)
	detail.LastSeen = lastEvent.Timestamp

	// 3. Bu ajanın son 5 HATASINI getir
	DB.Where("agent_id = ? AND level = ?", agentID, "ERROR").Order("timestamp desc").Limit(5).Find(&detail.RecentErrors)

	// 4. Durum kontrolü: Son 5 dakikadır veri gelmiyorsa OFFLINE
	detail.Status = "Online"
	if time.Since(detail.LastSeen) > 5*time.Minute {
		detail.Status = "Offline"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(detail)
}

func handleAgentConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGetAgentConfig(w, r)
	} else if r.Method == http.MethodPost {
		handleUpdateAgentConfig(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetAgentConfig(w http.ResponseWriter, r *http.Request) {
	// URL'den Agent ID'yi alıyoruz (Örn: /api/configs/web-server-03 -> web-server-03)
	agentID := strings.TrimPrefix(r.URL.Path, "/api/configs/")
	if agentID == "" {
		http.Error(w, "Agent ID gerekli", http.StatusBadRequest)
		return
	}

	var config AgentConfig
	// Veritabanında ajana ait bir yapılandırma olup olmadığına bakıyoruz
	err := DB.First(&config, "agent_id = ?", agentID).Error

	if err != nil {
		// EĞER KAYIT YOKSA: Hata fırlatmıyoruz!
		// Arayüzün boş kalmaması için varsayılan eşik değerlerini döndürüyoruz
		config = AgentConfig{
			AgentID:  agentID,
			CpuWarn:  0.0,
			CpuErr:   0.0,
			RamWarn:  0.0,
			RamErr:   0.0,
			DiskWarn: 0.0,
			DiskErr:  0.0,
		}
	}

	// JSON olarak Dashboard'a veya Agent'a geri yolluyoruz
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// Updates the thresholds for specific agent
func handleUpdateAgentConfig(w http.ResponseWriter, r *http.Request) {

	agentID := strings.TrimPrefix(r.URL.Path, "/api/configs/")
	if agentID == "" {
		http.Error(w, "Agent ID gerekli", http.StatusBadRequest)
		return
	}

	var updatedConfig AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&updatedConfig); err != nil {
		fmt.Printf("Geçersiz JSON formatı: %v\n", err)
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}

	updatedConfig.AgentID = agentID // Güvenlik için ID'yi zorluyoruz

	// Veritabanında varsa güncelle, yoksa yeni oluştur (Save komutu bunu yapar)
	if err := DB.Save(&updatedConfig).Error; err != nil {
		fmt.Printf("Veritabanı kayıt hatası: %v\n", err)
		http.Error(w, "Veritabanına yazılamadı", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ayarlar başarıyla güncellendi"))
}
