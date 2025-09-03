package yt

import (
	"context"
	"log"
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
