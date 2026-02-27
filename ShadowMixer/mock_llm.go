package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request")
		
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Mock streaming response
		chunks := []string{
			"Hello", " ", "world", "!", " This", " is", " a", " test", " stream.",
		}

		for _, chunk := range chunks {
			// Simulate processing time
			time.Sleep(200 * time.Millisecond)
			
			// Send chunk in OpenAI format (optional, but good for realism)
			// For simplicity, just send raw text first to verify our pipe
			// fmt.Fprintf(w, "data: %s\n\n", chunk)
			
			// Actually, let's just send raw text since our worker reads raw bytes
			// If we want to simulate OpenAI fully, we'd send SSE events.
			// But our worker just blindly forwards whatever it reads.
			// So if we send raw text here, the client gets raw text chunks.
			fmt.Fprintf(w, "%s", chunk)
			flusher.Flush()
		}
	})

	fmt.Println("Mock LLM Server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
