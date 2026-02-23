package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func resolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Resolution token management",
	}

	allowOnceCmd := &cobra.Command{
		Use:   "allow-once",
		Short: "Generate an allow-once resolution token",
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID, _ := cmd.Flags().GetString("agent")
			actionType, _ := cmd.Flags().GetString("action")
			ttlStr, _ := cmd.Flags().GetString("ttl")

			if agentID == "" || actionType == "" {
				return fmt.Errorf("agent and action are required")
			}

			// Parse TTL (e.g., "10m", "1h")
			ttl, err := time.ParseDuration(ttlStr)
			if err != nil {
				return fmt.Errorf("invalid TTL format: %w", err)
			}

			// TODO: Call API to generate token
			// For now, just print a placeholder
			fmt.Printf("Resolution token for agent %s, action %s, TTL %v\n", agentID, actionType, ttl)
			fmt.Printf("TODO: Implement token generation via API\n")

			return nil
		},
	}
	allowOnceCmd.Flags().String("agent", "", "Agent ID")
	allowOnceCmd.Flags().String("action", "", "Action type")
	allowOnceCmd.Flags().String("ttl", "10m", "Time to live (e.g., 10m, 1h)")

	cmd.AddCommand(allowOnceCmd)
	return cmd
}
