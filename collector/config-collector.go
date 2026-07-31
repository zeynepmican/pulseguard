package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Security struct {
		SecretKey string `yaml:"secret_key"`
	} `yaml:"security"`
}

var appConfig Config

func loadConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config.yaml file: %v", err)
	}

	err = yaml.Unmarshal(data, &appConfig)
	if err != nil {
		log.Fatalf("config.yaml format error: %v", err)
	}
}
