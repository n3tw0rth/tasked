package auth

import (
	"errors"
	"fmt"
	"os"
)

const (
	envClientID     = "TASKED_CLIENT_ID"
	envClientSecret = "TASKED_CLIENT_SECRET"
)

var errNoClient = errors.New("this build has no built-in OAuth client — set TASKED_CLIENT_ID and " +
	"TASKED_CLIENT_SECRET to your own Desktop app client (see the README)")

// Client is an OAuth client ID and secret, and where they were loaded from.
type Client struct {
	ID     string
	Secret string
	Source string // "environment" or "built-in"
}

// LoadClient returns the OAuth client to sign in with: TASKED_CLIENT_ID and
// TASKED_CLIENT_SECRET if both are set, else the built-in client.
func LoadClient() (Client, error) {
	id, secret := os.Getenv(envClientID), os.Getenv(envClientSecret)
	switch {
	case id != "" && secret != "":
		return Client{ID: id, Secret: secret, Source: "environment"}, nil
	case id != "" || secret != "":
		return Client{}, fmt.Errorf("set both %s and %s, or neither", envClientID, envClientSecret)
	}
	if builtinClient.ID == "" || builtinClient.Secret == "" {
		return Client{}, errNoClient
	}
	return builtinClient, nil
}
