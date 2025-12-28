package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for LazyWork.

Bash:
  # Add to ~/.bashrc:
  source <(lazywork completion bash)

  # Or save to a file:
  lazywork completion bash > /etc/bash_completion.d/lazywork

Zsh:
  # Add to ~/.zshrc (before compinit):
  source <(lazywork completion zsh)

  # Or if using oh-my-zsh, save to completions folder:
  lazywork completion zsh > ~/.oh-my-zsh/completions/_lazywork

Fish:
  # Add to ~/.config/fish/config.fish:
  lazywork completion fish | source

  # Or save to completions folder:
  lazywork completion fish > ~/.config/fish/completions/lazywork.fish`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	}
	return nil
}
