package main

import "time"

type Event struct {
	ID      uint   `gorm:"primaryKey" json:"-"`
	EventID string `gorm:"uniqueIndex" json:"event_id"`
	// EventID   string    `json:"event_id"`     // Event's unique identifier
	Timestamp time.Time `json:"timestamp"`    // Timestamp of the event
	CheckType string    `json:"check_type"`   // Which metric had an issue? (cpu, ram, disk)
	Level     string    `json:"level"`        // Level: INFO, WARNING, ERROR
	Message   string    `json:"message"`      // Human-readable description
	Value     float64   `json:"metric_value"` // Current problematic metric value
}
