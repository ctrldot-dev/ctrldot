package commands

import (
	"github.com/futurematic/kernel/internal/domain"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create new resources",
	}

	nodeCmd := &cobra.Command{
		Use:   "node <title>",
		Short: "Create a new node",
		Long:  "Create a new node with the given title",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			title := args[0]
			metaFlags, _ := cmd.Flags().GetStringToString("meta")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			yes, _ := cmd.Flags().GetBool("yes")

			// Build meta map
			meta := make(map[string]interface{})
			for k, v := range metaFlags {
				meta[k] = v
			}

			// Generate node ID
			nodeID := "node:" + uuid.New().String()

			// Create intent
			intent := domain.Intent{
				Kind:        domain.IntentCreateNode,
				NamespaceID: getNamespaceID(ctx.Config),
				Payload: map[string]interface{}{
					"node_id": nodeID,
					"title":   title,
					"meta":    meta,
				},
			}

			return planAndApply(ctx, []domain.Intent{intent}, dryRun, yes)
		},
	}

	nodeCmd.Flags().StringToString("meta", nil, "Metadata key=value pairs")

	cmd.AddCommand(nodeCmd)
	return cmd
}
