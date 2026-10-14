package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func callback(t *testing.T, h http.Handler, query string) int {
	t.Helper()
	done := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/callback?"+query, nil))
		done <- rec.Code
	}()
	select {
	case code := <-done:
		return code
	case <-time.After(2 * time.Second):
		t.Fatalf("callback %q blocked", query)
		return 0
	}
}

func TestCallbackHandler(t *testing.T) {
	resCh := make(chan callbackResult, 1)
	h := callbackHandler("s3cret", resCh)

	// A request with the wrong state is rejected and doesn't end the login.
	if code := callback(t, h, "state=wrong&code=evil"); code != http.StatusBadRequest {
		t.Errorf("bad state: status %d, want 400", code)
	}
	select {
	case r := <-resCh:
		t.Fatalf("bad-state request was delivered: %+v", r)
	default:
	}

	if code := callback(t, h, "state=s3cret&code=abc"); code != http.StatusOK {
		t.Errorf("good callback: status %d, want 200", code)
	}
	// Regression: a second hit (reload, prefetch) used to block forever on
	// the channel, which in turn hung server shutdown and `tasked login`.
	callback(t, h, "state=s3cret&code=again")

	r := <-resCh
	if r.err != nil || r.code != "abc" {
		t.Errorf("delivered %+v, want code abc", r)
	}
	select {
	case r := <-resCh:
		t.Errorf("second callback was delivered: %+v", r)
	default:
	}
}

func TestCallbackHandlerError(t *testing.T) {
	resCh := make(chan callbackResult, 1)
	h := callbackHandler("s", resCh)
	if code := callback(t, h, "state=s&error=access_denied"); code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", code)
	}
	if r := <-resCh; r.err == nil || !strings.Contains(r.err.Error(), "access_denied") {
		t.Errorf("got %+v, want access_denied error", r)
	}
}

type fakeSource struct{ err error }

func (f fakeSource) Token() (*oauth2.Token, error) { return nil, f.err }

func TestPersistingSourceReauth(t *testing.T) {
	expired := &oauth2.RetrieveError{ErrorCode: "invalid_grant", ErrorDescription: "Token has been expired or revoked."}
	for _, err := range []error{expired, fmt.Errorf("refresh: %w", expired)} {
		p := &persistingSource{profile: "work", src: fakeSource{err: err}}
		_, got := p.Token()
		var re *ReauthError
		if !errors.As(got, &re) {
			t.Fatalf("Token() error = %v, want *ReauthError", got)
		}
		if msg := re.Error(); !strings.Contains(msg, "tasked login --profile work") || !strings.Contains(msg, "expired or revoked") {
			t.Errorf("unhelpful message: %s", msg)
		}
	}

	// A token issued to a different OAuth client (e.g. after switching
	// between TASKED_CLIENT_ID and the built-in client) can't be refreshed.
	mismatch := &oauth2.RetrieveError{ErrorCode: "unauthorized_client", ErrorDescription: "Unauthorized"}
	p := &persistingSource{profile: "work", src: fakeSource{err: mismatch}}
	_, got := p.Token()
	var re *ReauthError
	if !errors.As(got, &re) || !strings.Contains(re.Error(), "different OAuth client") {
		t.Errorf("client mismatch: Token() error = %v, want a ReauthError about the client", got)
	}

	other := errors.New("network down")
	p = &persistingSource{profile: "work", src: fakeSource{err: other}}
	if _, got := p.Token(); got != other {
		t.Errorf("non-auth error = %v, want it passed through", got)
	}
}
