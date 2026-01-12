package main

import (
	"fmt"
	"os"

	"github.com/futurematic/kernel/cmd/dot/commands"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dot",
		Short: "Dot CLI - Command-line interface for Futurematic Kernel",
		Long:  `Dot is a git-like CLI for interacting with the Futurematic Kernel.`,
		Version: version,
	}

	// Add global flags
	rootCmd.PersistentFlags().String("server", "", "Override server URL")
	rootCmd.PersistentFlags().String("actor", "", "Override actor ID")
	rootCmd.PersistentFlags().String("ns", "", "Override namespace")
	rootCmd.PersistentFlags().String("cap", "", "Override capabilities (comma-separated)")
	rootCmd.PersistentFlags().Bool("json", false, "Output JSON")
	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "Plan only (mutations)")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation (mutations)")

	// Register commands
	commands.RegisterCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
