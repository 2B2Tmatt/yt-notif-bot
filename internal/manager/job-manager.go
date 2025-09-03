package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

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

type UploadJob struct {
	Ctx         context.Context
	Client      *http.Client
	Token       string
	RedisClient *redis.Client
	Cutoff      time.Time
	Playlist    string
	Point       string
}

func UploadWorker(id int, uploadJob <-chan UploadJob, wg *sync.WaitGroup) error {
	log.Println("Worker:", id, "reporting")
	defer wg.Done()
	for {
		Job, open := <-uploadJob
		if !open {
			log.Println("Worker", id, "returning")
			return nil
		}
		params := url.Values{}
		params.Add("part", "contentDetails")
		params.Add("maxResults", "50")
		params.Add("playlistId", Job.Playlist)
		encoded := params.Encode()
		fullURL := fmt.Sprintf("%s?%s", Job.Point, encoded)
		req, err := http.NewRequest(http.MethodGet, fullURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Job.Token))
		req.Header.Set("Accept", "application/json")
		resp, err := Job.Client.Do(req)
		if err != nil {
			return err
		}
		log.Println("Worker:", id, "status:", resp.StatusCode)
		var uploads Uploads
		json.NewDecoder(resp.Body).Decode(&uploads)
		for _, upload := range uploads.UploadItems {
			if upload.UploadDetails.VideoPublishedAt.After(Job.Cutoff) {
				Job.RedisClient.ZAdd(Job.Ctx, "uploads", redis.Z{Member: upload.UploadDetails.VideoID,
					Score: float64(upload.UploadDetails.VideoPublishedAt.Add(time.Hour).Unix())})
				log.Println("Upload Added:", upload.UploadDetails.VideoID)
			} else {
				break
			}
		}
		log.Println("Worker", id, "done")
	}
}
