# LazyWork

Git workflow automation tool with worktree management, shell integration, and AI-powered features.

> **Status**: Alpha - Worktree management and shell integration are functional. AI-powered commit and branch features coming soon.

## Installation

```bash
# From releases
curl -fsSL https://raw.githubusercontent.com/miltonparedes/lazywork/main/scripts/install.sh | bash

# Specific version
LAZYWORK_VERSION=v0.1.0-alpha curl -fsSL https://raw.githubusercontent.com/miltonparedes/lazywork/main/scripts/install.sh | bash

# From source
go install github.com/miltonparedes/lazywork@latest
```

## Shell Integration

Add to your shell config for aliases and auto-cd support:

```bash
# Bash (~/.bashrc)
eval "$(lazywork shell init bash)"

# Zsh (~/.zshrc)
eval "$(lazywork shell init zsh)"

# Fish (~/.config/fish/config.fish)
lazywork shell init fish | source
```

This enables:
- `lw` - alias for `lazywork`
- `lwt` - alias for `lazywork worktree`
- Auto-cd when using `lwt go`

## Worktree Management

Simplified Git worktree workflow for parallel development.

```bash
# List worktrees
lwt list

# Create new worktree with branch
lwt add feature-auth

# Navigate to worktree (requires shell integration)
lwt go feature-auth

# Work on worktree branch from main repo
lwt use feature-auth
# ... work ...
lwt return

# Merge and cleanup
lwt finish feature-auth

# Remove worktree
lwt remove feature-auth
```

### Commands

| Command | Description |
|---------|-------------|
| `lwt list` | List all worktrees |
| `lwt add <name>` | Create worktree with new branch |
| `lwt go <name>` | Navigate to worktree directory |
| `lwt use <name>` | Checkout worktree branch in main repo |
| `lwt return` | Return to previous branch after `use` |
| `lwt finish <name>` | Merge branch and optionally cleanup |
| `lwt remove <name>` | Remove worktree |
| `lwt prune` | Clean stale worktree entries |

## Configuration

Config path: `~/.config/lazywork/config.json`

```bash
# View config
lazywork config show

# Set worktree directory (default: .worktrees)
lazywork config set worktree_dir .worktrees
```

## Roadmap

AI-powered features planned:

- **Smart commits**: AI-generated commit messages from diffs
- **Branch naming**: Semantic branch names based on changes
- **Feature separation**: Automatic atomic commits by analyzing code boundaries
- **Multi-provider support**: OpenAI and Anthropic backends

## Architecture

Modular provider system for AI backends:

```
pkg/types     - Provider interface and common types
pkg/config    - Configuration management
pkg/provider  - OpenAI and Anthropic implementations
internal/git  - Git operations wrapper
internal/tui  - Interactive forms (huh)
```

## Building

```bash
go build -o lazywork .
go test ./...
gofumpt -w .
```
