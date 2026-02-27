package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"shadowmixer/config"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// Task payload
type Task struct {
	ID   string          `json:"id"`
	Body json.RawMessage `json:"body"`
}

// StreamMessage for Redis Pub/Sub
type StreamMessage struct {
	Type    string `json:"type"`              // "chunk", "error", "done"
	Payload string `json:"payload,omitempty"` // content or error message
}

type Worker struct {
	rdb      *redis.Client
	cfg      *config.Config
	keyIndex uint64
}

func New(rdb *redis.Client, cfg *config.Config) *Worker {
	return &Worker{
		rdb: rdb,
		cfg: cfg,
	}
}

// Start consumes tasks from Redis queue
func (w *Worker) Start() {
	fmt.Println("[Worker] Started. Waiting for tasks in 'llm_queue'...")
	for {
		// 1. Blocking Pop from Queue
		result, err := w.rdb.BLPop(context.Background(), 0, "llm_queue").Result()
		if err != nil {
			fmt.Printf("[Worker] Redis connection error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// result[1] contains the payload
		go w.process(result[1])
	}
}

func (w *Worker) process(rawPayload string) {
	var task Task
	if err := json.Unmarshal([]byte(rawPayload), &task); err != nil {
		fmt.Printf("[Worker] Invalid JSON: %v\n", err)
		return
	}

	fmt.Printf("[Worker] Processing Task %s\n", task.ID)

	// 2. Queue & Jitter (0.1s - 2.0s random delay)
	delayMs := 100 + rand.Intn(1900)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// 3. Key Pooling (Round-Robin)
	keys := w.cfg.LLM.APIKeys
	if len(keys) == 0 {
		w.publishError(task.ID, "No API keys configured")
		return
	}
	currentKeyIndex := atomic.AddUint64(&w.keyIndex, 1)
	apiKey := keys[currentKeyIndex%uint64(len(keys))]

	// 4. Request External API (Stateless)
	// Retry loop for transient errors
	var resp *http.Response
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		// We create a new request each time, effectively stripping original headers/IP
		req, err := http.NewRequest("POST", w.cfg.LLM.Target, bytes.NewReader(task.Body))
		if err != nil {
			w.publishError(task.ID, "Failed to create request")
			return
		}

		// Set minimal headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		// Suggest streaming response if upstream supports it (e.g. SSE)
		req.Header.Set("Accept", "text/event-stream")

		// Do request
		// Use a longer timeout or no timeout for streaming
		client := &http.Client{Timeout: 300 * time.Second}
		resp, err = client.Do(req)
		if err == nil {
			break
		}

		fmt.Printf("[Worker] Attempt %d/%d failed: %v. Retrying...\n", i+1, maxRetries, err)
		time.Sleep(time.Duration(1<<i) * time.Second) // Exponential backoff: 1s, 2s, 4s
	}

	if err != nil {
		w.publishError(task.ID, fmt.Sprintf("Upstream error after retries: %v", err))
		return
	}
	defer resp.Body.Close()

	// 5. Transparent Return (Streaming)
	// Read chunks and publish them immediately
	buf := make([]byte, 4096) // 4KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// Publish chunk
			msg := StreamMessage{
				Type:    "chunk",
				Payload: string(buf[:n]),
			}
			w.publish(task.ID, msg)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			// Log error but we might have already sent some data
			fmt.Printf("[Worker] Error reading response: %v\n", err)
			w.publishError(task.ID, "Error reading upstream stream")
			return
		}
	}

	// Send Done signal
	w.publish(task.ID, StreamMessage{Type: "done"})

	fmt.Printf("[Worker] Task %s completed (Delay: %dms, Key: ...%s)\n",
		task.ID, delayMs, apiKey[len(apiKey)-4:])
}

func (w *Worker) publish(taskID string, msg StreamMessage) {
	data, _ := json.Marshal(msg)
	w.rdb.Publish(context.Background(), "response:"+taskID, data)
}

func (w *Worker) publishError(taskID, errMsg string) {
	msg := StreamMessage{
		Type:    "error",
		Payload: errMsg,
	}
	w.publish(taskID, msg)
}
