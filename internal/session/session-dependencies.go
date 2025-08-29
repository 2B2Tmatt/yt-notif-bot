package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

func randomB64(length int) string {
	byteSlice := make([]byte, length)
	_, err := rand.Read(byteSlice)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(byteSlice)
}

func EnsureSessionID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("sid")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	sid := randomB64(32)
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sid,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(10) * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return sid
}

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string ``
	Type             string `json:"token_type"`
	RefreshExpiresIn int    `json:"refresh_token_expires_in"`
}
