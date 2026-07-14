package main

import (
	"fmt"
	"time"

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
	// Hangi diski okuyacağını config dosyasındaki 'Path' değerinden alıyor ("/" veya "C:\" gibi)
	usage, err := disk.Usage(d.config.Path)
	if err != nil {
		return nil, fmt.Errorf("disk değeri okunamadı (%s): %v", d.config.Path, err)
	}

	diskUsage := usage.UsedPercent

	var level string
	var message string

	if diskUsage >= d.config.ThresholdError {
		level = "ERROR"
		message = fmt.Sprintf("Kritik: Disk (%s) kullanımı %%%.2f seviyesine ulaştı!", d.config.Path, diskUsage)
	} else if diskUsage >= d.config.ThresholdWarning {
		level = "WARNING"
		message = fmt.Sprintf("Uyarı: Disk (%s) kullanımı %%%.2f seviyesine ulaştı.", d.config.Path, diskUsage)
	} else {
		return nil, nil
	}

	event := &Event{
		EventID:   fmt.Sprintf("disk-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		CheckType: "disk",
		Level:     level,
		Message:   message,
		Value:     diskUsage,
	}

	return event, nil
}
