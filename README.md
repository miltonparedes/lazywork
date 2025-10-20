# lazycommit

AI-powered commit message generator with support for multiple AI providers.

## Features

- Multiple AI provider support (OpenAI, Anthropic)
- Streaming and completion modes
- Configurable models and parameters
- Environment-based API key management

## AI Providers

### Supported Providers

- **OpenAI**: GPT-4o, GPT-4o Mini
- **Anthropic**: Claude 3.5 Sonnet, Claude 3.5 Haiku

### Configuration

Configuration is stored in `~/.config/lazycommit/config.json`. If no config exists, defaults will be used.

Example configuration:

```json
{
  "default_provider": "openai",
  "providers": {
    "openai": {
      "type": "openai",
      "base_url": "https://api.openai.com/v1",
      "api_key": "$OPENAI_API_KEY",
      "max_tokens": 1000,
      "models": [
        {
          "id": "gpt-4o",
          "name": "GPT-4o",
          "context_window": 128000,
          "max_tokens": 4096,
          "temperature": 0.7
        }
      ]
    },
    "anthropic": {
      "type": "anthropic",
      "base_url": "https://api.anthropic.com/v1",
      "api_key": "$ANTHROPIC_API_KEY",
      "max_tokens": 1000,
      "models": [
        {
          "id": "claude-3-5-sonnet-20241022",
          "name": "Claude 3.5 Sonnet",
          "context_window": 200000,
          "max_tokens": 8192,
          "temperature": 0.7
        }
      ]
    }
  }
}
```

### Environment Variables

Set your API keys as environment variables:

```bash
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
```

Switch providers:

```bash
export LAZYCOMMIT_PROVIDER="anthropic"
```

## Building

```bash
go build -o lazycommit .
```

## Usage

```bash
./lazycommit
```

## Architecture

The codebase follows Go best practices with a modular provider system:

- `pkg/types`: Common types and Provider interface
- `pkg/config`: Configuration management
- `pkg/provider`: Provider implementations (OpenAI, Anthropic)

### Provider Interface

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    Name() string
    Models() []string
}
```