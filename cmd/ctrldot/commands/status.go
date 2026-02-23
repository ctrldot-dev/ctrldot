package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Ctrl Dot daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			outputJSON, _ := cmd.Flags().GetBool("json")

			resp, err := http.Get(serverURL + "/v1/health")
			if err != nil {
				if outputJSON {
					json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
						"ok":      false,
						"error":   err.Error(),
						"version": "0.1.0",
					})
				} else {
					fmt.Printf("Daemon: not running (%v)\n", err)
					fmt.Printf("Version: 0.1.0\n")
				}
				return nil
			}
			defer resp.Body.Close()

			var health map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
				return err
			}

			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(health)
			} else {
				fmt.Printf("Daemon: running\n")
				fmt.Printf("Version: %v\n", health["version"])
				fmt.Printf("Server: %s\n", serverURL)
			}

			return nil
		},
	}
	return cmd
}
