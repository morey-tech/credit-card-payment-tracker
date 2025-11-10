package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	DiscordWebhookURL string `yaml:"discord_webhook_url"`
}

// LoadConfig loads configuration from a YAML file
// If the file doesn't exist, returns default configuration
// If CONFIG_PATH environment variable is set, uses that path
// Otherwise defaults to ./config.yaml
func LoadConfig(path string) (*Config, error) {
	// Use environment variable if set
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		path = envPath
	}

	// If no path specified, use default
	if path == "" {
		path = "./config.yaml"
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			return &Config{
				DiscordWebhookURL: "",
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(path string, cfg *Config) error {
	// Use environment variable if set
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		path = envPath
	}

	// If no path specified, use default
	if path == "" {
		path = "./config.yaml"
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Discord webhook URL validation
	if c.DiscordWebhookURL != "" {
		if !strings.HasPrefix(c.DiscordWebhookURL, "https://discord.com/api/webhooks/") &&
			!strings.HasPrefix(c.DiscordWebhookURL, "https://discordapp.com/api/webhooks/") {
			return fmt.Errorf("discord webhook URL must start with https://discord.com/api/webhooks/ or https://discordapp.com/api/webhooks/")
		}
	}

	return nil
}
