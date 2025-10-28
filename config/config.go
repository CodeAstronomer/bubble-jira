package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
		APIToken: "",
		JQL:      "assignee = currentuser() ORDER BY updated DESC",
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

// ---------- Config Editor helpers ----------

type ConfigField struct {
	Key   string
	Value string
}

func (cf ConfigField) Title() string       { return cf.Key }
func (cf ConfigField) Description() string { return cf.Value }
func (cf ConfigField) FilterValue() string { return cf.Key }

func (c *Config) ToFields() []ConfigField {
	return []ConfigField{
		{"Base URL", c.BaseURL},
		{"Email", c.Email},
		{"API Token", c.APIToken},
		{"JQL Query", c.JQL},
	}
}

func (c *Config) UpdateFromFields(fields []ConfigField) error {
	for _, field := range fields {
		switch strings.ToLower(field.Key) {
		case "base url":
			c.BaseURL = field.Value
		case "email":
			c.Email = field.Value
		case "api token":
			c.APIToken = field.Value
		case "jql query":
			c.JQL = field.Value
		}
	}
	// Persist immediately
	return Save(c)
}
