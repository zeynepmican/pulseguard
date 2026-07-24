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
	var message string

	if cpuUsage >= c.config.ThresholdError {
		level = "ERROR"
		message = fmt.Sprintf("Critical: CPU usage reached %.2f%%!", cpuUsage)
	} else if cpuUsage >= c.config.ThresholdWarning {
		level = "WARNING"
		message = fmt.Sprintf("Warning: CPU usage reached %.2f%%!", cpuUsage)
	} else {
		// if the percentage is below the warning threshold, we can return nil to indicate no event
		return nil, nil
	}

	event := &Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		CheckType: "cpu",
		Level:     level,
		Message:   message,
		Value:     cpuUsage,
	}

	return event, nil

}
