package main

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

// for encapsulation, we will use a struct to represent the local queue
type LocalQueue struct {
	db *bbolt.DB
}

func NewLocalQueue(dbPath string) (*LocalQueue, error) {
	// if the database file does not exist, it will be created automatically
	// 0600 is the file permission for the database file, which means read and write permissions for the agent only
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("Database could not be opened: %v", err)
	}

	// everytime we write to the database, we open an update block to guarantee that the process will not be interrupted.
	// we create a bucket named "events" to store the events. If it already exists, we do nothing.
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("events"))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("Bucket could not be created: %v", err)
	}

	return &LocalQueue{db: db}, nil
}

// Enqueue function adds an event to the local queue. It serializes the event to JSON and stores it in the BoltDB database.
// LocalQueue'ya ait bir function.
func (q *LocalQueue) Enqueue(event *Event) error {
	return q.db.Update(func(tx *bbolt.Tx) error {
		// gets the events bucket from the database. If it doesn't exist, it returns an error.
		bucket := tx.Bucket([]byte("events"))

		// Serialize the event to JSON format.
		eventData, err := json.Marshal(event)
		if err != nil {
			return err
		}

		// Store the serialized event in the bucket with the EventID as the key.
		return bucket.Put([]byte(event.EventID), eventData)
	})
}

// PrintAll, reads all events from the local queue and prints them to the console.
// It is useful for debugging and testing purposes.
func (q *LocalQueue) PrintAll() {
	fmt.Println("\n---  READING ALL DATA ---")

	q.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("events"))

		cursor := bucket.Cursor()
		count := 0

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			fmt.Printf("Key: %s\nValue: %s\n\n", k, v)
			count++
		}

		fmt.Printf("--- TOTAL %d RECORDS FOUND ---\n\n", count)
		return nil
	})
}
