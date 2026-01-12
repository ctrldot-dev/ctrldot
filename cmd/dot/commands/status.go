package commands

import (
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check kernel server status",
		Long:  "Check the health of the kernel server and display current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			// Check health
			health, err := ctx.Client.Healthz()
			if err != nil {
				handleError(err)
				return nil
			}

			// Prepare status response (use struct for type safety)
			type Status struct {
				OK          bool   `json:"ok"`
				Server      string `json:"server"`
				ActorID     string `json:"actor_id"`
				NamespaceID string `json:"namespace_id"`
			}
			status := &Status{
				OK:          health.OK,
				Server:      ctx.Config.Server,
				ActorID:     ctx.Config.ActorID,
				NamespaceID: ctx.Config.NamespaceID,
			}

			return ctx.Formatter.PrintStatus(status)
		},
	}

	return cmd
}
