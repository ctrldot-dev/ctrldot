package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func panicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "panic",
		Short: "Panic mode: toggle strict/safe behaviour",
	}
	cmd.AddCommand(panicOnCmd())
	cmd.AddCommand(panicOffCmd())
	cmd.AddCommand(panicStatusCmd())
	return cmd
}

func panicOnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "on",
		Short: "Enable panic mode",
		RunE:  runPanicOn,
	}
	cmd.Flags().Duration("ttl", 0, "Auto-disable after duration (e.g. 30m)")
	cmd.Flags().String("reason", "", "Reason for enabling (e.g. \"trying new tool\")")
	return cmd
}

func runPanicOn(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server")
	ttl, _ := cmd.Flags().GetDuration("ttl")
	reason, _ := cmd.Flags().GetString("reason")
	ttlSec := 0
	if ttl > 0 {
		ttlSec = int(ttl.Seconds())
	}
	body := map[string]interface{}{"ttl_seconds": ttlSec, "reason": reason}
	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, serverURL+"/v1/panic/on", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		msg := errBody["error"]
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("%s", msg)
	}
	var state map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return err
	}
	outputJSON, _ := cmd.Flags().GetBool("json")
	if outputJSON {
		json.NewEncoder(os.Stdout).Encode(state)
	} else {
		fmt.Println("Panic mode: ON")
		if reason != "" {
			fmt.Printf("  Reason: %s\n", reason)
		}
		if ttlSec > 0 {
			fmt.Printf("  TTL: %s\n", ttl)
		}
	}
	return nil
}

func panicOffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "off",
		Short: "Disable panic mode",
		RunE:  runPanicOff,
	}
}

func runPanicOff(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server")
	req, _ := http.NewRequest(http.MethodPost, serverURL+"/v1/panic/off", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("%s", errBody["error"])
	}
	outputJSON, _ := cmd.Flags().GetBool("json")
	if outputJSON {
		json.NewEncoder(os.Stdout).Encode(map[string]bool{"enabled": false})
	} else {
		fmt.Println("Panic mode: OFF")
	}
	return nil
}

func panicStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current panic mode state",
		RunE:  runPanicStatus,
	}
}

func runPanicStatus(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server")
	resp, err := http.Get(serverURL + "/v1/panic")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	var state struct {
		Enabled   bool      `json:"enabled"`
		EnabledAt time.Time `json:"enabled_at"`
		ExpiresAt *string   `json:"expires_at,omitempty"`
		TTLSeconds int     `json:"ttl_seconds"`
		Reason    string    `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return err
	}
	outputJSON, _ := cmd.Flags().GetBool("json")
	if outputJSON {
		json.NewEncoder(os.Stdout).Encode(state)
	} else {
		if state.Enabled {
			fmt.Println("Panic mode: ON")
			if !state.EnabledAt.IsZero() {
				fmt.Printf("  Enabled at: %s\n", state.EnabledAt.Format(time.RFC3339))
			}
			if state.Reason != "" {
				fmt.Printf("  Reason: %s\n", state.Reason)
			}
			if state.TTLSeconds > 0 {
				fmt.Printf("  TTL: %d seconds\n", state.TTLSeconds)
			}
			if state.ExpiresAt != nil && *state.ExpiresAt != "" {
				fmt.Printf("  Expires at: %s\n", *state.ExpiresAt)
			}
		} else {
			fmt.Println("Panic mode: OFF")
		}
	}
	return nil
}
