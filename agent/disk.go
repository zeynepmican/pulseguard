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
}

func NewDiskChecker(cfg DiskCheckConfig) *DiskChecker {
	return &DiskChecker{
		config: cfg,
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

	level = "NORMAL"

	if diskUsage >= d.config.ThresholdError {
		level = "ERROR"
	} else if diskUsage >= d.config.ThresholdWarning {
		level = "WARNING"
	}

	event := &Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		CheckType: "disk",
		Level:     level,
		Value:     diskUsage,
	}

	return event, nil
}
