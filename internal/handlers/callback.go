package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	sessions "yt-notif/internal/session"

	"github.com/redis/go-redis/v9"
)

func Callback(w http.ResponseWriter, r *http.Request, redisClient *redis.Client, jobChan chan<- sessions.Token) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		log.Println("No cookie in request", err)
		return
	}
	sid := cookie.Value
	ctx := context.Background()
	if storedSid := redisClient.Get(ctx, "sid"); storedSid.Val() != sid {
		log.Println("Incorrect session id")
		http.Error(w, "Incorrect session id", http.StatusBadRequest)
		return
	}
	if authTTL := redisClient.TTL(ctx, "state"); authTTL.Val() <= 0 {
		log.Println("The TTL is ", authTTL)
		log.Println("Auth session expired")
		http.Error(w, "Auth sesion expired", http.StatusGatewayTimeout)
		return
	}
	paramErr := r.URL.Query().Get("error")
	if paramErr == "access_denied" {
		log.Println("Access denied")
		http.Error(w, "access denied", http.StatusUnauthorized)
		return
	}
	state := r.URL.Query().Get("state")
	if storedState := redisClient.Get(ctx, "state"); storedState.Val() != state {
		log.Println("Incorrect State, Stored:", storedState, "Sent:", state)
		http.Error(w, "incorrect state", http.StatusUnauthorized)
		return
	}
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	var secrets clientInfo
	json.Unmarshal(b, &secrets)
	code := r.URL.Query().Get("code")
	client := &http.Client{Timeout: 10 * time.Second}
	bodyData := url.Values{
		"grant_type":    []string{"authorization_code"},
		"code":          []string{code},
		"redirect_uri":  []string{"http://127.0.0.1:8888/callback"},
		"client_id":     []string{secrets.WebLayer.ClientID},
		"client_secret": []string{secrets.WebLayer.ClientSecret},
	}
	encodedBody := bodyData.Encode()
	reader := strings.NewReader(encodedBody)
	req, err := http.NewRequest(http.MethodPost, "https://oauth2.googleapis.com/token", reader)
	if err != nil {
		log.Println("Error creating token request", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error executing request")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Error getting response token")
		return
	}
	var accessToken sessions.Token
	err = json.NewDecoder(resp.Body).Decode(&accessToken)
	if err != nil {
		log.Println("Error decoding response token")
		return
	}
	redisClient.Set(ctx, "access_token", accessToken.AccessToken, time.Duration(accessToken.ExpiresIn)*time.Second)
	if accessToken.RefreshExpiresIn < 0 {
		redisClient.Set(ctx, "refresh_token", accessToken.RefreshToken, time.Duration(accessToken.RefreshExpiresIn)*time.Second)
	} else {
		redisClient.Set(ctx, "refresh_token", accessToken.RefreshToken, 0)
	}
	jobChan <- accessToken
	log.Println("access_token:", accessToken.AccessToken)
	fmt.Fprintln(w, "Success!")
}

type clientInfo struct {
	WebLayer ClientInfo `json:"web"`
}

type ClientInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
