package main

import (
	"encoding/json"

	"fmt"
	"log"
	"time"

	"bytes"
	"net/http"
)

func main() {
	// Read the configuration file
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Critical error while loading config: %v", err)
	}

	// Open or create the local queue database
	localQueue, err := NewLocalQueue("pulseguard_queue.db")
	if err != nil {
		log.Fatalf("Database could not be opened: %v", err)
	}

	defer localQueue.db.Close() // defer: Release the database lock when the agent stops or crashes.

	// for testing
	// localQueue.PrintAll()

	// Create a list for active checkers based on the configuration. If a check is enabled, we create a new checker and add it to the list.
	var activeCheckers []Checker
	if config.Checks.CPU.Enabled {
		activeCheckers = append(activeCheckers, NewCPUChecker(config.Checks.CPU))
	}
	if config.Checks.RAM.Enabled {
		activeCheckers = append(activeCheckers, NewRAMChecker(config.Checks.RAM))
	}
	if config.Checks.Disk.Enabled {
		activeCheckers = append(activeCheckers, NewDiskChecker(config.Checks.Disk))
	}
	// no checker enabled, exit
	if len(activeCheckers) == 0 {
		log.Fatal("There are no enabled checks in the configuration. Exiting.")
	}

	// Create a channel to hold events.
	eventQueue := make(chan *Event, 100)

	// eventShipper listens to the eventQueue channel and writes events to the local queue (BoltDB).
	go eventShipper(eventQueue, localQueue)

	// batchProcessor simulates sending events to the Collector in batches.
	go batchProcessor(localQueue, config.Agent.SecretKey, config.Collector.URL, config.Agent.HostID)

	// ticker is a channel that sends a signal at regular intervals.
	// The main loop waits for the ticker to send a signal, and when it does, it runs the checks and sends any events to the eventQueue channel.
	// it helps to control the frequency of checks and avoid overloading the system with too many checks in a short period of time.
	ticker := time.NewTicker(time.Duration(config.Agent.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		runChecks(activeCheckers, eventQueue)
	}
}

// eventShipper function listens to the eventQueue channel and writes events to the local queue (BoltDB).
func eventShipper(queueChannel <-chan *Event, localQueue *LocalQueue) {

	for event := range queueChannel {
		// We retry writing to the local queue up to 3 times if there is an error.
		// If all retries fail, we log a critical error and move on to the next event.
		maxRetries := 3

		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := localQueue.Enqueue(event)
			if err == nil {

				//fmt.Printf("   ADDED TO VAULT | %s [%s]: %.2f\n", event.CheckType, event.Level, event.Value)
				break // quit retry loop if successful
			}
			log.Printf("DISK WRITE ERROR (Attempt %d/%d) - EventID: %s: %v\n", attempt, maxRetries, event.EventID, err)

			if attempt < maxRetries {
				// Wait for 2 seconds before retrying (Backoff)
				time.Sleep(2 * time.Second)
			} else {
				log.Printf("CRITICAL: Event %s could not be saved to disk and is LOST!\n", event.EventID)
			}
		}

	}
}

// batchProcessor simulates sending events to the Collector in batches.
func batchProcessor(dbQueue *LocalQueue, secretKey string, collectorUrl string, agentId string) {
	// Create an HTTP client with a timeout of 5 seconds. This client will be used to send the batch of events to the Collector.
	// Timeout is important to avoid hanging the agent if the Collector is not responding.
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		// Read a batch of events from the local queue
		batchSize := 5
		events, err := dbQueue.GetBatch(batchSize)

		// If there is an error or if there are no events in the batch, we wait for 5 seconds before trying again.
		if err != nil || len(events) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		// If there are events in the batch, we simulate sending them to the Collector.
		fmt.Println("\n--------------------------------------------------")
		fmt.Printf("Kargo Kamyonu Geldi! %d adet paket Toplayıcıya gönderiliyor...\n", len(events))

		var eventIDs []string
		for _, ev := range events {
			fmt.Printf("-> Giden: %s | %s\n", ev.EventID, ev.CheckType)
			eventIDs = append(eventIDs, ev.EventID) // We collect the EventIDs of the events we are sending so we can delete them from the local queue after a successful send.
		}

		// Serialize the batch of events to JSON format.
		batchJSON, err := json.Marshal(events)
		if err != nil {
			fmt.Printf("JSON MARSHAL ERROR: %v\n", err)
			continue
		}
		// Generate the HMAC-SHA256 signature for the batch using the secret key.
		signature := GenerateSignature(batchJSON, secretKey)

		// NEW HTTP REQUEST STARTING

		// New POST request to the Collector with the batch of events in the body.
		req, err := http.NewRequest("POST", collectorUrl, bytes.NewBuffer(batchJSON))
		if err != nil {
			log.Printf("REQUEST OLUŞTURMA HATASI: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Add the headers.
		req.Header.Set("X-Agent-ID", agentId)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Signature", signature)
		// req.Header.Set("X-Agent-ID", events[0].HostID) // Ekstra bilgi

		// Send the request and get the response
		resp, err := client.Do(req)

		// If there is an error, we log it and wait for 5 seconds before trying again.
		// This could be a network error, timeout, etc.
		if err != nil {
			log.Printf("AĞ HATASI (Sunucuya ulaşılamadı): %v\n", err)
			fmt.Println("Paketler kasada tutuluyor, sonra tekrar denenecek...")
			time.Sleep(5 * time.Second)
			continue
		}
		if resp.StatusCode == http.StatusOK { // 200 OK
			// Delete the successfully sent events.
			err = dbQueue.DeleteBatch(eventIDs)
			if err != nil {
				log.Printf("BATCH DELETE ERROR (Veriler hala kasada): %v\n", err)
			} else {
				fmt.Printf("Başarıyla gönderilen %d paket kasadan silindi.\n", len(events))
			}
		} else {
			// 500 Server Error, 401 Unauthorized
			log.Printf("SUNUCU HATASI DÖNDÜ (Kod: %d). Veriler silinmedi, tekrar denenecek.\n", resp.StatusCode)
		}

		// Close the response body to free up resources. This is important to avoid memory leaks.
		resp.Body.Close()

		fmt.Println("--------------------------------------------------")

		// Rest for 5 sec before sending the next batch.
		time.Sleep(5 * time.Second)
	}
}

// runChecks function iterates over the list of checkers, executes each one, and sends any resulting events to the eventQueue channel.
func runChecks(checkers []Checker, queue chan<- *Event) {
	fmt.Printf("\n[%s] Checking systems...\n", time.Now().Format("15:04:05"))
	for _, worker := range checkers {
		event, err := worker.Execute()

		if err != nil {
			log.Printf("READ ERROR: %v\n", err)
			continue
		}
		if event != nil {
			queue <- event
		}
	}
}
