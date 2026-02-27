package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func main() {
	http.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("Received request: %s\n", string(body))
		
		// Set headers
		w.Header().Set("Content-Type", "application/json")

		// Mock standard response (non-streaming)
		// We'll just echo back "Processed: <input>"
		
		// Extract user content for realism
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		json.Unmarshal(body, &req)
		
		userContent := "unknown"
		if len(req.Messages) > 0 {
			userContent = req.Messages[0].Content
		}

		// Simulate processing time
		time.Sleep(500 * time.Millisecond)

		resp := ChatCompletionResponse{
			ID:      "chatcmpl-mock",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "mock-gpt",
		}
		resp.Choices = append(resp.Choices, struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			Index: 0,
			Message: struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    "assistant",
				Content: "[Processed] " + userContent,
			},
			FinishReason: "stop",
		})

		json.NewEncoder(w).Encode(resp)
	})

	fmt.Println("Mock LLM Server running on :8081 (Non-Streaming Mode)")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
