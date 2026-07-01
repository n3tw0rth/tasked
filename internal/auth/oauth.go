package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	tasksapi "google.golang.org/api/tasks/v1"

	"github.com/n3tw0rth/tasked/internal/config"
)

const userinfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

var Scopes = []string{
	tasksapi.TasksScope,
	"https://www.googleapis.com/auth/userinfo.email",
}

func oauthConfig(redirectURL string) (*oauth2.Config, error) {
	id, secret := ClientID(), ClientSecret()
	if id == "" || secret == "" {
		return nil, errors.New("OAuth client credentials are not configured (build with -ldflags or set TASKED_CLIENT_ID/TASKED_CLIENT_SECRET)")
	}
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     google.Endpoint,
		Scopes:       Scopes,
		RedirectURL:  redirectURL,
	}, nil
}

// Login runs the loopback OAuth flow and saves the resulting token under
// the given profile name. It returns the user's email address.
func Login(ctx context.Context, profileName string) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("bind loopback: %w", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	redirect := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	cfg, err := oauthConfig(redirect)
	if err != nil {
		return "", err
	}

	stateBytes := make([]byte, 24)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", err
	}
	state := hex.EncodeToString(stateBytes)

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	type result struct {
		code string
		err  error
	}
	resCh := make(chan result, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("state"); got != state {
			http.Error(w, "bad state", http.StatusBadRequest)
			resCh <- result{err: errors.New("state mismatch")}
			return
		}
		if e := q.Get("error"); e != "" {
			http.Error(w, e, http.StatusBadRequest)
			resCh <- result{err: fmt.Errorf("oauth error: %s", e)}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			resCh <- result{err: errors.New("missing code")}
			return
		}
		fmt.Fprintln(w, "<!doctype html><html><body><h2>tasked: signed in.</h2><p>You can close this tab.</p></body></html>")
		resCh <- result{code: code}
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Shutdown(context.Background())

	fmt.Println("Opening browser for Google sign-in...")
	fmt.Println("(listening on", redirect+")")
	if err := openBrowser(authURL); err != nil {
		fmt.Println("Could not open browser automatically. Visit:")
		fmt.Println("  ", authURL)
	}

	var code string
	select {
	case r := <-resCh:
		if r.err != nil {
			return "", r.err
		}
		code = r.code
	case <-time.After(5 * time.Minute):
		return "", errors.New("timed out waiting for browser callback")
	case <-ctx.Done():
		return "", ctx.Err()
	}

	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("token exchange: %w", err)
	}
	if err := saveToken(profileName, tok); err != nil {
		return "", err
	}

	email, err := fetchEmail(ctx, cfg, tok)
	if err != nil {
		return "", fmt.Errorf("fetch userinfo: %w", err)
	}
	return email, nil
}

// TokenSource returns an oauth2.TokenSource for the given profile that persists
// refreshed tokens back to disk automatically.
func TokenSource(ctx context.Context, profileName string) (oauth2.TokenSource, error) {
	tok, err := loadToken(profileName)
	if err != nil {
		return nil, err
	}
	cfg, err := oauthConfig("")
	if err != nil {
		return nil, err
	}
	src := cfg.TokenSource(ctx, tok)
	return &persistingSource{profile: profileName, src: src, last: tok}, nil
}

type persistingSource struct {
	profile string
	src     oauth2.TokenSource
	mu      sync.Mutex
	last    *oauth2.Token
}

func (p *persistingSource) Token() (*oauth2.Token, error) {
	tok, err := p.src.Token()
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	changed := p.last == nil || tok.AccessToken != p.last.AccessToken || tok.RefreshToken != p.last.RefreshToken
	p.last = tok
	p.mu.Unlock()
	if changed {
		_ = saveToken(p.profile, tok)
	}
	return tok, nil
}

func saveToken(profile string, tok *oauth2.Token) error {
	p, err := config.TokenPath(profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func loadToken(profile string) (*oauth2.Token, error) {
	p, err := config.TokenPath(profile)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no saved token for profile %q — run `tasked login`", profile)
		}
		return nil, err
	}
	tok := &oauth2.Token{}
	if err := json.Unmarshal(data, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func Logout(profile string) error {
	p, err := config.TokenPath(profile)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func fetchEmail(ctx context.Context, cfg *oauth2.Config, tok *oauth2.Token) (string, error) {
	client := cfg.Client(ctx, tok)
	resp, err := client.Get(userinfoURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("userinfo status %d", resp.StatusCode)
	}
	var info struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.Email == "" {
		return "", errors.New("empty email in userinfo response")
	}
	return info.Email, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
