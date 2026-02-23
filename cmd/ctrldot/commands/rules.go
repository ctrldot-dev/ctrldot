package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/futurematic/kernel/internal/config"
	"github.com/spf13/cobra"
)

func rulesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules show",
		Short: "Show current rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := os.Getenv("CTRLDOT_CONFIG")
			if configPath == "" {
				home, _ := os.UserHomeDir()
				configPath = filepath.Join(home, ".ctrldot", "config.yaml")
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			outputJSON, _ := cmd.Flags().GetBool("json")
			if outputJSON {
				json.NewEncoder(os.Stdout).Encode(cfg.Rules)
			} else {
				fmt.Printf("Rules:\n")
				fmt.Printf("  Require Resolution: %v\n", cfg.Rules.RequireResolution)
				fmt.Printf("  Filesystem Allow Roots: %v\n", cfg.Rules.Filesystem.AllowRoots)
				fmt.Printf("  Network Deny All: %v\n", cfg.Rules.Network.DenyAll)
				fmt.Printf("  Network Allow Domains: %v\n", cfg.Rules.Network.AllowDomains)
			}

			return nil
		},
	}
	return cmd
}

func rulesEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules edit",
		Short: "Edit rules (opens config file in editor)",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := os.Getenv("CTRLDOT_CONFIG")
			if configPath == "" {
				home, _ := os.UserHomeDir()
				configPath = filepath.Join(home, ".ctrldot", "config.yaml")
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			cmdExec := exec.Command(editor, configPath)
			cmdExec.Stdin = os.Stdin
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr

			return cmdExec.Run()
		},
	}
	return cmd
}
