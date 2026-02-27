package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"shadowmixer/config"
)

// StreamMessage structure matching the worker's definition
type StreamMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type TaskPayload struct {
	ID   string          `json:"id"`
	Body json.RawMessage `json:"body"`
}

func SetupRouter(rdb *redis.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Generic proxy handler
	r.POST("/*any", func(c *gin.Context) {
		// 1. Identity Stripping (Stateless)
		// We read the body and ignore client headers/IP
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
			return
		}

		// Generate a simple unique ID
		taskID := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

		// Prepare task
		task := TaskPayload{
			ID:   taskID,
			Body: json.RawMessage(body),
		}
		taskJSON, _ := json.Marshal(task)

		// Subscribe to response channel BEFORE pushing to queue
		// This ensures we don't miss the response if worker is super fast (unlikely with jitter)
		ctx := context.Background()
		pubsub := rdb.Subscribe(ctx, "response:"+taskID)
		defer pubsub.Close()

		// 2. Queue (Push to Redis List)
		if err := rdb.RPush(ctx, "llm_queue", taskJSON).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue request"})
			return
		}

		// Determine Content-Type based on client request
		// If client asks for event-stream, we give it. Otherwise default to json.
		if c.GetHeader("Accept") == "text/event-stream" {
			c.Header("Content-Type", "text/event-stream")
		} else {
			c.Header("Content-Type", "application/json")
		}
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		// 3. Wait for Transparent Return (Streaming)
		c.Stream(func(w io.Writer) bool {
			select {
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return false // Channel closed
				}

				var streamMsg StreamMessage
				if err := json.Unmarshal([]byte(msg.Payload), &streamMsg); err != nil {
					// Fallback: treat as raw string if unmarshal fails (unlikely)
					return true
				}

				switch streamMsg.Type {
				case "chunk":
					c.Writer.Write([]byte(streamMsg.Payload))
					c.Writer.Flush()
					return true
				case "error":
					// If we haven't sent headers yet, we could change status code,
					// but here we are streaming, so we just append error text.
					// In a better impl, we might wrap error in JSON.
					errMsg := fmt.Sprintf(`{"error": "%s"}`, streamMsg.Payload)
					c.Writer.Write([]byte(errMsg))
					return false
				case "done":
					return false
				default:
					return true
				}
			case <-time.After(300 * time.Second):
				// Timeout
				return false
			case <-c.Request.Context().Done():
				// Client disconnected
				return false
			}
		})
	})

	return r
}
