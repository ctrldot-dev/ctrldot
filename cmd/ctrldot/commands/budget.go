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
		Short: "Show agent budget/limits status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			outputJSON, _ := cmd.Flags().GetBool("json")
			agentID := args[0]

			resp, err := http.Get(serverURL + "/v1/agents/" + agentID + "/limits")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("agent limits: %s", resp.Status)
			}

			var lim map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&lim); err != nil {
				return err
			}

			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(lim)
			} else {
				spent, _ := lim["spent_gbp"].(float64)
				limit, _ := lim["limit_gbp"].(float64)
				pct, _ := lim["percentage"].(float64)
				fmt.Printf("Budget for agent %s (daily):\n", agentID)
				fmt.Printf("  Spent: £%.2f\n", spent)
				fmt.Printf("  Limit: £%.2f\n", limit)
				fmt.Printf("  Usage: %.1f%%\n", pct*100)
			}

			return nil
		},
	}
	return cmd
}
