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
	SecretKey       string `yaml:"secret_key"`
}

type CollectorConfig struct {
	URL string `yaml:"url"`
}

type ChecksConfig struct {
	CPU  BaseCheckConfig `yaml:"cpu"`
	RAM  BaseCheckConfig `yaml:"ram"`
	Disk DiskCheckConfig `yaml:"disk"`
}

// both CPU and RAM checks have the same configuration structure, so we define a base configuration struct for them.
type BaseCheckConfig struct {
	Enabled          bool    `yaml:"enabled"`
	ThresholdWarning float64 `yaml:"threshold_warning"`
	ThresholdError   float64 `yaml:"threshold_error"`
}

// disk check has a specific configuration structure, so we define a separate struct for it.
type DiskCheckConfig struct {
	BaseCheckConfig `yaml:",inline"` // with inline, the fields of BaseCheckConfig are embedded directly into DiskCheckConfig.
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

	err = yaml.Unmarshal(fileContent, &config) // we send the address of the config variable to Unmarshal so that it can populate it with the data from the YAML file.
	if err != nil {
		// if the file content cannot be unmarshaled into the Config struct, return the error
		return nil, err
	}
	return &config, nil
}
