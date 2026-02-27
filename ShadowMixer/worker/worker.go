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
	"shadowmixer/store"
	"sync"
	"time"
)

// Fragment definition (same as Router)
type Fragment struct {
	BigTaskID  string `json:"big_task_id"`
	SequenceID int    `json:"sequence_id"`
	Total      int    `json:"total"`
	Content    string `json:"content"`
	Model      string `json:"model"`
}

type Worker struct {
	store    store.Store
	cfg      *config.Config
	keyIndex int
	mu       sync.Mutex
}

func New(s store.Store, cfg *config.Config) *Worker {
	return &Worker{
		store: s,
		cfg:   cfg,
	}
}

// Start consumes tasks from Redis queue
func (w *Worker) Start() {
	fmt.Println("[Worker] Started. Waiting for fragments in 'llm_fragment_queue'...")
	for {
		// 1. Blocking Pop from Queue
		// BLPop(0) means block indefinitely.
		rawPayload, err := w.store.PopQueue(context.Background(), "llm_fragment_queue", 0)
		if err != nil {
			// If error is timeout or nil (should not happen with 0), just continue
			// If connection error, sleep
			fmt.Printf("[Worker] Queue pop error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// result is the payload
		go w.process(rawPayload)
	}
}

func (w *Worker) process(rawPayload string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[Worker] Panic recovered: %v\n", r)
		}
	}()

	var frag Fragment
	if err := json.Unmarshal([]byte(rawPayload), &frag); err != nil {
		fmt.Printf("[Worker] Invalid JSON: %v\n", err)
		return
	}

	fmt.Printf("[Worker] Processing Fragment %s-%d\n", frag.BigTaskID, frag.SequenceID)

	// 2. Jitter (0.1s - 2.0s random delay)
	delayMs := 100 + rand.Intn(1900)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// 3. Key Pooling (Round-Robin)
	keys := w.cfg.LLM.APIKeys
	if len(keys) == 0 {
		w.saveError(frag, "No API keys configured")
		return
	}
	
	w.mu.Lock()
	w.keyIndex = (w.keyIndex + 1) % len(keys)
	apiKey := keys[w.keyIndex]
	w.mu.Unlock()

	// 4. Request External API (Stateless)
	// We need to construct a valid LLM request for this fragment.
	// For simplicity, we treat the fragment content as the user prompt.
	llmReqBody := map[string]interface{}{
		"model": frag.Model,
		"messages": []map[string]string{
			{"role": "user", "content": frag.Content},
		},
	}
	bodyBytes, _ := json.Marshal(llmReqBody)

	// Retry loop for transient errors
	var resp *http.Response
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", w.cfg.LLM.Target, bytes.NewReader(bodyBytes))
		if err != nil {
			w.saveError(frag, "Failed to create request")
			return
		}

		// Set minimal headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		// We use non-streaming here to simplify the aggregator logic for MVP,
		// but we could stream to DB if needed.
		// req.Header.Set("Accept", "application/json") 

		client := &http.Client{Timeout: 300 * time.Second}
		resp, err = client.Do(req)
		if err == nil {
			break
		}

		fmt.Printf("[Worker] Attempt %d/%d failed: %v. Retrying...\n", i+1, maxRetries, err)
		time.Sleep(time.Duration(1<<i) * time.Second)
	}

	if err != nil {
		w.saveError(frag, fmt.Sprintf("Upstream error: %v", err))
		return
	}
	defer resp.Body.Close()

	// 5. Store Result
	// We read the full response and extract the content.
	respBody, _ := io.ReadAll(resp.Body)
	
	// Parse LLM response to get just the content text
	var llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	// Note: If upstream returns error JSON, unmarshal might partially fail or succeed.
	// We should check status code ideally.
	if resp.StatusCode != 200 {
		w.saveError(frag, fmt.Sprintf("Upstream status %d: %s", resp.StatusCode, string(respBody)))
		return
	}

	if err := json.Unmarshal(respBody, &llmResp); err != nil {
		// Fallback: save raw body if parsing fails
		w.saveResult(frag, string(respBody))
	} else if len(llmResp.Choices) > 0 {
		w.saveResult(frag, llmResp.Choices[0].Message.Content)
	} else {
		w.saveResult(frag, "")
	}

	fmt.Printf("[Worker] Fragment %s-%d completed\n", frag.BigTaskID, frag.SequenceID)
}

func (w *Worker) saveResult(frag Fragment, content string) {
	ctx := context.Background()
	key := "results:" + frag.BigTaskID
	// Store in Hash: field=SequenceID, value=Content
	w.store.SaveResult(ctx, key, fmt.Sprintf("%d", frag.SequenceID), content, 24*time.Hour)
}

func (w *Worker) saveError(frag Fragment, errMsg string) {
	fmt.Printf("[Worker] Error on %s-%d: %s\n", frag.BigTaskID, frag.SequenceID, errMsg)
	w.saveResult(frag, "[Error: "+errMsg+"]")
}
