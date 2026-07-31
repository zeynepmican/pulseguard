package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
)

type CPUChecker struct {
	config BaseCheckConfig
}

// constructor function for CPUChecker
func NewCPUChecker(cfg BaseCheckConfig) *CPUChecker {
	return &CPUChecker{
		config: cfg,
	}
}

func (c *CPUChecker) Execute() (*Event, error) {

	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to read CPU usage: %v", err)
	}

	cpuUsage := percentages[0]

	var level string

	level = "NORMAL"
	if cpuUsage >= c.config.ThresholdError {
		level = "ERROR"
	} else if cpuUsage >= c.config.ThresholdWarning {
		level = "WARNING"

	} else {

	}

	event := &Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		CheckType: "cpu",
		Level:     level,
		Value:     cpuUsage,
	}

	return event, nil

}
