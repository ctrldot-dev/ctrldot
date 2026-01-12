package commands

import (
	"github.com/spf13/cobra"
)

// RegisterCommands registers all dot CLI commands
func RegisterCommands(rootCmd *cobra.Command) {
	// Connectivity & config
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newUseCmd())
	rootCmd.AddCommand(newWhereamiCmd())

	// Read commands
	rootCmd.AddCommand(newShowCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newLsCmd())

	// Write commands
	rootCmd.AddCommand(newNewCmd())
	rootCmd.AddCommand(newRoleCmd())
	rootCmd.AddCommand(newLinkCmd())
	rootCmd.AddCommand(newMoveCmd())
}
