package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/futurematic/kernel/cmd/dot/config"
	"github.com/futurematic/kernel/cmd/dot/output"
	"github.com/spf13/cobra"
)

// CommandContext holds shared context for commands
type CommandContext struct {
	Client    *client.Client
	Formatter output.Formatter
	Config    *config.Config
}

// getContext creates a command context from flags and config
func getContext(cmd *cobra.Command) (*CommandContext, error) {
	// Load base config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Apply flag overrides
	if server, _ := cmd.Flags().GetString("server"); server != "" {
		cfg.Server = server
	}
	if actor, _ := cmd.Flags().GetString("actor"); actor != "" {
		cfg.ActorID = actor
	}
	if namespace, _ := cmd.Flags().GetString("ns"); namespace != "" {
		cfg.NamespaceID = namespace
	}
	if caps, _ := cmd.Flags().GetString("cap"); caps != "" {
		cfg.Capabilities = strings.Split(caps, ",")
		for i := range cfg.Capabilities {
			cfg.Capabilities[i] = strings.TrimSpace(cfg.Capabilities[i])
		}
	}

	// Create client
	apiClient := client.NewClient(cfg.Server)

	// Create formatter
	format := "text"
	if json, _ := cmd.Flags().GetBool("json"); json {
		format = "json"
	}
	formatter := output.NewFormatter(format, os.Stdout)

	return &CommandContext{
		Client:    apiClient,
		Formatter: formatter,
		Config:    cfg,
	}, nil
}

// handleError handles API errors and exits with appropriate code
func handleError(err error) {
	if apiErr, ok := err.(*client.APIError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Message)
		os.Exit(apiErr.ExitCode())
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(4)
}

// getNamespaceID returns namespace ID from config or nil if empty
func getNamespaceID(cfg *config.Config) *string {
	if cfg.NamespaceID == "" {
		return nil
	}
	return &cfg.NamespaceID
}
