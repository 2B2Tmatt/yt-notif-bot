package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os/exec"
	"runtime"
)

func GenerateRandomString(length int) (string, error) {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	randString := base64.RawURLEncoding.EncodeToString(randBytes)
	return randString, nil
}

func OpenBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}
