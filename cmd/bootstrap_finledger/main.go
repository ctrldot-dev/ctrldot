// bootstrap_finledger registers the Financial Ledger policy set for FinLedger:*.
// Run after migrations (0002_finledger_namespaces.sql). Requires DB_URL.
//
// Usage: go run ./cmd/bootstrap_finledger [path/to/financialledger-policyset.yaml]
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/policy"
	"github.com/futurematic/kernel/internal/store"
)

func main() {
	ctx := context.Background()
	yamlPath := "financialledger-policyset.yaml"
	if len(os.Args) > 1 {
		yamlPath = os.Args[1]
	}

	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read policy YAML: %v\n", err)
		os.Exit(1)
	}

	engine := policy.NewEngine()
	policyHash, err := engine.ComputePolicyHash(string(yamlContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "compute policy hash: %v\n", err)
		os.Exit(1)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DB_URL is required")
		os.Exit(1)
	}

	s, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	tx, err := s.OpenTx(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open tx: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback()

	ps := domain.PolicySet{
		ID:          "financialledger-policyset",
		NamespaceID: "FinLedger",
		PolicyYAML:  string(yamlContent),
		PolicyHash:  policyHash,
		CreatedAt:   time.Now(),
		CreatedSeq:  1,
	}
	if err := tx.StorePolicySet(ctx, ps, 1); err != nil {
		fmt.Fprintf(os.Stderr, "store policy set: %v\n", err)
		os.Exit(1)
	}
	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "commit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Registered financialledger-policyset for FinLedger:*")
}
