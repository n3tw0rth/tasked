package auth

import "os"

// Set at build time with:
//   -ldflags "-X github.com/n3tw0rth/tasked/internal/auth.clientID=... -X github.com/n3tw0rth/tasked/internal/auth.clientSecret=..."
var (
	clientID     = ""
	clientSecret = ""
)

func ClientID() string {
	if v := os.Getenv("TASKED_CLIENT_ID"); v != "" {
		return v
	}
	return clientID
}

func ClientSecret() string {
	if v := os.Getenv("TASKED_CLIENT_SECRET"); v != "" {
		return v
	}
	return clientSecret
}
