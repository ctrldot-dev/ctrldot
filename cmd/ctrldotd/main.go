package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	ctrldotapi "github.com/futurematic/kernel/internal/api/ctrldot"
	"github.com/futurematic/kernel/internal/config"
	ctrldotsvc "github.com/futurematic/kernel/internal/ctrldot"
	"github.com/futurematic/kernel/internal/ledger/autobundle"
	"github.com/futurematic/kernel/internal/ledger/sink"
	"github.com/futurematic/kernel/internal/ledger/sink/bundle"
	"github.com/futurematic/kernel/internal/ledger/sink/kernel_http"
	"github.com/futurematic/kernel/internal/ledger/sink/noop"
	"github.com/futurematic/kernel/internal/limits"
	"github.com/futurematic/kernel/internal/loop"
	"github.com/futurematic/kernel/internal/resolution"
	"github.com/futurematic/kernel/internal/rules"
	"github.com/futurematic/kernel/internal/runtime"
	"github.com/futurematic/kernel/internal/runtime/sqlite"
	"github.com/futurematic/kernel/internal/store"
)

func main() {
	configPath := os.Getenv("CTRLDOT_CONFIG")
	if configPath == "" {
		configPath = "~/.ctrldot/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var runtimeStore runtime.RuntimeStore
	var closeStore func() error

	switch cfg.RuntimeStore.Kind {
	case "postgres":
		// Legacy: full Postgres store, wrapped as RuntimeStore
		dbURL := cfg.RuntimeStore.DBURL
		if dbURL == "" {
			dbURL = cfg.Ledger.DBURL
		}
		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			log.Fatalf("Failed to initialize Postgres store: %v", err)
		}
		runtimeStore = runtime.NewPostgresStore(st)
		closeStore = st.Close
	default:
		// Default: SQLite (no Postgres required); "" or "sqlite"
		sqlitePath := cfg.RuntimeStore.SQLitePath
		if sqlitePath == "" {
			home, _ := os.UserHomeDir()
			if home != "" {
				sqlitePath = home + "/.ctrldot/ctrldot.sqlite"
			} else {
				sqlitePath = ".ctrldot/ctrldot.sqlite"
			}
		}
		st, err := sqlite.Open(context.Background(), sqlitePath)
		if err != nil {
			log.Fatalf("Failed to open SQLite runtime store: %v", err)
		}
		runtimeStore = st
		closeStore = st.Close
	}
	defer func() {
		if closeStore != nil {
			_ = closeStore()
		}
	}()

	limitsEngine := limits.NewEngine(runtimeStore, cfg)
	rulesEngine := rules.NewEngine(cfg)
	loopDetector := loop.NewDetector(runtimeStore, cfg)

	// Resolution manager is stateless (HMAC tokens); store is unused in current impl
	resolutionMgr := resolution.NewManager(nil, "")

	var ledgerSink sink.LedgerSink = noop.New()
	runtimeKind := cfg.RuntimeStore.Kind
	if runtimeKind == "" {
		runtimeKind = "sqlite"
	}
	switch cfg.LedgerSink.Kind {
	case "bundle":
		if cfg.LedgerSink.Bundle.OutputDir == "" {
			if home, _ := os.UserHomeDir(); home != "" {
				cfg.LedgerSink.Bundle.OutputDir = filepath.Join(home, ".ctrldot", "bundles")
			}
		}
		b, err := bundle.NewSink(cfg, runtimeKind, "0.1.0")
		if err != nil {
			log.Fatalf("Failed to create bundle sink: %v", err)
		}
		ledgerSink = b
	case "kernel_http":
		baseURL := cfg.LedgerSink.KernelHTTP.BaseURL
		if baseURL == "" {
			baseURL = "http://127.0.0.1:8080"
		}
		ledgerSink = kernel_http.NewSink(
			baseURL,
			cfg.LedgerSink.KernelHTTP.APIKey,
			cfg.LedgerSink.KernelHTTP.TimeoutMs,
			cfg.LedgerSink.KernelHTTP.Required,
		)
	default:
		ledgerSink = noop.New()
	}
	defer func() {
		if ledgerSink != nil {
			_ = ledgerSink.Close()
		}
	}()

	autobundleMgr := autobundle.NewManager(cfg, runtimeStore, "0.1.0")

	ctrldotService := ctrldotsvc.NewService(
		runtimeStore,
		limitsEngine,
		rulesEngine,
		loopDetector,
		resolutionMgr,
		ledgerSink,
		autobundleMgr,
		cfg,
	)

	apiServer := ctrldotapi.NewServer(cfg.Server.Port, ctrldotService, autobundleMgr)

	go func() {
		if err := apiServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Ctrl Dot daemon started on %s:%d (runtime_store=%s)", cfg.Server.Host, cfg.Server.Port, cfg.RuntimeStore.Kind)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	ctx := context.Background()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}
