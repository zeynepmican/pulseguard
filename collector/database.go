package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDB() {
	var err error
	// pulseguard.db adında bir dosya oluşturur ve bağlanır
	DB, err = gorm.Open(sqlite.Open("pulseguard.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// SİHİRLİ KISIM: Event struct'ına bakar ve SQL tablosunu otomatik oluşturur!
	err = DB.AutoMigrate(&Event{}, &AgentConfig{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully!")
}
