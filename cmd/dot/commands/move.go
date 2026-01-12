package commands

import (
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/spf13/cobra"
)

func newMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <child> --to <parent>",
		Short: "Move a node to a new parent",
		Long:  "Move a child node to a new parent by updating PARENT_OF links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			childID := args[0]
			parentID, _ := cmd.Flags().GetString("to")
			if parentID == "" {
				return fmt.Errorf("--to flag is required")
			}
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			yes, _ := cmd.Flags().GetBool("yes")

			// Check if kernel supports Move intent
			// For now, we'll use Move intent if available
			// Otherwise, we'd need to: expand to find current parent, retire old link, create new link
			
			// Try Move intent first (kernel may support it)
			intent := domain.Intent{
				Kind:        domain.IntentMove,
				NamespaceID: getNamespaceID(ctx.Config),
				Payload: map[string]interface{}{
					"node_id":      childID,
					"to_parent_id": parentID,
				},
			}

			// If Move doesn't work, we'll need to implement the retire+create approach
			// For now, try Move and let the kernel handle it
			return planAndApply(ctx, []domain.Intent{intent}, dryRun, yes)
		},
	}

	cmd.Flags().String("to", "", "New parent node ID")
	cmd.MarkFlagRequired("to")

	return cmd
}
