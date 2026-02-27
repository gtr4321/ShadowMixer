package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"shadowmixer/config"
	"shadowmixer/router"
	"shadowmixer/worker"
)

func main() {
	// 1. Seed Random for Jitter
	rand.Seed(time.Now().UnixNano())

	// 2. Load Config
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("[Main] Config loaded successfully.")

	// 3. Init Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Global.Redis.Addr,
		Password: config.Global.Redis.Password,
		DB:       config.Global.Redis.DB,
	})
	// Simple ping check
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("[Main] Redis connected.")

	// 4. Start Worker (Background Process)
	// In a real deployment, this might run in a separate container/process.
	// For this MVP, we run it as a goroutine.
	wk := worker.New(rdb, config.Global)
	go wk.Start()

	// 5. Start HTTP Router
	r := router.SetupRouter(rdb, config.Global)
	
	port := config.Global.Server.Port
	if port == "" {
		port = ":8080"
	}
	fmt.Printf("[Main] ShadowMixer Gateway starting on %s...\n", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
