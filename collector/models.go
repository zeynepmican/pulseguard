package main

import "time"

type Event struct {
	ID      uint   `gorm:"primaryKey" json:"-"`
	EventID string `gorm:"uniqueIndex" json:"event_id"`
	AgentID string `gorm:"index" json:"agent_id"`
	// EventID   string    `json:"event_id"`     // Event's unique identifier
	Timestamp time.Time `json:"timestamp"`    // Timestamp of the event
	CheckType string    `json:"check_type"`   // Which metric had an issue? (cpu, ram, disk)
	Level     string    `json:"level"`        // Level: INFO, WARNING, ERROR
	Value     float64   `json:"metric_value"` // Current problematic metric value
}

// HostDetail, tek bir ajanın sağlık özetini temsil eder
type HostDetail struct {
	AgentID      string    `json:"agent_id"`
	TotalEvents  int64     `json:"total_events"`
	LastSeen     time.Time `json:"last_seen"`
	Status       string    `json:"status"`
	RecentErrors []Event   `json:"recent_errors"`
}

// YENİ: Ajanların eşik değerlerini tutacak ayar tablomuz
type AgentConfig struct {
	AgentID  string  `json:"agent_id" gorm:"primaryKey"`
	CpuWarn  float64 `json:"cpu_warn"`
	CpuErr   float64 `json:"cpu_err"`
	RamWarn  float64 `json:"ram_warn"`
	RamErr   float64 `json:"ram_err"`
	DiskWarn float64 `json:"disk_warn"`
	DiskErr  float64 `json:"disk_err"`
}
