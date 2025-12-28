package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	Bash = "bash"
	Zsh  = "zsh"
	Fish = "fish"
)

func DetectShell() string {
	shell := os.Getenv("SHELL")
	base := filepath.Base(shell)

	switch base {
	case "fish":
		return Fish
	case "zsh":
		return Zsh
	default:
		return Bash
	}
}

func IsValidShell(shell string) bool {
	switch shell {
	case Bash, Zsh, Fish:
		return true
	default:
		return false
	}
}

func SupportedShells() []string {
	return []string{Bash, Zsh, Fish}
}

func InitScript(shell string) string {
	switch shell {
	case Fish:
		return fishScript
	case Zsh:
		return zshScript
	case Bash:
		return bashScript
	default:
		return bashScript
	}
}

func CompletionInstructions(shell string) string {
	switch shell {
	case Fish:
		return `# Add to ~/.config/fish/config.fish:
lazywork completion fish | source`
	case Zsh:
		return `# Add to ~/.zshrc (before compinit):
source <(lazywork completion zsh)`
	case Bash:
		return `# Add to ~/.bashrc:
source <(lazywork completion bash)`
	default:
		return ""
	}
}

const bashScript = `# LazyWork shell integration
# Add to ~/.bashrc: eval "$(lazywork shell init bash)"

# Wrapper function that handles cd commands from lazywork
__lazywork_exec() {
  local output
  output=$(command lazywork "$@" --shell-helper 2>&1)
  local exit_code=$?

  if [[ $output == cd\ * ]]; then
    eval "$output"
  else
    printf '%s\n' "$output"
    return $exit_code
  fi
}

# Aliases
alias lw='__lazywork_exec'
alias lwt='__lazywork_exec worktree'
`

const zshScript = `# LazyWork shell integration
# Add to ~/.zshrc: eval "$(lazywork shell init zsh)"

# Wrapper function that handles cd commands from lazywork
__lazywork_exec() {
  local output
  output=$(command lazywork "$@" --shell-helper 2>&1)
  local exit_code=$?

  if [[ $output == cd\ * ]]; then
    eval "$output"
  else
    printf '%s\n' "$output"
    return $exit_code
  fi
}

# Aliases
alias lw='__lazywork_exec'
alias lwt='__lazywork_exec worktree'
`

const fishScript = `# LazyWork shell integration
# Add to ~/.config/fish/config.fish: lazywork shell init fish | source

# Wrapper function that handles cd commands from lazywork
function __lazywork_exec
    set -l output (command lazywork $argv --shell-helper 2>&1)
    set -l exit_code $status

    if string match -q 'cd *' -- $output
        eval $output
    else
        printf '%s\n' $output
        return $exit_code
    end
end

# Aliases
alias lw='__lazywork_exec'
alias lwt='__lazywork_exec worktree'
`

func RcFile(shell string) string {
	home, _ := os.UserHomeDir()

	switch shell {
	case Fish:
		return filepath.Join(home, ".config", "fish", "config.fish")
	case Zsh:
		return filepath.Join(home, ".zshrc")
	case Bash:
		return filepath.Join(home, ".bashrc")
	default:
		return filepath.Join(home, ".bashrc")
	}
}

func InitLine(shell string) string {
	switch shell {
	case Fish:
		return "lazywork shell init fish | source"
	default:
		return fmt.Sprintf(`eval "$(lazywork shell init %s)"`, shell)
	}
}

func HasInitLine(shell string) bool {
	rcFile := RcFile(shell)
	content, err := os.ReadFile(rcFile)
	if err != nil {
		return false
	}

	initLine := InitLine(shell)
	return strings.Contains(string(content), "lazywork shell init") ||
		strings.Contains(string(content), initLine)
}
