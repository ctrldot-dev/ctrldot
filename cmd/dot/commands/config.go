package commands

import (
	"fmt"

	"github.com/futurematic/kernel/cmd/dot/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Get or set configuration values",
	}

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value, err := config.Get(key)
			if err != nil {
				return err
			}

			json, _ := cmd.Flags().GetBool("json")
			if json {
				fmt.Printf(`{"%s": "%s"}\n`, key, value)
			} else {
				fmt.Println(value)
			}
			return nil
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			if err := config.Set(key, value); err != nil {
				return err
			}

			json, _ := cmd.Flags().GetBool("json")
			if json {
				fmt.Printf(`{"%s": "%s"}\n`, key, value)
			} else {
				fmt.Printf("Set %s = %s\n", key, value)
			}
			return nil
		},
	}

	cmd.AddCommand(getCmd)
	cmd.AddCommand(setCmd)

	return cmd
}
