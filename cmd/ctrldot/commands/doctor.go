package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/runtime/sqlite"
	"github.com/futurematic/kernel/internal/store"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run checks and print actionable fixes",
		Long:  "Checks: runtime DB (open + migrate), ledger sink (kernel_http reachable, bundle keys).",
		RunE:  runDoctor,
	}
	return cmd
}

func runDoctor(cmd *cobra.Command, args []string) error {
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
		fmt.Printf("✗ Load config: %v\n", err)
		return nil
	}
	fmt.Printf("✓ Config loaded from %s\n", configPath)

	// Runtime store
	kind := cfg.RuntimeStore.Kind
	if kind == "" {
		kind = "sqlite"
	}
	switch kind {
	case "sqlite":
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
		st, err := sqlite.Open(context.Background(), sqlitePath)
		if err != nil {
			fmt.Printf("✗ Runtime store (SQLite): %v\n", err)
			fmt.Printf("  Fix: ensure directory exists and path is writable: %s\n", filepath.Dir(sqlitePath))
		} else {
			_ = st.Close()
			fmt.Printf("✓ Runtime store (SQLite): %s\n", sqlitePath)
		}
	case "postgres":
		dbURL := cfg.RuntimeStore.DBURL
		if dbURL == "" {
			dbURL = cfg.Ledger.DBURL
		}
		if dbURL == "" {
			fmt.Printf("✗ Runtime store (Postgres): no db_url in config or DB_URL in env\n")
		} else {
			st, err := store.NewPostgresStore(dbURL)
			if err != nil {
				fmt.Printf("✗ Runtime store (Postgres): %v\n", err)
				fmt.Printf("  Fix: start Postgres, set DB_URL or runtime_store.db_url\n")
			} else {
				_ = st.Close()
				fmt.Printf("✓ Runtime store (Postgres): connected\n")
			}
		}
	default:
		fmt.Printf("✗ Unknown runtime_store.kind: %q (use sqlite or postgres)\n", kind)
	}

	// Ledger sink
	sinkKind := cfg.LedgerSink.Kind
	if sinkKind == "" {
		sinkKind = "none"
	}
	fmt.Printf("  Ledger sink: %s\n", sinkKind)
	if sinkKind == "kernel_http" {
		// Would need ledger_kernel_http.base_url from config; spec says CTRLDOT_KERNEL_URL
		kernelURL := os.Getenv("CTRLDOT_KERNEL_URL")
		if kernelURL == "" {
			kernelURL = "http://127.0.0.1:8080"
		}
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(kernelURL + "/health")
		if err != nil {
			fmt.Printf("✗ Kernel (kernel_http): %v\n", err)
			fmt.Printf("  Fix: start Kernel or set CTRLDOT_KERNEL_URL\n")
		} else {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				fmt.Printf("✓ Kernel (kernel_http): %s reachable\n", kernelURL)
			} else {
				fmt.Printf("✗ Kernel (kernel_http): %s returned %d\n", kernelURL, resp.StatusCode)
			}
		}
	}
	if sinkKind == "bundle" {
		keysDir := filepath.Join(os.Getenv("HOME"), ".ctrldot", "keys")
		if os.Getenv("HOME") == "" {
			keysDir = ".ctrldot/keys"
		}
		if _, err := os.Stat(keysDir); os.IsNotExist(err) {
			fmt.Printf("  Bundle: keys dir missing (%s); will be generated on first run\n", keysDir)
		} else {
			fmt.Printf("✓ Bundle: keys dir exists: %s\n", keysDir)
		}
	}

	fmt.Printf("  Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	return nil
}
