package yt

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func Sweeper(ctx context.Context, redisClient *redis.Client) {
	log.Println("Sweeper Online")
	ticker := time.NewTicker(time.Duration(15) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Tick:", time.Now())
		endcond := time.Now().Unix()
		redisClient.ZRemRangeByScore(ctx, "uploads", "-inf", strconv.Itoa(int(endcond)))
	}
}

func Poller(ctx context.Context, redisClient *redis.Client, client http.Client, tokenChan chan string) {
	token := <-tokenChan
	log.Println("Poller Online")
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Poll:", time.Now())
		err := LoadUploadsIntoMemory(ctx, client, token, redisClient)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		log.Println("Polling complete")
	}
}
