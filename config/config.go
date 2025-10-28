package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds Jira connection settings.
type Config struct {
	BaseURL  string `json:"base_url"`
	Email    string `json:"email"`
	APIToken string `json:"api_token"`
	JQL      string `json:"jql"`
}

// DefaultConfig returns a sensible default.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:  "https://your-domain.atlassian.net",
		Email:    "you@example.com",
		APIToken: "your-atlassian-api-token",
		JQL:      "assignee = currentuser() AND (status != Done AND status != Erledigt) ORDER BY updated DESC",
	}
}

// ConfigFile returns the path to config.json
func ConfigFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(home, ".config", "bubble-jira", "config.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return "", err
	}
	return p, nil
}

// Save writes the config to JSON
func Save(cfg *Config) error {
	p, err := ConfigFile()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ") // pretty print
	return encoder.Encode(cfg)
}

// Load reads the config.json or creates default if missing
func Load() (*Config, error) {
	p, err := ConfigFile()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsValid checks if the config has all required fields properly filled
func (c *Config) IsValid() bool {
	if c.BaseURL == "" || c.Email == "" || c.APIToken == "" || c.JQL == "" {
		return false
	}

	if c.BaseURL == "https://your-domain.atlassian.net" && c.Email == "you@example.com" && c.APIToken == "your-atlassian-api-token" {
		return false
	}

	return true
}