package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config
	testConfig := `discord_webhook_url: "https://discord.com/api/webhooks/123456/abcdef"`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	expected := "https://discord.com/api/webhooks/123456/abcdef"
	if cfg.DiscordWebhookURL != expected {
		t.Errorf("Expected webhook URL %q, got %q", expected, cfg.DiscordWebhookURL)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Load config (should return default)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed on missing file: %v", err)
	}

	// Verify default config
	if cfg.DiscordWebhookURL != "" {
		t.Errorf("Expected empty webhook URL for default config, got %q", cfg.DiscordWebhookURL)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML
	invalidYAML := `discord_webhook_url: [invalid: yaml`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config (should fail)
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_EnvironmentVariable(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yaml")

	// Write test config
	testConfig := `discord_webhook_url: "https://discord.com/api/webhooks/env-test/xyz"`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set environment variable
	os.Setenv("CONFIG_PATH", configPath)
	defer os.Unsetenv("CONFIG_PATH")

	// Load config (should use env var path)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	expected := "https://discord.com/api/webhooks/env-test/xyz"
	if cfg.DiscordWebhookURL != expected {
		t.Errorf("Expected webhook URL %q, got %q", expected, cfg.DiscordWebhookURL)
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	// Ensure CONFIG_PATH is not set
	os.Unsetenv("CONFIG_PATH")

	// Create config.yaml in current directory for test
	testConfig := `discord_webhook_url: "https://discord.com/api/webhooks/default/test"`
	if err := os.WriteFile("./config.yaml", []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("./config.yaml")

	// Load config with empty path (should default to ./config.yaml)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	expected := "https://discord.com/api/webhooks/default/test"
	if cfg.DiscordWebhookURL != expected {
		t.Errorf("Expected webhook URL %q, got %q", expected, cfg.DiscordWebhookURL)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config
	cfg := &Config{
		DiscordWebhookURL: "https://discord.com/api/webhooks/save-test/abc123",
	}

	// Save config
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load it back
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	if loadedCfg.DiscordWebhookURL != cfg.DiscordWebhookURL {
		t.Errorf("Expected webhook URL %q, got %q", cfg.DiscordWebhookURL, loadedCfg.DiscordWebhookURL)
	}
}

func TestSaveConfig_EnvironmentVariable(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-save-config.yaml")

	// Set environment variable
	os.Setenv("CONFIG_PATH", configPath)
	defer os.Unsetenv("CONFIG_PATH")

	// Create config
	cfg := &Config{
		DiscordWebhookURL: "https://discord.com/api/webhooks/env-save/xyz789",
	}

	// Save config (should use env var path)
	if err := SaveConfig("", cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created at env var path
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created at CONFIG_PATH location")
	}

	// Load it back
	loadedCfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	if loadedCfg.DiscordWebhookURL != cfg.DiscordWebhookURL {
		t.Errorf("Expected webhook URL %q, got %q", cfg.DiscordWebhookURL, loadedCfg.DiscordWebhookURL)
	}
}

func TestValidate_ValidDiscordWebhook(t *testing.T) {
	tests := []struct {
		name       string
		webhookURL string
		shouldPass bool
	}{
		{
			name:       "Valid discord.com webhook",
			webhookURL: "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
			shouldPass: true,
		},
		{
			name:       "Valid discordapp.com webhook",
			webhookURL: "https://discordapp.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
			shouldPass: true,
		},
		{
			name:       "Empty webhook (allowed)",
			webhookURL: "",
			shouldPass: true,
		},
		{
			name:       "Invalid webhook - wrong domain",
			webhookURL: "https://example.com/api/webhooks/123/abc",
			shouldPass: false,
		},
		{
			name:       "Invalid webhook - http instead of https",
			webhookURL: "http://discord.com/api/webhooks/123/abc",
			shouldPass: false,
		},
		{
			name:       "Invalid webhook - wrong path",
			webhookURL: "https://discord.com/wrong/path/123/abc",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DiscordWebhookURL: tt.webhookURL,
			}

			err := cfg.Validate()
			if tt.shouldPass && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}
			if !tt.shouldPass && err == nil {
				t.Error("Expected validation to fail, got nil error")
			}
		})
	}
}

func TestLoadConfig_WithValidation(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config with valid webhook
	testConfig := `discord_webhook_url: "https://discord.com/api/webhooks/123456/abcdef"`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validation failed for valid config: %v", err)
	}
}

func TestRoundTrip(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "roundtrip-config.yaml")

	// Create original config
	original := &Config{
		DiscordWebhookURL: "https://discord.com/api/webhooks/roundtrip/test123",
	}

	// Save
	if err := SaveConfig(configPath, original); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Validate
	if err := loaded.Validate(); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Verify
	if loaded.DiscordWebhookURL != original.DiscordWebhookURL {
		t.Errorf("Round trip failed: expected %q, got %q", original.DiscordWebhookURL, loaded.DiscordWebhookURL)
	}
}
