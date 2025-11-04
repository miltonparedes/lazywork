package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DefaultProvider string              `json:"default_provider"`
	Providers       map[string]Provider `json:"providers"`
}

type Provider struct {
	Type      string  `json:"type"`
	BaseURL   string  `json:"base_url,omitempty"`
	APIKey    string  `json:"api_key,omitempty"`
	Models    []Model `json:"models,omitempty"`
	MaxTokens int     `json:"max_tokens,omitempty"`
}

type Model struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContextWindow int     `json:"context_window"`
	MaxTokens     int     `json:"max_tokens,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "lazywork", "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	resolveEnvironmentVariables(&cfg)

	return &cfg, nil
}

func resolveEnvironmentVariables(cfg *Config) {
	for name, provider := range cfg.Providers {
		if len(provider.APIKey) > 0 && provider.APIKey[0] == '$' {
			envVar := provider.APIKey[1:]
			provider.APIKey = os.Getenv(envVar)
			cfg.Providers[name] = provider
		}
	}
}

func getDefaultConfig() *Config {
	return &Config{
		DefaultProvider: "anthropic",
		Providers: map[string]Provider{
			"openai": {
				Type:      "openai",
				BaseURL:   "https://api.openai.com/v1",
				APIKey:    "$OPENAI_API_KEY",
				MaxTokens: 2000,
				Models: []Model{
					{
						ID:            "gpt-4o",
						Name:          "GPT-4o",
						ContextWindow: 128000,
						MaxTokens:     4096,
						Temperature:   0.3,
					},
					{
						ID:            "gpt-4o-mini",
						Name:          "GPT-4o Mini",
						ContextWindow: 128000,
						MaxTokens:     16384,
						Temperature:   0.3,
					},
				},
			},
			"anthropic": {
				Type:      "anthropic",
				BaseURL:   "https://api.anthropic.com/v1",
				APIKey:    "$ANTHROPIC_API_KEY",
				MaxTokens: 2000,
				Models: []Model{
					{
						ID:            "claude-3-5-sonnet-20241022",
						Name:          "Claude 3.5 Sonnet",
						ContextWindow: 200000,
						MaxTokens:     8192,
						Temperature:   0.3,
					},
					{
						ID:            "claude-3-5-haiku-20241022",
						Name:          "Claude 3.5 Haiku",
						ContextWindow: 200000,
						MaxTokens:     8192,
						Temperature:   0.3,
					},
				},
			},
		},
	}
}

func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "lazywork")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
