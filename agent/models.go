package main

import "time"

type Event struct {
	EventID   string    `json:"event_id"`     // Olayın benzersiz kimliği (Idempotency için)
	Timestamp time.Time `json:"timestamp"`    // Olayın gerçekleştiği zaman
	CheckType string    `json:"check_type"`   // Hangi metrikte sorun oldu? (cpu, ram, disk)
	Level     string    `json:"level"`        // Seviye: INFO, WARNING, ERROR
	Message   string    `json:"message"`      // İnsan okuyabilir açıklama
	Value     float64   `json:"metric_value"` // O anki sorunlu metrik değeri
}

// Checker, sisteme eklenecek her yeni kontrol modülünün uyması gereken sözleşmedir.
type Checker interface {
	// Execute, kontrolü çalıştırır. Eğer eşik aşılmışsa bir Event döndürür.
	// Eşik aşılmamışsa Event 'nil' (boş) döner.
	Execute() (*Event, error)
}
