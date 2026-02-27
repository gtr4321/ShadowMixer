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
	"shadowmixer/store"
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

	// 3. Init Store (Try Redis first, fallback to Memory)
	var s store.Store

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Global.Redis.Addr,
		Password: config.Global.Redis.Password,
		DB:       config.Global.Redis.DB,
	})

	// Check Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		fmt.Printf("[Main] Redis connection failed (%v). Falling back to In-Memory Store.\n", err)
		fmt.Println("[Main] WARNING: In-Memory Store is for demo/testing only. Data will be lost on restart.")
		s = store.NewMemoryStore()
	} else {
		fmt.Println("[Main] Redis connected successfully.")
		s = store.NewRedisStore(rdb)
	}
	defer s.Close()

	// 4. Start Worker (Background Process)
	// In a real deployment, this might run in a separate container/process.
	// For this MVP, we run it as a goroutine.
	wk := worker.New(s, config.Global)
	go wk.Start()

	// 5. Start HTTP Router
	r := router.SetupRouter(s, config.Global)
	
	port := config.Global.Server.Port
	if port == "" {
		port = ":8080"
	}
	fmt.Printf("[Main] ShadowMixer Gateway starting on %s...\n", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
