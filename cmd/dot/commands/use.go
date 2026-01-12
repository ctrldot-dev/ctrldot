package commands

import (
	"fmt"

	"github.com/futurematic/kernel/cmd/dot/config"
	"github.com/spf13/cobra"
)

func newUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <namespace>",
		Short: "Set the active namespace",
		Long:  "Set the namespace_id in configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := args[0]
			if err := config.Set("namespace_id", namespace); err != nil {
				return err
			}

			json, _ := cmd.Flags().GetBool("json")
			if json {
				fmt.Printf(`{"namespace_id": "%s"}\n`, namespace)
			} else {
				fmt.Printf("Using namespace: %s\n", namespace)
			}
			return nil
		},
	}

	return cmd
}
