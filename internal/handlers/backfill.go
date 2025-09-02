package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	yt "yt-notif/internal/youtube"

	"github.com/redis/go-redis/v9"
)

func Backfill(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	params := url.Values{}
	point := "https://www.googleapis.com/youtube/v3/subscriptions"
	params.Add("part", "snippet")
	params.Add("mine", "true")
	params.Add("maxResults", "50")
	client := http.Client{Timeout: time.Second * 2}
	encoded := params.Encode()
	fullURL := fmt.Sprintf("%s?%s", point, encoded)
	ctx := context.Background()
	token := redisClient.Get(ctx, "access_token").Val()
	yt.StoreAllChannelIDS(ctx, client, token, fullURL, redisClient)
	yt.LoadPlaylistsIntoMemory(ctx, client, token, redisClient)
	yt.LoadUploadsIntoMemory(ctx, client, token, redisClient)
	log.Println("Backfill Complete")
}
