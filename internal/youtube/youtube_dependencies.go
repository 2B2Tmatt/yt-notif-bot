package yt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

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

type Uploads struct {
	UploadItems []UploadItem `json:"items"`
}

type UploadItem struct {
	UploadDetails UploadDetail `json:"contentDetails"`
}

type UploadDetail struct {
	VideoID          string    `json:"videoId"`
	VideoPublishedAt time.Time `json:"videoPublishedAt"`
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
	for _, playlist := range PlaylistSlice {
		params := url.Values{}
		params.Add("part", "contentDetails")
		params.Add("maxResults", "50")
		params.Add("playlistId", playlist)
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
		var uploads Uploads
		json.NewDecoder(resp.Body).Decode(&uploads)
		for _, upload := range uploads.UploadItems {
			if upload.UploadDetails.VideoPublishedAt.After(cutoff) {
				redisClient.ZAdd(ctx, "uploads", redis.Z{Member: upload.UploadDetails.VideoID, Score: float64(upload.UploadDetails.VideoPublishedAt.Unix())})
				redisClient.Set(ctx, upload.UploadDetails.VideoID, true, upload.UploadDetails.VideoPublishedAt.Sub(cutoff)*time.Second)
				log.Println("Upload Added:", upload.UploadDetails.VideoID)
			} else {
				break
			}
		}
	}
	log.Println("All uploads backfilled")
	return nil
}
