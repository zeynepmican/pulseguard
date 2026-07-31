package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/mem"
)

type RAMChecker struct {
	config BaseCheckConfig
}

func NewRAMChecker(cfg BaseCheckConfig) *RAMChecker {
	return &RAMChecker{
		config: cfg,
	}
}

func (r *RAMChecker) Execute() (*Event, error) {

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("Failed to read RAM usage: %v", err)
	}

	ramUsage := vMem.UsedPercent

	var level string
	level = "NORMAL"

	if ramUsage >= r.config.ThresholdError {
		level = "ERROR"
	} else if ramUsage >= r.config.ThresholdWarning {
		level = "WARNING"
	}

	event := &Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		CheckType: "ram",
		Level:     level,
		Value:     ramUsage,
	}

	return event, nil
}
