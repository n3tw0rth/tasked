package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name            string `yaml:"-"`
	Email           string `yaml:"email"`
	ActiveListID    string `yaml:"active_list_id,omitempty"`
	ActiveListTitle string `yaml:"active_list_title,omitempty"`
}

type Config struct {
	ActiveProfile string              `yaml:"active_profile"`
	Profiles      map[string]*Profile `yaml:"profiles"`

	path string `yaml:"-"`
}

func Dir() (string, error) {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "tasked"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tasked"), nil
}

func TokenPath(profile string) (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "tokens", profile+".json"), nil
}

func configPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(filepath.Dir(p), "tokens"), 0o700); err != nil {
		return nil, err
	}
	cfg := &Config{path: p, Profiles: map[string]*Profile{}}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	for name, prof := range cfg.Profiles {
		prof.Name = name
	}
	cfg.path = p
	return cfg, nil
}

func (c *Config) Save() error {
	if c.path == "" {
		p, err := configPath()
		if err != nil {
			return err
		}
		c.path = p
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

func (c *Config) ResolveProfile(flag string) (*Profile, error) {
	name := flag
	if name == "" {
		name = c.ActiveProfile
	}
	if name == "" {
		return nil, errors.New("no active profile — run `tasked login` first")
	}
	p, ok := c.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

func (c *Config) UpsertProfile(p *Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	c.Profiles[p.Name] = p
	if c.ActiveProfile == "" {
		c.ActiveProfile = p.Name
	}
}

func (c *Config) RemoveProfile(name string) error {
	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(c.Profiles, name)
	if c.ActiveProfile == name {
		c.ActiveProfile = ""
		for n := range c.Profiles {
			c.ActiveProfile = n
			break
		}
	}
	tp, err := TokenPath(name)
	if err == nil {
		_ = os.Remove(tp)
	}
	return nil
}
