package main

import (
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"net/http"
	"os"
)

func main() {
	// Load settings:
	// ---
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	addr := ":" + port

	baseURL := os.Getenv("BASE_URL")
	if len(baseURL) == 0 {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	redisURL := os.Getenv("REDIS_URL")
	if len(redisURL) == 0 {
		redisURL = "redis://:@localhost:6379/1"
	}

	// Bootstrap:
	// ---
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	redisCache := cache.New(&cache.Options{
		Redis: redisClient,
	})
	server := &Server{
		BaseURL:    baseURL,
		RedisCache: redisCache,
	}

	// Start web server:
	// ---
	fmt.Printf("Starting web server, listening on %s\n", addr)
	err = http.ListenAndServe(addr, server)
	if err != nil {
		panic(err)
	}
}
