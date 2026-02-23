package commands

import (
	"github.com/spf13/cobra"
)

// RegisterCommands registers all CLI commands
func RegisterCommands(rootCmd *cobra.Command) {
	// Status
	rootCmd.AddCommand(statusCmd())

	// Agents
	rootCmd.AddCommand(agentsCmd())
	rootCmd.AddCommand(haltCmd())
	rootCmd.AddCommand(resumeCmd())
	
	// Agent subcommands
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent management commands",
	}
	agentCmd.AddCommand(agentShowCmd())
	rootCmd.AddCommand(agentCmd)

	// Budget
	rootCmd.AddCommand(budgetCmd())

	// Events
	rootCmd.AddCommand(eventsTailCmd())

	// Rules
	rootCmd.AddCommand(rulesShowCmd())
	rootCmd.AddCommand(rulesEditCmd())

	// Resolution
	rootCmd.AddCommand(resolveCmd())

	// Daemon
	rootCmd.AddCommand(daemonCmd())

	// Doctor
	rootCmd.AddCommand(doctorCmd())

	// Bundle
	rootCmd.AddCommand(bundleCmd())

	// Panic
	rootCmd.AddCommand(panicCmd())

	// Autobundle
	rootCmd.AddCommand(autobundleCmd())
}
