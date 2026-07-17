package main

import "time"

type Event struct {
	EventID   string    `json:"event_id"`     // Events unique identifier
	Timestamp time.Time `json:"timestamp"`    // Timestamp of the event
	CheckType string    `json:"check_type"`   // Which metric had an issue? (cpu, ram, disk)
	Level     string    `json:"level"`        // Level: INFO, WARNING, ERROR
	Message   string    `json:"message"`      // Human-readable description
	Value     float64   `json:"metric_value"` // Current problematic metric value
}

// Checker is an interface that defines the behavior of a health check.
// Each specific check (like CPU, RAM, Disk) will implement this interface.
type Checker interface {
	// If the threshold is exceeded, Execute returns an Event. If not, it returns nil.
	Execute() (*Event, error)
}
