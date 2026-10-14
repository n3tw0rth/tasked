package auth

import (
	"errors"
	"strings"
	"testing"
)

// setClients sets the env override and the built-in client for one test.
func setClients(t *testing.T, envID, envSecret string, builtin Client) {
	t.Helper()
	t.Setenv(envClientID, envID)
	t.Setenv(envClientSecret, envSecret)
	saved := builtinClient
	builtinClient = builtin
	t.Cleanup(func() { builtinClient = saved })
}

var testBuiltin = Client{ID: "builtin-id", Secret: "builtin-secret", Source: "built-in"}

func TestLoadClientBuiltin(t *testing.T) {
	setClients(t, "", "", testBuiltin)
	if c, err := LoadClient(); err != nil || c != testBuiltin {
		t.Errorf("LoadClient = %+v, %v; want the built-in client", c, err)
	}
}

func TestLoadClientEnvOverride(t *testing.T) {
	setClients(t, "env-id", "env-secret", testBuiltin)
	c, err := LoadClient()
	if err != nil || c.ID != "env-id" || c.Secret != "env-secret" || c.Source != "environment" {
		t.Errorf("LoadClient = %+v, %v; want the environment client", c, err)
	}

	t.Setenv(envClientSecret, "")
	if _, err := LoadClient(); err == nil || !strings.Contains(err.Error(), "set both") {
		t.Errorf("half-set env: error = %v, want 'set both'", err)
	}
}

func TestLoadClientNoneConfigured(t *testing.T) {
	setClients(t, "", "", Client{Source: "built-in"})
	if _, err := LoadClient(); !errors.Is(err, errNoClient) {
		t.Errorf("LoadClient() error = %v, want errNoClient", err)
	}
}
