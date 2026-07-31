package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	loadConfig()
	initDB()

	// Mevcut rotalarımızı CORS ara katmanı ile sarıyoruz
	http.HandleFunc("/api/events", corsMiddleware(eventsHandler))
	http.HandleFunc("/api/hosts", corsMiddleware(handleGetHosts))

	http.HandleFunc("/api/hosts/", corsMiddleware(handleGetHostDetails))

	http.HandleFunc("/api/configs/", corsMiddleware(handleAgentConfig))

	portStr := fmt.Sprintf(":%d", appConfig.Server.Port)
	fmt.Printf("Collector server is listening on port %s...\n", portStr)

	err := http.ListenAndServe(portStr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Tarayıcıların CORS (Güvenlik) engelini aşmamızı sağlayan ara katman
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Printf("KAPI ÇALINDI -> Method: %s | URL: %s\n", r.Method, r.URL.Path)
		// Güvenlik kilidini açan başlıklar (Headers)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Geliştirme aşamasında herkese açık (*)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Agent-ID, X-Signature")

		// Tarayıcı asıl isteği atmadan önce "Ön Uçuş" (OPTIONS) isteği atar.
		// Eğer istek OPTIONS ise, fonksiyona hiç gitmeden "Her şey yolunda" (200 OK) dönüyoruz.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Her şey tamamsa, isteği asıl işleve (handler) yolla
		next(w, r)
	}
}
