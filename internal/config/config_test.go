package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProfilesRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"work", "home", "alt"} {
		cfg.UpsertProfile(&Profile{Name: n, Email: n + "@example.com"})
	}
	if cfg.ActiveProfile != "work" {
		t.Errorf("active = %q, want the first profile added", cfg.ActiveProfile)
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.ProfileNames(), []string{"alt", "home", "work"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ProfileNames = %v, want %v", got, want)
	}
	if p, err := cfg.ResolveProfile(""); err != nil || p.Name != "work" || p.Email != "work@example.com" {
		t.Errorf("ResolveProfile = %+v, %v", p, err)
	}
	if _, err := cfg.ResolveProfile("nope"); err == nil {
		t.Error("ResolveProfile accepted an unknown profile")
	}
}

func TestRemoveProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"work", "zeta", "beta"} {
		cfg.UpsertProfile(&Profile{Name: n})
	}
	tok, err := TokenPath("work")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tok, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := cfg.RemoveProfile("work"); err != nil {
		t.Fatal(err)
	}
	// Used to pick a random remaining profile via map iteration.
	if cfg.ActiveProfile != "beta" {
		t.Errorf("active after removal = %q, want %q", cfg.ActiveProfile, "beta")
	}
	if _, err := os.Stat(tok); !os.IsNotExist(err) {
		t.Errorf("token %s not removed (stat err: %v)", filepath.Base(tok), err)
	}
	if err := cfg.RemoveProfile("work"); err == nil {
		t.Error("removing a missing profile succeeded")
	}
}
