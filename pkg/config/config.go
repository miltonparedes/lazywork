package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DefaultProvider string              `json:"default_provider"`
	WorktreeDir     string              `json:"worktree_dir,omitempty"`
	Providers       map[string]Provider `json:"providers"`
}

// GetWorktreeDir returns the worktree directory, defaulting to ".worktrees"
func (c *Config) GetWorktreeDir() string {
	if c.WorktreeDir == "" {
		return ".worktrees"
	}
	return c.WorktreeDir
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

func DefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "lazywork", "config.json")
}

func Load() (*Config, error) {
	return LoadFrom("")
}

// LoadFrom loads config from a custom path (empty string uses default)
func LoadFrom(customPath string) (*Config, error) {
	configPath := customPath
	if configPath == "" {
		configPath = DefaultConfigPath()
	}

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
				MaxTokens: 4000,
				Models: []Model{
					{
						ID:            "gpt-5",
						Name:          "GPT-5",
						ContextWindow: 272000,
						MaxTokens:     128000,
						Temperature:   0.3,
					},
					{
						ID:            "gpt-5-mini",
						Name:          "GPT-5 Mini",
						ContextWindow: 272000,
						MaxTokens:     128000,
						Temperature:   0.3,
					},
					{
						ID:            "gpt-5-nano",
						Name:          "GPT-5 Nano",
						ContextWindow: 272000,
						MaxTokens:     128000,
						Temperature:   0.3,
					},
				},
			},
			"anthropic": {
				Type:      "anthropic",
				BaseURL:   "https://api.anthropic.com/v1",
				APIKey:    "$ANTHROPIC_API_KEY",
				MaxTokens: 4000,
				Models: []Model{
					{
						ID:            "claude-sonnet-4-5",
						Name:          "Claude Sonnet 4.5",
						ContextWindow: 200000,
						MaxTokens:     64000,
						Temperature:   0.3,
					},
					{
						ID:            "claude-haiku-4-5",
						Name:          "Claude Haiku 4.5",
						ContextWindow: 200000,
						MaxTokens:     64000,
						Temperature:   0.3,
					},
					{
						ID:            "claude-opus-4-1",
						Name:          "Claude Opus 4.1",
						ContextWindow: 200000,
						MaxTokens:     64000,
						Temperature:   0.3,
					},
				},
			},
		},
	}
}

func (c *Config) Save() error {
	return c.SaveTo("")
}

// SaveTo saves config to a custom path (empty string uses default)
func (c *Config) SaveTo(customPath string) error {
	configPath := customPath
	if configPath == "" {
		configPath = DefaultConfigPath()
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
