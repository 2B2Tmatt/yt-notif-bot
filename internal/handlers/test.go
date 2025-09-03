package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func Test(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	ctx := context.Background()
	for _, upload := range redisClient.ZRange(ctx, "uploads", 0, -1).Val() {
		log.Println("Upload: ", upload)
	}
	log.Println("Upload list complete")
}
