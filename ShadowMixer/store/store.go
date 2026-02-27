package store

import (
	"context"
	"time"
)

// Store defines the interface for data persistence and queueing
type Store interface {
	// PushQueue pushes a message to the queue
	PushQueue(ctx context.Context, queueName string, message string) error

	// PopQueue pops a message from the queue (blocking or non-blocking)
	// For MemoryStore, we might implement blocking via channels
	PopQueue(ctx context.Context, queueName string, timeout time.Duration) (string, error)

	// SaveResult stores a key-value pair in a hash map
	SaveResult(ctx context.Context, hashKey string, field string, value string, ttl time.Duration) error

	// GetResults retrieves all fields and values from a hash map
	GetResults(ctx context.Context, hashKey string) (map[string]string, error)

	// SetMeta stores simple key-value metadata
	SetMeta(ctx context.Context, key string, value string, ttl time.Duration) error

	// GetMeta retrieves metadata
	GetMeta(ctx context.Context, key string) (string, error)
	
	// Close cleans up resources
	Close() error
}
