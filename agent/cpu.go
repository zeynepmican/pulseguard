package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"time"
)

type CPUChecker struct {
	config BaseCheckConfig
	hostID string
}

// constructor function for CPUChecker
func NewCPUChecker(cfg BaseCheckConfig, hostID string) *CPUChecker {
	return &CPUChecker{
		config: cfg,
		hostID: hostID,
	}
}

func (c *CPUChecker) Execute() (*Event, error) {

	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("cpu değeri okunamadı: %v", err)
	}

	cpuUsage := percentages[0]

	var level string
	var message string

	if cpuUsage >= c.config.ThresholdError {
		level = "ERROR"
		message = fmt.Sprintf("Kritik: CPU kullanımı %%%.2f seviyesine ulaştı!", cpuUsage)
	} else if cpuUsage >= c.config.ThresholdWarning {
		level = "WARNING"
		message = fmt.Sprintf("Uyarı: CPU kullanımı %%%.2f seviyesine ulaştı.", cpuUsage)
	} else {
		// herhangi bir sınırı geçmiyorsa Event oluşturulmaz
		return nil, nil
	}

	event := &Event{
		EventID:   fmt.Sprintf("cpu-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		CheckType: "cpu",
		Level:     level,
		Message:   message,
		Value:     cpuUsage,
	}

	return event, nil

}
