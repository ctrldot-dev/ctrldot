package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func budgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget <agent_id>",
		Short: "Show agent budget status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			outputJSON, _ := cmd.Flags().GetBool("json")
			agentID := args[0]

			// TODO: Implement budget endpoint in API
			resp, err := http.Get(serverURL + "/v1/agents/" + agentID + "/budget")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var budget map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&budget); err != nil {
				return err
			}

			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(budget)
			} else {
				fmt.Printf("Budget for agent %s:\n", agentID)
				fmt.Printf("  Spent: £%.2f\n", budget["spent_gbp"])
				fmt.Printf("  Limit: £%.2f\n", budget["limit_gbp"])
				fmt.Printf("  Percentage: %.1f%%\n", budget["percentage"])
			}

			return nil
		},
	}
	return cmd
}
