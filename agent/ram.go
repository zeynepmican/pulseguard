package main

import (
	"fmt"
	"time"

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
		return nil, fmt.Errorf("ram değeri okunamadı: %v", err)
	}

	ramUsage := vMem.UsedPercent

	var level string
	var message string

	if ramUsage >= r.config.ThresholdError {
		level = "ERROR"
		message = fmt.Sprintf("Kritik: RAM kullanımı %%%.2f seviyesine ulaştı!", ramUsage)
	} else if ramUsage >= r.config.ThresholdWarning {
		level = "WARNING"
		message = fmt.Sprintf("Uyarı: RAM kullanımı %%%.2f seviyesine ulaştı.", ramUsage)
	} else {
		return nil, nil
	}

	event := &Event{
		EventID:   fmt.Sprintf("ram-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		CheckType: "ram",
		Level:     level,
		Message:   message,
		Value:     ramUsage,
	}

	return event, nil
}
