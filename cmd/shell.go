package cmd

import (
	"fmt"
	"strings"

	"github.com/miltonparedes/lazywork/internal/output"
	"github.com/miltonparedes/lazywork/internal/shell"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Shell integration commands",
	Long:  "Commands for integrating LazyWork with your shell (bash, zsh, fish).",
}

var shellInitCmd = &cobra.Command{
	Use:   "init [bash|zsh|fish]",
	Short: "Print shell initialization script",
	Long: `Print the shell initialization script for LazyWork.

This script sets up:
  - lw:  alias for 'lazywork'
  - lwt: alias for 'lazywork worktree'

The script also handles special commands like 'worktree go' that need
to change the current directory.

Usage:
  # Auto-detect shell and print script
  lazywork shell init

  # Specify shell explicitly
  lazywork shell init fish

Setup:
  # Bash - add to ~/.bashrc:
  eval "$(lazywork shell init bash)"

  # Zsh - add to ~/.zshrc:
  eval "$(lazywork shell init zsh)"

  # Fish - add to ~/.config/fish/config.fish:
  lazywork shell init fish | source`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: shell.SupportedShells(),
	RunE:      runShellInit,
}

var shellStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check shell integration status",
	RunE:  runShellStatus,
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellInitCmd)
	shellCmd.AddCommand(shellStatusCmd)
}

func runShellInit(cmd *cobra.Command, args []string) error {
	var shellType string

	if len(args) > 0 {
		shellType = strings.ToLower(args[0])
		if !shell.IsValidShell(shellType) {
			return fmt.Errorf("unsupported shell '%s'. Supported: %s",
				shellType, strings.Join(shell.SupportedShells(), ", "))
		}
	} else {
		shellType = shell.DetectShell()
	}

	script := shell.InitScript(shellType)
	fmt.Print(script)

	return nil
}

func runShellStatus(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)
	shellType := shell.DetectShell()

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"shell":     shellType,
			"rc_file":   shell.RcFile(shellType),
			"installed": shell.HasInitLine(shellType),
		})
	}

	out.Bold("Shell Integration Status")
	out.Println()

	out.Print("  Shell:   %s\n", shellType)
	out.Print("  RC file: %s\n", shell.RcFile(shellType))

	if shell.HasInitLine(shellType) {
		out.Success("LazyWork integration is installed")
	} else {
		out.Warning("LazyWork integration not found in RC file")
		out.Println()
		out.Info("Add this to " + shell.RcFile(shellType) + ":")
		out.Dim("  " + shell.InitLine(shellType))
	}

	return nil
}
