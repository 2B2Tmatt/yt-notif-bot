package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"yt-notif/internal/handlers"
	sessions "yt-notif/internal/session"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env")
	}
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	user := os.Getenv("REDIS_USER")
	pass := os.Getenv("REDIS_PASSWORD")
	addr := net.JoinHostPort(host, port)

	redisClient := redis.NewClient(&redis.Options{
		Addr:       addr,
		Username:   user,
		Password:   pass,
		DB:         0,
		MaxRetries: 2,
	})
	jobChannel := make(chan sessions.Token)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connect failed: %v", err)
	}
	fmt.Println("Connected to Redis")
	// go jobs.ManageJobs(jobChannel)
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { handlers.Auth(w, r, redisClient) })
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) { handlers.Callback(w, r, redisClient, jobChannel) })
	http.HandleFunc("/backfill", func(w http.ResponseWriter, r *http.Request) { handlers.Backfill(w, r, redisClient) })
	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Println(err)
	}
	defer redisClient.Del(ctx, "channels")
}
