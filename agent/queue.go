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

// GetBatch function retrieves a batch of events from the local queue.
// It reads the events from the BoltDB database and deserializes them from JSON format.
func (q *LocalQueue) GetBatch(limit int) ([]*Event, error) {
	var events []*Event
	// We will read the events from the database in a read-only transaction.
	err := q.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("events"))
		if bucket == nil {
			return fmt.Errorf("Bucket not found")
		}

		cursor := bucket.Cursor()
		count := 0

		// Start from the first key-value pair in the bucket and iterate through the bucket until we reach the limit or there are no more items.
		for k, v := cursor.First(); k != nil && count < limit; k, v = cursor.Next() {
			var event Event

			// Deserialization of the event from JSON format.
			err := json.Unmarshal(v, &event)
			if err != nil {
				// If there is an error during deserialization, we log the error and continue to the next item in the bucket.
				continue
			}

			events = append(events, &event)
			count++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

// DeleteBatch function removes a batch of events from the local queue based on their EventIDs.
func (q *LocalQueue) DeleteBatch(eventIDs []string) error {
	return q.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("events"))
		if bucket == nil {
			return fmt.Errorf("Bucket not found")
		}

		// For each EventID in the provided list, we delete the corresponding key-value pair from the bucket.
		for _, id := range eventIDs {
			err := bucket.Delete([]byte(id))
			if err != nil {
				return fmt.Errorf("Failed to delete event %s: %v", id, err)
			}
		}
		return nil
	})
}
