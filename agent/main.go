package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Critical error while loading config: %v", err)
	}

	localQueue, err := NewLocalQueue("pulseguard_queue.db")
	if err != nil {
		log.Fatalf("Database could not be opened: %v", err)
	}

	// defer: Release the database lock when the agent (main function) stops or crashes.
	// This prevents file corruption.
	defer localQueue.db.Close()

	// for testing
	// localQueue.PrintAll()

	// list for active checkers based on the configuration. If a check is enabled, we create a new checker and add it to the list.
	var activeCheckers []Checker
	if config.Checks.CPU.Enabled {
		activeCheckers = append(activeCheckers, NewCPUChecker(config.Checks.CPU, config.Agent.HostID))
	}
	if config.Checks.RAM.Enabled {
		activeCheckers = append(activeCheckers, NewRAMChecker(config.Checks.RAM, config.Agent.HostID))
	}
	if config.Checks.Disk.Enabled {
		activeCheckers = append(activeCheckers, NewDiskChecker(config.Checks.Disk, config.Agent.HostID))
	}
	// no checker enabled, exit
	if len(activeCheckers) == 0 {
		log.Fatal("There are no enabled checks in the configuration. Exiting.")
	}

	// Create a channel to hold events.
	eventQueue := make(chan *Event, 100)

	// Pass both the channel and the database instance to the shipper.
	go eventShipper(eventQueue, localQueue)

	ticker := time.NewTicker(time.Duration(config.Agent.IntervalSeconds) * time.Second)
	defer ticker.Stop()
	// ticker is a channel that sends a signal at regular intervals.
	// The main loop waits for the ticker to send a signal, and when it does, it runs the checks and sends any events to the eventQueue channel.
	// it helps to control the frequency of checks and avoid overloading the system with too many checks in a short period of time.
	for {
		<-ticker.C
		runChecks(activeCheckers, eventQueue)
	}
}

// runChecks SENSOR LOOP (Pushes data to the channel asynchronously
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

// 2nd goroutine
// eventShipper function listens to the eventQueue channel and writes events to the local queue (BoltDB).
func eventShipper(queueChannel <-chan *Event, localQueue *LocalQueue) {
	// range queue: Kanal açık olduğu sürece döngü burada bekler.
	// Kanala yeni bir veri düştüğü anda uyanır, veriyi alır ve işler.
	for event := range queueChannel {
		// we retry writing to the local queue up to 3 times if there is an error. If all retries fail, we log a critical error and move on to the next event.
		maxRetries := 3

		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := localQueue.Enqueue(event)
			if err == nil {

				fmt.Printf("   ADDED TO VAULT | %s [%s]: %.2f\n", event.CheckType, event.Level, event.Value)
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
