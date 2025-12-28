package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/miltonparedes/lazywork/internal/output"
	"github.com/miltonparedes/lazywork/pkg/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage LazyWork configuration",
	Long:  "View, modify, or initialize LazyWork configuration.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	RunE:  runConfigPath,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default configuration file",
	RunE:  runConfigInit,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value. Supported keys:
  - default_provider: Set the default AI provider (openai, anthropic)
  - worktree_dir: Set the directory for worktrees (default: .worktrees)`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
}

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return config.DefaultConfigPath()
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	cfg, err := config.LoadFrom(cfgFile)
	if err != nil {
		out.ErrorResult(err, "CONFIG_LOAD_ERROR")
		return err
	}

	configPath := getConfigPath()
	exists := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exists = false
	}

	if jsonOutput {
		providers := make([]string, 0, len(cfg.Providers))
		for name := range cfg.Providers {
			providers = append(providers, name)
		}

		return out.JSON(map[string]interface{}{
			"path":             configPath,
			"exists":           exists,
			"default_provider": cfg.DefaultProvider,
			"worktree_dir":     cfg.GetWorktreeDir(),
			"providers":        providers,
			"config":           cfg,
		})
	}

	out.Bold("LazyWork Configuration")
	out.Println()

	out.Print("  Path: %s\n", configPath)
	if exists {
		out.Success("Config file exists")
	} else {
		out.Dim("  (using defaults, no config file)")
	}
	out.Println()

	out.Print("  Default Provider: %s\n", cfg.DefaultProvider)
	out.Print("  Worktree Dir:     %s\n", cfg.GetWorktreeDir())
	out.Println()

	out.Bold("Providers:")
	for name, provider := range cfg.Providers {
		marker := "  "
		if name == cfg.DefaultProvider {
			marker = "→ "
		}
		out.Print("%s%s (%s)\n", marker, name, provider.Type)
		for _, model := range provider.Models {
			out.Dim(fmt.Sprintf("    • %s (%s)", model.Name, model.ID))
		}
	}

	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)
	configPath := getConfigPath()

	exists := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exists = false
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"path":   configPath,
			"exists": exists,
		})
	}

	out.Println(configPath)
	if !exists {
		out.Dim("(file does not exist)")
	}

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		err := fmt.Errorf("config file already exists at %s", configPath)
		out.ErrorResult(err, "CONFIG_EXISTS")
		return err
	}

	cfg, err := config.LoadFrom(cfgFile)
	if err != nil {
		out.ErrorResult(err, "CONFIG_LOAD_ERROR")
		return err
	}

	if err := cfg.SaveTo(cfgFile); err != nil {
		out.ErrorResult(err, "CONFIG_SAVE_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"path":    configPath,
			"created": true,
		})
	}

	out.Success(fmt.Sprintf("Created config file at %s", configPath))
	out.Dim("Edit this file to add your API keys and customize settings.")

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	key := args[0]
	value := args[1]

	cfg, err := config.LoadFrom(cfgFile)
	if err != nil {
		out.ErrorResult(err, "CONFIG_LOAD_ERROR")
		return err
	}

	switch strings.ToLower(key) {
	case "default_provider":
		if _, exists := cfg.Providers[value]; !exists {
			validProviders := make([]string, 0, len(cfg.Providers))
			for name := range cfg.Providers {
				validProviders = append(validProviders, name)
			}
			err := fmt.Errorf("unknown provider '%s'. Valid providers: %s", value, strings.Join(validProviders, ", "))
			out.ErrorResult(err, "INVALID_PROVIDER")
			return err
		}
		cfg.DefaultProvider = value

	case "worktree_dir":
		cfg.WorktreeDir = value

	default:
		err := fmt.Errorf("unknown config key '%s'. Supported keys: default_provider, worktree_dir", key)
		out.ErrorResult(err, "INVALID_KEY")
		return err
	}

	if err := cfg.SaveTo(cfgFile); err != nil {
		out.ErrorResult(err, "CONFIG_SAVE_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"key":     key,
			"value":   value,
			"updated": true,
		})
	}

	out.Success(fmt.Sprintf("Set %s = %s", key, value))

	return nil
}
