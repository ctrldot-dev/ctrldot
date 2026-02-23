package commands

import (
	"encoding/json"
	"fmt"
	"bytes"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func autobundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autobundle",
		Short: "Auto-bundle: create signed bundles on DENY/STOP/shutdown",
	}
	cmd.AddCommand(autobundleStatusCmd())
	cmd.AddCommand(autobundleTestCmd())
	return cmd
}

func autobundleStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show autobundle config (enabled, output_dir, triggers)",
		RunE:  runAutobundleStatus,
	}
}

func runAutobundleStatus(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server")
	resp, err := http.Get(serverURL + "/v1/autobundle")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}
	outputJSON, _ := cmd.Flags().GetBool("json")
	if outputJSON {
		json.NewEncoder(os.Stdout).Encode(status)
	} else {
		enabled, _ := status["enabled"].(bool)
		if enabled {
			fmt.Println("Auto-bundles: ON")
		} else {
			fmt.Println("Auto-bundles: OFF")
		}
		if dir, ok := status["output_dir"].(string); ok && dir != "" {
			fmt.Printf("  Output dir: %s\n", dir)
		}
		if debounce, ok := status["debounce_seconds"].(float64); ok && debounce > 0 {
			fmt.Printf("  Debounce: %.0fs\n", debounce)
		}
	}
	return nil
}

func autobundleTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Force creation of one bundle (trigger=manual_test)",
		RunE:  runAutobundleTest,
	}
}

func runAutobundleTest(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server")
	req, _ := http.NewRequest(http.MethodPost, serverURL+"/v1/autobundle/test", bytes.NewReader(nil))
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
	var out map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&out)
	path := out["path"]
	outputJSON, _ := cmd.Flags().GetBool("json")
	if outputJSON {
		json.NewEncoder(os.Stdout).Encode(out)
	} else if path != "" {
		fmt.Printf("Bundle written: %s\n", path)
	} else {
		fmt.Println("Autobundle disabled; no bundle written.")
	}
	return nil
}
