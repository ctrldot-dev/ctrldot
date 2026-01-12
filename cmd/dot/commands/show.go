package commands

import (
	"time"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <node-id>",
		Short: "Show a node with its relationships",
		Long:  "Display a node with its roles, links, and materials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			nodeID := args[0]
			depth, _ := cmd.Flags().GetInt("depth")
			asofSeq, _ := cmd.Flags().GetInt64("asof-seq")
			asofTimeStr, _ := cmd.Flags().GetString("asof-time")

			// Build expand request
			req := client.ExpandRequest{
				IDs:         []string{nodeID},
				NamespaceID: getNamespaceID(ctx.Config),
				Depth:       depth,
				AsOfSeq:     asofSeq,
			}

			// Handle asof-time
			if asofTimeStr != "" {
				t, err := time.Parse(time.RFC3339, asofTimeStr)
				if err != nil {
					return err
				}
				req.AsOfTime = &t
			}

			// Call API
			result, err := ctx.Client.Expand(req)
			if err != nil {
				handleError(err)
				return nil
			}

			return ctx.Formatter.PrintExpand(result)
		},
	}

	cmd.Flags().Int("depth", 1, "Expansion depth")
	cmd.Flags().Int64("asof-seq", 0, "As-of sequence number")
	cmd.Flags().String("asof-time", "", "As-of time (ISO format)")

	return cmd
}
