package jobs

// func ManageJobs(jobChan chan sessions.Token) {
// 	for {
// 		job := <-jobChan
// 		client := http.Client{Timeout: time.Second * 3}
// 		youtubeCall, _ := url.Parse("https://www.googleapis.com/youtube/v3/subscriptions")
// 		params := youtubeCall.Query()
// 		params.Add("part", "snippet")
// 		params.Add("part", "contentDetails")
// 		params.Add("mine", "true")
// 		youtubeCall.RawQuery = params.Encode()
// 		req, err := http.NewRequest(http.MethodGet, youtubeCall.String(), nil)
// 		if err != nil {
// 			log.Println("Error creating youtube request")
// 			return
// 		}

// 		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", job.AccessToken))
// 		resp, err := client.Do(req)
// 		if resp.StatusCode != http.StatusOK {
// 			log.Println("Error receiving response:", resp.StatusCode)
// 		}
// 		var channelIDs []string

// 	}
// }
