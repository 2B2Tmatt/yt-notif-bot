package jobs

import (
	"log"
	"net/http"
	"net/url"
	"time"
	sessions "yt-notif/internal/session"
)

func ManageJobs(jobChan chan sessions.Token) {
	for {
		job := <-jobChan
		client := http.Client{Timeout: time.Second * 3}
		youtubeCall, _ := url.Parse("https://www.googleapis.com/youtube/v3/channels")
		params := youtubeCall.Query()
		params.Add("part", "id")
		params.Add("mine", "true")
		youtubeCall.RawQuery = params.Encode()
		req, err := http.NewRequest(http.MethodGet, youtubeCall.String(), nil)
		if err != nil {
			log.Println("Error creating youtube request")
		}
	}
}
