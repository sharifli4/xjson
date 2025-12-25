package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	BearerToken  string `yaml:"bearer_token"`
	RedirectURL  string `yaml:"redirect_url"`
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	// Look for config in current directory first
	if _, err := os.Stat("xjson.yaml"); err == nil {
		return "xjson.yaml"
	}
	// Fall back to executable directory
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		return filepath.Join(dir, "xjson.yaml")
	}
	return "xjson.yaml"
}

// Load reads config from the default path
func Load() (*Config, error) {
	return LoadFromPath(DefaultConfigPath())
}

// LoadFromPath reads config from a specific path
func LoadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s - run 'xjson init' to create one", path)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.RedirectURL == "" {
		cfg.RedirectURL = "http://localhost:8080/callback"
	}

	return &cfg, nil
}

// Save writes config to the default path
func Save(cfg *Config) error {
	return SaveToPath(cfg, DefaultConfigPath())
}

// SaveToPath writes config to a specific path
func SaveToPath(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// CreateDefault creates a default config file
func CreateDefault() error {
	cfg := &Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		BearerToken:  "YOUR_BEARER_TOKEN",
		RedirectURL:  "http://localhost:8080/callback",
	}
	return Save(cfg)
}
