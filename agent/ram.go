package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/mem"
)

type RAMChecker struct {
	config BaseCheckConfig
	hostID string
}

func NewRAMChecker(cfg BaseCheckConfig, hostID string) *RAMChecker {
	return &RAMChecker{
		config: cfg,
		hostID: hostID,
	}
}

func (r *RAMChecker) Execute() (*Event, error) {

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("Failed to read RAM usage: %v", err)
	}

	ramUsage := vMem.UsedPercent

	var level string
	var message string

	if ramUsage >= r.config.ThresholdError {
		level = "ERROR"
		message = fmt.Sprintf("Critical: RAM usage reached %.2f%%!", ramUsage)
	} else if ramUsage >= r.config.ThresholdWarning {
		level = "WARNING"
		message = fmt.Sprintf("Warning: RAM usage reached %.2f%%!", ramUsage)
	} else {
		return nil, nil
	}

	event := &Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		CheckType: "ram",
		Level:     level,
		Message:   message,
		Value:     ramUsage,
	}

	return event, nil
}
