package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version information including build metadata.",
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	info := map[string]string{
		"version":   Version,
		"commit":    Commit,
		"buildDate": BuildDate,
		"go":        runtime.Version(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}

	if jsonOutput {
		enc := json.NewEncoder(Stdout())
		enc.SetIndent("", "  ")
		enc.Encode(info)
		return
	}

	fmt.Printf("lazywork %s\n", Version)
	if Version == "dev" {
		fmt.Println("  (development build)")
	}
	fmt.Printf("  commit:  %s\n", Commit)
	fmt.Printf("  built:   %s\n", BuildDate)
	fmt.Printf("  go:      %s\n", runtime.Version())
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
