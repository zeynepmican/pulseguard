package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	loadConfig()

	http.HandleFunc("/api/events", handleEvents)

	portStr := fmt.Sprintf(":%d", appConfig.Server.Port)
	fmt.Printf("Collector server is listening on port %s...\n", portStr)

	err := http.ListenAndServe(portStr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
