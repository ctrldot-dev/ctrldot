package commands

import (
	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history <target>",
		Short: "Get operation history for a target",
		Long:  "Display operation history for a node or namespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			target := args[0]
			limit, _ := cmd.Flags().GetInt("limit")

			req := client.HistoryRequest{
				Target: target,
				Limit:  limit,
			}

			ops, err := ctx.Client.History(req)
			if err != nil {
				handleError(err)
				return nil
			}

			return ctx.Formatter.PrintHistory(ops)
		},
	}

	cmd.Flags().Int("limit", 100, "Limit number of operations")

	return cmd
}
