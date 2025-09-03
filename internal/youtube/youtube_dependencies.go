package yt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
	jobs "yt-notif/internal/manager"

	"github.com/redis/go-redis/v9"
)

type PlaylistList struct {
	Playlistitems []PlaylistItem `json:"items"`
}

type PlaylistItem struct {
	ContentDetails ContentDetail `json:"contentDetails"`
}

type ContentDetail struct {
	RelatedPlaylists RelatedPlaylist `json:"relatedPlaylists"`
}

type RelatedPlaylist struct {
	Uploads string `json:"uploads"`
}

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
	defer resp.Body.Close()
	var channelList ChannelList
	json.NewDecoder(resp.Body).Decode(&channelList)
	for _, Item := range channelList.Items {
		intcmd := redisClient.SAdd(ctx, "channels", Item.Snip.ResourceID.ChannelID)
		log.Println("Added channel: ", Item.Snip.ResourceID.ChannelID)
		if intcmd.Val() == int64(0) {
			log.Println("Error on SAdd channel:", Item.Snip.ResourceID.ChannelID)
		}
	}
	redisClient.Expire(ctx, "channels", 30*time.Second)
	log.Println("Channel store completed at", time.Now())
	return nil
}

func LoadPlaylistsIntoMemory(ctx context.Context, client http.Client, token string, redisClient *redis.Client) error {
	params := url.Values{}
	point := "https://www.googleapis.com/youtube/v3/channels"
	params.Add("part", "contentDetails")
	channelsSlice := redisClient.SMembers(ctx, "channels")
	for _, channel := range channelsSlice.Val() {
		params.Add("id", channel)
	}
	encoded := params.Encode()
	fullURL := fmt.Sprintf("%s?%s", point, encoded)
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var Playlists PlaylistList
	json.NewDecoder(resp.Body).Decode(&Playlists)
	for _, Playlist := range Playlists.Playlistitems {
		intcmd := redisClient.SAdd(ctx, "playlists", Playlist.ContentDetails.RelatedPlaylists.Uploads)
		log.Println("Playlist added: ", Playlist.ContentDetails.RelatedPlaylists.Uploads)
		if intcmd.Val() == int64(0) {
			log.Println("Error on SAdd playlist", Playlist.ContentDetails.RelatedPlaylists.Uploads)
		}
	}
	redisClient.Expire(ctx, "playlists", time.Second*100)
	log.Println("Playlist store completed at", time.Now())
	return nil
}

func LoadUploadsIntoMemory(ctx context.Context, client http.Client, token string, redisClient *redis.Client) error {
	cutoff := time.Now().Add(-time.Hour)
	point := "https://www.googleapis.com/youtube/v3/playlistItems"
	PlaylistSlice := redisClient.SMembers(ctx, "playlists").Val()
	var wg sync.WaitGroup
	UploadJobChan := make(chan jobs.UploadJob, 8)
	numWorkers := 8
	wg.Add(numWorkers)
	for i := range numWorkers {
		go jobs.UploadWorker(i, UploadJobChan, &wg)
	}
	for _, playlist := range PlaylistSlice {
		uploadJob := jobs.UploadJob{Ctx: ctx, Client: &client, Token: token, RedisClient: redisClient, Cutoff: cutoff, Playlist: playlist, Point: point}
		UploadJobChan <- uploadJob
	}
	close(UploadJobChan)
	wg.Wait()
	log.Println("All uploads backfilled")
	return nil
}
