# lazywork

AI-powered Git workflow automation tool that handles the entire development workflow intelligently.

## Vision

LazyWork goes beyond simple commit message generation. It's an intelligent Git workflow assistant that:

- **Auto-generates commit messages** with AI understanding of your changes
- **Manages branches intelligently** with AI-generated semantic names
- **Separates features automatically** by analyzing code changes and grouping related modifications
- **Creates atomic commits** by understanding feature boundaries and dependencies
- **Integrates with CLI agents** like Claude Code for seamless workflow automation
- **Manages Git worktrees** automatically for parallel development streams

## Features

- **Smart Commit Generation**: AI analyzes diffs to create meaningful, conventional commit messages
- **Intelligent Branch Management**: Auto-generates branch names based on feature analysis
- **Feature Separation**: AI identifies distinct features in your changes and creates separate commits
- **Hook Integration**: Works with git hooks and CLI agents for automated workflows
- **Multiple AI Providers**: OpenAI (GPT-4o) and Anthropic (Claude 3.5) support
- **Streaming & Completion**: Real-time or batch processing modes
- **Flexible Configuration**: Per-provider and per-model settings

## AI Providers

### Supported Providers

- **OpenAI**: GPT-4o, GPT-4o Mini
- **Anthropic**: Claude 3.5 Sonnet, Claude 3.5 Haiku

### Configuration

Configuration is stored in `~/.config/lazywork/config.json`. If no config exists, defaults will be used.

The AI providers are configured with models optimized for code analysis and natural language generation, essential for:
- Analyzing complex diffs and separating features
- Generating semantic branch names
- Creating contextual commit messages
- Understanding code structure and dependencies

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
export LAZYWORK_PROVIDER="anthropic"
```

### Provider Recommendations

**For Complex Analysis** (feature separation, branch naming):
- Claude 3.5 Sonnet: Excellent code understanding and reasoning
- GPT-4o: Strong at pattern recognition and semantic analysis

**For Fast Operations** (simple commits):
- Claude 3.5 Haiku: Fast with good accuracy
- GPT-4o Mini: Quick responses for straightforward tasks

## Building

```bash
go build -o lazywork .
```

## Usage

```bash
# Basic usage (coming soon)
./lazywork commit

# Auto-generate branch for current changes
./lazywork branch

# Separate features and create atomic commits
./lazywork separate

# Full auto-workflow
./lazywork auto
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