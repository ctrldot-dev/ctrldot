package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/futurematic/kernel/internal/config"
	"github.com/spf13/cobra"
)

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Daemon control commands",
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start Ctrl Dot daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := os.Getenv("CTRLDOT_CONFIG")
			if configPath == "" {
				configPath = "~/.ctrldot/config.yaml"
			}
			if len(configPath) >= 2 && configPath[:2] == "~/" {
				home, _ := os.UserHomeDir()
				if home != "" {
					configPath = filepath.Join(home, configPath[2:])
				}
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			kind := cfg.RuntimeStore.Kind
			if kind == "" {
				kind = "sqlite"
			}
			sqlitePath := cfg.RuntimeStore.SQLitePath
			if sqlitePath == "" {
				home, _ := os.UserHomeDir()
				if home != "" {
					sqlitePath = filepath.Join(home, ".ctrldot", "ctrldot.sqlite")
				}
			}
			if len(sqlitePath) >= 2 && sqlitePath[:2] == "~/" {
				home, _ := os.UserHomeDir()
				if home != "" {
					sqlitePath = filepath.Join(home, sqlitePath[2:])
				}
			}

			if kind == "sqlite" {
				dir := filepath.Dir(sqlitePath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("ensure runtime DB dir: %w", err)
				}
			}

			fmt.Printf("Runtime store: %s", kind)
			if kind == "sqlite" {
				fmt.Printf(" (%s)", sqlitePath)
			}
			fmt.Println()
			sinkKind := cfg.LedgerSink.Kind
			if sinkKind == "" {
				sinkKind = "none"
			}
			fmt.Printf("Ledger sink: %s\n", sinkKind)
			fmt.Printf("Server: http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)

			exe, err := os.Executable()
			if err != nil {
				return err
			}
			exeDir := filepath.Dir(exe)
			daemonPath := filepath.Join(exeDir, "ctrldotd")

			daemonCmd := exec.Command(daemonPath)
			daemonCmd.Stdout = os.Stdout
			daemonCmd.Stderr = os.Stderr

			if err := daemonCmd.Start(); err != nil {
				return fmt.Errorf("failed to start daemon: %w", err)
			}

			fmt.Printf("Ctrl Dot daemon started (PID: %d)\n", daemonCmd.Process.Pid)
			return nil
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop Ctrl Dot daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Find daemon process and send SIGTERM
			fmt.Printf("TODO: Implement daemon stop\n")
			return nil
		},
	}

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Show daemon logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Show logs (could tail log file or use systemd/journalctl)
			fmt.Printf("TODO: Implement daemon logs\n")
			return nil
		},
	}

	cmd.AddCommand(startCmd)
	cmd.AddCommand(stopCmd)
	cmd.AddCommand(logsCmd)
	return cmd
}
