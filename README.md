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

- **OpenAI**: GPT-5, GPT-5 Mini, GPT-5 Nano (latest models with 272K context window)
- **Anthropic**: Claude Sonnet 4.5, Claude Haiku 4.5, Claude Opus 4.1 (200K context window)

### Configuration

Configuration is stored in `~/.config/lazywork/config.json`. If no config exists, defaults will be used.

LazyWork uses the **latest frontier models** from OpenAI and Anthropic:

**OpenAI GPT-5** (August 2025):
- State-of-the-art coding model (74.9% on SWE-bench Verified, 88% on Aider polyglot)
- 272K input tokens, 128K output tokens
- 45% less likely to hallucinate than GPT-4o
- Available in three sizes: gpt-5, gpt-5-mini, gpt-5-nano

**Anthropic Claude 4.x** (August-October 2025):
- **Sonnet 4.5**: Best coding model in the world, optimized for agentic workflows
- **Haiku 4.5**: 90% of Sonnet 4.5 performance at 1/3 cost and 2x speed
- **Opus 4.1**: Maximum intelligence for complex reasoning
- 200K context window, 64K output tokens

These models are optimized for:
- Analyzing complex diffs and separating features
- Generating semantic branch names
- Creating contextual commit messages
- Understanding code structure and dependencies
- Agentic workflows and multi-step planning

Example configuration:

```json
{
  "default_provider": "anthropic",
  "providers": {
    "openai": {
      "type": "openai",
      "base_url": "https://api.openai.com/v1",
      "api_key": "$OPENAI_API_KEY",
      "max_tokens": 4000,
      "models": [
        {
          "id": "gpt-5",
          "name": "GPT-5",
          "context_window": 272000,
          "max_tokens": 128000,
          "temperature": 0.3
        },
        {
          "id": "gpt-5-mini",
          "name": "GPT-5 Mini",
          "context_window": 272000,
          "max_tokens": 128000,
          "temperature": 0.3
        }
      ]
    },
    "anthropic": {
      "type": "anthropic",
      "base_url": "https://api.anthropic.com/v1",
      "api_key": "$ANTHROPIC_API_KEY",
      "max_tokens": 4000,
      "models": [
        {
          "id": "claude-sonnet-4-5",
          "name": "Claude Sonnet 4.5",
          "context_window": 200000,
          "max_tokens": 64000,
          "temperature": 0.3
        },
        {
          "id": "claude-haiku-4-5",
          "name": "Claude Haiku 4.5",
          "context_window": 200000,
          "max_tokens": 64000,
          "temperature": 0.3
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

**For Maximum Intelligence** (complex feature separation, multi-step planning):
- **Claude Sonnet 4.5** ⭐ (default): Best coding model, optimized for agents
- **Claude Opus 4.1**: Maximum reasoning capabilities
- **GPT-5**: State-of-the-art coding, excellent at complex diffs

**For Speed & Efficiency** (simple commits, rapid iterations):
- **Claude Haiku 4.5**: 90% of Sonnet performance, 2x faster, 1/3 cost
- **GPT-5 Mini**: Balanced performance and speed
- **GPT-5 Nano**: Ultra-fast for simple operations

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