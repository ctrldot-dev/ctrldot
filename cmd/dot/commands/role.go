package commands

import (
	"github.com/futurematic/kernel/internal/domain"
	"github.com/spf13/cobra"
)

func newRoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role",
		Short: "Manage role assignments",
	}

	assignCmd := &cobra.Command{
		Use:   "assign <node-id> <role>",
		Short: "Assign a role to a node",
		Long:  "Assign a role to a node within the current namespace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			nodeID := args[0]
			role := args[1]
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			yes, _ := cmd.Flags().GetBool("yes")

			intent := domain.Intent{
				Kind:        domain.IntentAssignRole,
				NamespaceID: getNamespaceID(ctx.Config),
				Payload: map[string]interface{}{
					"node_id": nodeID,
					"role":    role,
				},
			}

			return planAndApply(ctx, []domain.Intent{intent}, dryRun, yes)
		},
	}

	cmd.AddCommand(assignCmd)
	return cmd
}
