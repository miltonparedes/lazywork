package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version info (injected via ldflags)
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"

	// Global flags
	jsonOutput  bool
	noColor     bool
	cfgFile     string
	shellHelper bool
)

var rootCmd = &cobra.Command{
	Use:   "lazywork",
	Short: "AI-powered Git workflow automation",
	Long: `LazyWork automates your Git workflow using AI.

Generate commit messages, manage worktrees, separate features,
and more - all powered by AI providers like OpenAI and Anthropic.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (agent-friendly)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file path (default ~/.config/lazywork/config.json)")
	rootCmd.PersistentFlags().BoolVar(&shellHelper, "shell-helper", false, "Output for shell function evaluation (used by lw function)")
	rootCmd.PersistentFlags().MarkHidden("shell-helper")
}

func IsJSONOutput() bool {
	return jsonOutput
}

func IsNoColor() bool {
	return noColor
}

func ConfigFile() string {
	return cfgFile
}

func IsShellHelper() bool {
	return shellHelper
}

func Stdout() *os.File {
	return os.Stdout
}

func Stderr() *os.File {
	return os.Stderr
}
