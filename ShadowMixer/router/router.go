package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"shadowmixer/config"
	"shadowmixer/store"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Fragment represents a sub-task of a larger request
type Fragment struct {
	BigTaskID  string `json:"big_task_id"`
	SequenceID int    `json:"sequence_id"`
	Total      int    `json:"total"`
	Content    string `json:"content"`
	Model      string `json:"model"`
}

type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

func SetupRouter(s store.Store, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Aggregator Endpoint: Check status of a BigTask
	r.GET("/v1/tasks/:id", func(c *gin.Context) {
		taskID := c.Param("id")
		ctx := context.Background()

		// Get all fragments from Redis Hash: results:taskID -> {seqID: result}
		results, err := s.GetResults(ctx, "results:"+taskID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
			return
		}

		if len(results) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"status": "pending", "message": "No results yet"})
			return
		}

		// We need to know total count. It's stored in a separate key or we infer it.
		// For simplicity, let's store metadata separately.
		totalStr, err := s.GetMeta(ctx, "meta:"+taskID+":total")
		if err != nil {
			// If meta is missing, we might still be processing
			c.JSON(http.StatusOK, gin.H{"status": "processing", "completed": len(results)})
			return
		}
		
		var total int
		fmt.Sscanf(totalStr, "%d", &total)

		if len(results) < total {
			c.JSON(http.StatusOK, gin.H{
				"status":    "processing",
				"completed": len(results),
				"total":     total,
			})
			return
		}

		// All done! Reassemble.
		type ResultFragment struct {
			SeqID   int
			Content string
		}
		var frags []ResultFragment
		for k, v := range results {
			var seq int
			fmt.Sscanf(k, "%d", &seq)
			frags = append(frags, ResultFragment{SeqID: seq, Content: v})
		}

		// Sort by sequence
		sort.Slice(frags, func(i, j int) bool {
			return frags[i].SeqID < frags[j].SeqID
		})

		// Join content
		var fullContent strings.Builder
		for i, f := range frags {
			if i > 0 {
				fullContent.WriteString("\n\n")
			}
			fullContent.WriteString(f.Content)
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      taskID,
			"status":  "completed",
			"content": fullContent.String(),
		})
	})

	// Submit Task Endpoint
	r.POST("/v1/secure/chat", func(c *gin.Context) {
		// 1. Parse Request
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
			return
		}

		var req OpenAIRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// Fallback for simple message format (if users send just a string or different format)
			// But for now, let's just log and return error
			fmt.Printf("JSON Unmarshal Error: %v. Body: %s\n", err, string(bodyBytes))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
			return
		}

		// 2. Decompose (Simple Splitter Logic)
		// We take the last user message and split it by newlines for this demo.
		// In a real app, this would be more sophisticated.
		lastMsg := ""
		if len(req.Messages) > 0 {
			lastMsg = req.Messages[len(req.Messages)-1].Content
		}

		// Split by newline, filtering empty lines
		lines := strings.Split(lastMsg, "\n")
		var fragments []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fragments = append(fragments, line)
			}
		}

		if len(fragments) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Empty content"})
			return
		}

		bigTaskID := fmt.Sprintf("task-%d", time.Now().UnixNano())
		ctx := context.Background()

		// Store metadata
		s.SetMeta(ctx, "meta:"+bigTaskID+":total", fmt.Sprintf("%d", len(fragments)), 24*time.Hour)

		// 3. Shuffle & Queue
		// We create fragment objects and push them to Redis
		// To demonstrate shuffle, we could randomize insertion order, 
		// but since the queue is FIFO and workers are concurrent, just pushing them is fine.
		// The "Shuffle" happens because multiple users are pushing at once.
		
		for i, content := range fragments {
			frag := Fragment{
				BigTaskID:  bigTaskID,
				SequenceID: i,
				Total:      len(fragments),
				Content:    content,
				Model:      req.Model,
			}
			
			fragJSON, _ := json.Marshal(frag)
			if err := s.PushQueue(ctx, "llm_fragment_queue", string(fragJSON)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue fragments"})
				return
			}
		}

		// Return the Task ID immediately (Async processing)
		c.JSON(http.StatusAccepted, gin.H{
			"id":      bigTaskID,
			"status":  "queued",
			"fragments": len(fragments),
			"poll_url": fmt.Sprintf("/v1/tasks/%s", bigTaskID),
		})
	})

	return r
}
