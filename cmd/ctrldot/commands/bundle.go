package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/ledger/sink/bundle"
	"github.com/spf13/cobra"
)

func bundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "List and verify signed bundle artefacts",
	}
	cmd.AddCommand(bundleLsCmd())
	cmd.AddCommand(bundleVerifyCmd())
	return cmd
}

func bundleLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List bundle directories",
		RunE:  runBundleLs,
	}
}

func runBundleLs(cmd *cobra.Command, args []string) error {
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
		return err
	}
	outputDir := cfg.LedgerSink.Bundle.OutputDir
	if outputDir == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			outputDir = filepath.Join(home, ".ctrldot", "bundles")
		}
	}
	names, err := bundle.ListBundles(outputDir)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("No bundles found.")
		return nil
	}
	fmt.Printf("Bundles in %s:\n", outputDir)
	for _, n := range names {
		fmt.Println(" ", n)
	}
	return nil
}

func bundleVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [path]",
		Short: "Verify signature and manifest hashes of a bundle directory",
		Args:  cobra.ExactArgs(1),
		RunE:  runBundleVerify,
	}
}

func runBundleVerify(cmd *cobra.Command, args []string) error {
	dir := args[0]
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	if err := bundle.VerifyBundle(abs); err != nil {
		fmt.Printf("✗ Verify failed: %v\n", err)
		return err
	}
	fmt.Printf("✓ Bundle verified: %s\n", abs)
	return nil
}
