package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func eventsTailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events tail",
		Short: "Tail events feed",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			outputJSON, _ := cmd.Flags().GetBool("json")
			agentID, _ := cmd.Flags().GetString("agent")
			limit, _ := cmd.Flags().GetInt("n")

			url := serverURL + "/v1/events?limit=" + strconv.Itoa(limit)
			if agentID != "" {
				url += "&agent_id=" + agentID
			}

			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var events []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
				return err
			}

			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(events)
			} else {
				fmt.Printf("Events:\n")
				for _, event := range events {
					eventType, _ := event["type"].(string)
					agentID, _ := event["agent_id"].(string)
					eventID, _ := event["event_id"].(string)
					if eventType == "" {
						eventType = "unknown"
					}
					if agentID == "" {
						agentID = "unknown"
					}
					fmt.Printf("  [%s] %s: %s\n", eventID, eventType, agentID)
				}
			}

			return nil
		},
	}
	cmd.Flags().String("agent", "", "Filter by agent ID")
	cmd.Flags().Int("n", 50, "Number of events")
	return cmd
}
