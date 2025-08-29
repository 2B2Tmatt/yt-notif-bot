package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
	"yt-notif/internal/auth"
	sessions "yt-notif/internal/session"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

func Auth(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	state, err := auth.GenerateRandomString(32)
	if err != nil {
		log.Fatal("Unable to generate state", err)
	}
	sid := sessions.EnsureSessionID(w, r)
	ctx := context.Background()
	redisClient.Set(ctx, "state", state, time.Duration(10)*time.Minute)
	redisClient.Set(ctx, "sid", sid, time.Duration(10)*time.Minute)
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
