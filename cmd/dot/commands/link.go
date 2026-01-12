package commands

import (
	"github.com/futurematic/kernel/internal/domain"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func newLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link <from> <type> <to>",
		Short: "Create a link between nodes",
		Long:  "Create a typed link from one node to another",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			fromNodeID := args[0]
			linkType := args[1]
			toNodeID := args[2]
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			yes, _ := cmd.Flags().GetBool("yes")

			linkID := "link:" + uuid.New().String()

			intent := domain.Intent{
				Kind:        domain.IntentCreateLink,
				NamespaceID: getNamespaceID(ctx.Config),
				Payload: map[string]interface{}{
					"link_id":      linkID,
					"from_node_id": fromNodeID,
					"to_node_id":   toNodeID,
					"type":         linkType,
				},
			}

			return planAndApply(ctx, []domain.Intent{intent}, dryRun, yes)
		},
	}

	return cmd
}
