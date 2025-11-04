package provider

import (
	"fmt"

	"github.com/miltonparedes/lazywork/pkg/config"
	"github.com/miltonparedes/lazywork/pkg/types"
)

func New(name string, cfg config.Provider) (types.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for provider %s", name)
	}

	switch cfg.Type {
	case "openai":
		return NewOpenAI(cfg), nil
	case "anthropic":
		return NewAnthropic(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
}

func NewFromConfig(cfg *config.Config, providerName string) (types.Provider, error) {
	if providerName == "" {
		providerName = cfg.DefaultProvider
	}

	providerCfg, ok := cfg.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not found in configuration", providerName)
	}

	return New(providerName, providerCfg)
}
