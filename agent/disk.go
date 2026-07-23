package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/disk"
)

type DiskChecker struct {
	// we use DiskCheckConfig
	config DiskCheckConfig
	hostID string
}

func NewDiskChecker(cfg DiskCheckConfig, hostID string) *DiskChecker {
	return &DiskChecker{
		config: cfg,
		hostID: hostID,
	}
}

func (d *DiskChecker) Execute() (*Event, error) {
	// Gets the disk usage statistics for the specified path using the gopsutil library. If there's an error, it returns nil and the error.
	usage, err := disk.Usage(d.config.Path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read disk usage (%s): %v", d.config.Path, err)
	}

	diskUsage := usage.UsedPercent

	var level string
	var message string

	if diskUsage >= d.config.ThresholdError {
		level = "ERROR"
		message = fmt.Sprintf("Critical: Disk (%s) usage reached %.2f%%!", d.config.Path, diskUsage)
	} else if diskUsage >= d.config.ThresholdWarning {
		level = "WARNING"
		message = fmt.Sprintf("Warning: Disk (%s) usage reached %.2f%%!", d.config.Path, diskUsage)
	} else {
		return nil, nil
	}

	event := &Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		CheckType: "disk",
		Level:     level,
		Message:   message,
		Value:     diskUsage,
	}

	return event, nil
}
