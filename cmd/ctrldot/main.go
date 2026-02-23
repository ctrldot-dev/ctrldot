package main

import (
	"fmt"
	"os"

	"github.com/futurematic/kernel/cmd/ctrldot/commands"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "ctrldot",
		Short:   "Ctrl Dot CLI - Command-line interface for Ctrl Dot",
		Long:    `Ctrl Dot is a CLI for controlling agent actions via Ctrl Dot daemon.`,
		Version: version,
	}

	// Add global flags
	rootCmd.PersistentFlags().String("server", "http://127.0.0.1:7777", "Ctrl Dot server URL")
	rootCmd.PersistentFlags().Bool("json", false, "Output JSON")

	// Register commands
	commands.RegisterCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
