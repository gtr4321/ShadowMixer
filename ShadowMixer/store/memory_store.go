package store

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MemoryStore struct {
	queues map[string]chan string
	hashes map[string]map[string]string
	meta   map[string]string
	
	qLock sync.Mutex
	hLock sync.RWMutex
	mLock sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		queues: make(map[string]chan string),
		hashes: make(map[string]map[string]string),
		meta:   make(map[string]string),
	}
}

func (s *MemoryStore) getQueue(name string) chan string {
	s.qLock.Lock()
	defer s.qLock.Unlock()
	if _, ok := s.queues[name]; !ok {
		// Create a buffered channel to act as a queue
		s.queues[name] = make(chan string, 1000)
	}
	return s.queues[name]
}

func (s *MemoryStore) PushQueue(ctx context.Context, queueName string, message string) error {
	q := s.getQueue(queueName)
	select {
	case q <- message:
		return nil
	default:
		// Try to resize or just return error. For memory store, we can use a larger buffer.
		return fmt.Errorf("queue %s is full", queueName)
	}
}

func (s *MemoryStore) PopQueue(ctx context.Context, queueName string, timeout time.Duration) (string, error) {
	q := s.getQueue(queueName)
	
	// Create a channel for timeout signal
	var timeoutCh <-chan time.Time
	if timeout > 0 {
		timeoutCh = time.After(timeout)
	} else {
		// If timeout is 0, we block indefinitely (or until context cancel)
		// but time.After(0) returns immediately, which is not what we want for BLPop(0).
		// We just leave timeoutCh nil to block forever.
	}

	select {
	case msg := <-q:
		return msg, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timeoutCh:
		return "", fmt.Errorf("redis: nil") // Simulate Redis Nil error on timeout
	}
}

func (s *MemoryStore) SaveResult(ctx context.Context, hashKey string, field string, value string, ttl time.Duration) error {
	s.hLock.Lock()
	defer s.hLock.Unlock()
	
	if _, ok := s.hashes[hashKey]; !ok {
		s.hashes[hashKey] = make(map[string]string)
	}
	s.hashes[hashKey][field] = value
	
	// Note: MemoryStore doesn't implement TTL cleanup for simplicity in this MVP.
	return nil
}

func (s *MemoryStore) GetResults(ctx context.Context, hashKey string) (map[string]string, error) {
	s.hLock.RLock()
	defer s.hLock.RUnlock()
	
	// Return a copy to be safe
	res := make(map[string]string)
	if original, ok := s.hashes[hashKey]; ok {
		for k, v := range original {
			res[k] = v
		}
	}
	return res, nil
}

func (s *MemoryStore) SetMeta(ctx context.Context, key string, value string, ttl time.Duration) error {
	s.mLock.Lock()
	defer s.mLock.Unlock()
	s.meta[key] = value
	return nil
}

func (s *MemoryStore) GetMeta(ctx context.Context, key string) (string, error) {
	s.mLock.RLock()
	defer s.mLock.RUnlock()
	val, ok := s.meta[key]
	if !ok {
		return "", fmt.Errorf("key not found")
	}
	return val, nil
}

func (s *MemoryStore) Close() error {
	return nil
}
