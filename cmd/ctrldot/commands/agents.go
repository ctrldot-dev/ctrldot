package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func agentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "List all agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			outputJSON, _ := cmd.Flags().GetBool("json")

			resp, err := http.Get(serverURL + "/v1/agents")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var agents []interface{}
			if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
				return err
			}

			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(agents)
			} else {
				fmt.Printf("Agents:\n")
				for _, agent := range agents {
					agentMap := agent.(map[string]interface{})
					fmt.Printf("  %s: %s\n", agentMap["agent_id"], agentMap["display_name"])
				}
			}

			return nil
		},
	}
	return cmd
}

func agentShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <agent_id>",
		Short: "Show agent details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			outputJSON, _ := cmd.Flags().GetBool("json")
			agentID := args[0]

			resp, err := http.Get(serverURL + "/v1/agents/" + agentID)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var agent map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
				return err
			}

			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(agent)
			} else {
				fmt.Printf("Agent: %s\n", agentID)
				fmt.Printf("  Display Name: %v\n", agent["display_name"])
				fmt.Printf("  Default Mode: %v\n", agent["default_mode"])
			}

			return nil
		},
	}
	return cmd
}

func haltCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "halt <agent_id>",
		Short: "Halt an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			agentID := args[0]
			reason, _ := cmd.Flags().GetString("reason")
			if reason == "" {
				reason = "Halted via CLI"
			}

			reqBody := map[string]string{"reason": reason}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(serverURL+"/v1/agents/"+agentID+"/halt", "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			fmt.Printf("Agent %s halted\n", agentID)
			return nil
		},
	}
	cmd.Flags().String("reason", "", "Reason for halting")
	return cmd
}

func resumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <agent_id>",
		Short: "Resume a halted agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			agentID := args[0]

			resp, err := http.Post(serverURL+"/v1/agents/"+agentID+"/resume", "application/json", nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			fmt.Printf("Agent %s resumed\n", agentID)
			return nil
		},
	}
	return cmd
}
