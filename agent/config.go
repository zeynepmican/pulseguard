package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent     AgentConfig     `yaml:"agent"`
	Collector CollectorConfig `yaml:"collector"`
	Checks    ChecksConfig    `yaml:"checks"`
}

type AgentConfig struct {
	HostID          string `yaml:"host_id"`
	IntervalSeconds int    `yaml:"interval_seconds"`
}

type CollectorConfig struct {
	URL string `yaml:"url"`
}

type ChecksConfig struct {
	CPU  BaseCheckConfig `yaml:"cpu"`
	RAM  BaseCheckConfig `yaml:"ram"`
	Disk DiskCheckConfig `yaml:"disk"`
}

// for CPU and RAM checks
// ikisi de aynı kontrol yapılandırmasına sahip olduğundan BaseCheckConfig adı altında ikisini de tutuyoruz.
type BaseCheckConfig struct {
	Enabled          bool    `yaml:"enabled"`
	ThresholdWarning float64 `yaml:"threshold_warning"`
	ThresholdError   float64 `yaml:"threshold_error"`
}

// disk CPU ve RAM'in sahip olduğu kontrollere ek olarak path alanına sahip olduğundan ayrı bir yapılandırma olarak tanımlıyoruz.
type DiskCheckConfig struct {
	BaseCheckConfig `yaml:",inline"` // inline etiketi ile BaseCheckConfig alanını gömülü olarak ekliyoruz.
	Path            string           `yaml:"path"`
}

// public function to load the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		// if file reading fails, return the error
		return nil, err
	}
	var config Config

	err = yaml.Unmarshal(fileContent, &config) // configin adresini yolluyoruz ki configi doldurabilsin, kopyasını yaratmasın.
	if err != nil {
		// eğer dosya bozuk yaml formatında ise, unmarshal işlemi başarısız olur ve hata döneriz.
		return nil, err
	}
	return &config, nil
}
