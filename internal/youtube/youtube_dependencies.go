package yt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type ChannelList struct {
	NextPageToken string `json:"nextPageToken"`
	Items         []Item `json:"items"`
}

type Item struct {
	Snip Snippet `json:"snippet"`
}

type Snippet struct {
	ResourceID Resource `json:"resourceId"`
}

type Resource struct {
	ChannelID string `json:"channelId"`
}

func StoreAllChannelIDS(ctx context.Context, client http.Client, token string, URL string, redisClient *redis.Client) error {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	var channelList ChannelList
	json.NewDecoder(resp.Body).Decode(&channelList)
	for _, Item := range channelList.Items {
		intcmd := redisClient.SAdd(ctx, "channels", Item.Snip.ResourceID.ChannelID)
		if intcmd.Val() == int64(0) {
			log.Println("Error on SAdd channel:", Item.Snip.ResourceID.ChannelID)
		}
	}
	redisClient.Expire(ctx, "channels", 300*time.Second)
	return nil
}

func LoadUploadsIntoMemory(ctx context.Context, client http.Client, token string, URL string, redisClient *redis.Client) error {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")

	return nil
}
