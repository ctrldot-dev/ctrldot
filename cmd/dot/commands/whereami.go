package commands

import (
	"github.com/spf13/cobra"
)

func newWhereamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whereami",
		Short: "Show resolved configuration",
		Long:  "Display the current configuration including environment overrides",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			return ctx.Formatter.PrintConfig(ctx.Config)
		},
	}

	return cmd
}
