package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"yt-notif/internal/handlers"
	sessions "yt-notif/internal/session"
	yt "yt-notif/internal/youtube"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

func main() {

	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env")
	}
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	user := os.Getenv("REDIS_USER")
	pass := os.Getenv("REDIS_PASSWORD")
	botToken := os.Getenv("BOT_TOKEN")
	addr := net.JoinHostPort(host, port)

	redisClient := redis.NewClient(&redis.Options{
		Addr:       addr,
		Username:   user,
		Password:   pass,
		DB:         0,
		MaxRetries: 2,
	})
	redisClient.Del(ctx, "channels")
	redisClient.Del(ctx, "playlists")
	redisClient.Del(ctx, "uploads")
	jobChannel := make(chan sessions.Token)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connect failed: %v", err)
	}
	fmt.Println("Connected to Redis")
	tokenChan := make(chan string)
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { handlers.Auth(w, r, redisClient) })
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) { handlers.Callback(w, r, redisClient, jobChannel) })
	http.HandleFunc("/backfill", func(w http.ResponseWriter, r *http.Request) { handlers.Backfill(w, r, redisClient, tokenChan) })
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) { handlers.Test(w, r, redisClient) })
	client := http.Client{Timeout: time.Second * 2}
	go yt.Sweeper(ctx, redisClient)
	go yt.Poller(ctx, redisClient, client, tokenChan)
	sess, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		if m.Content == "GET" {
			log.Println("Bot Working")
			for _, upload := range redisClient.ZRange(ctx, "uploads", 0, -1).Val() {
				_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("https://www.youtube.com/watch?v=%s", upload))
				if err != nil {
					log.Println("Error:", err)
				}
			}
		}
	})
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()
	log.Println("Bot online")
	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Println(err)
	}
	redisClient.Del(ctx, "channels")
	redisClient.Del(ctx, "playlists")
	redisClient.Del(ctx, "uploads")
}
