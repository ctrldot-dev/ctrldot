package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the dot CLI configuration
type Config struct {
	Server      string   `json:"server"`
	ActorID     string   `json:"actor_id"`
	NamespaceID string   `json:"namespace_id"`
	Capabilities []string `json:"capabilities"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server:      "http://localhost:8080",
		ActorID:     "user:anonymous",
		NamespaceID: "",
		Capabilities: []string{"read", "write"},
	}

	// Load from file
	configPath := getConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if server := os.Getenv("DOT_SERVER"); server != "" {
		cfg.Server = server
	}
	if actor := os.Getenv("DOT_ACTOR"); actor != "" {
		cfg.ActorID = actor
	}
	if namespace := os.Getenv("DOT_NAMESPACE"); namespace != "" {
		cfg.NamespaceID = namespace
	}
	if caps := os.Getenv("DOT_CAPABILITIES"); caps != "" {
		cfg.Capabilities = strings.Split(caps, ",")
		for i := range cfg.Capabilities {
			cfg.Capabilities[i] = strings.TrimSpace(cfg.Capabilities[i])
		}
	}

	return cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get gets a config value by key
func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "server":
		return cfg.Server, nil
	case "actor_id":
		return cfg.ActorID, nil
	case "namespace_id":
		return cfg.NamespaceID, nil
	case "capabilities":
		return strings.Join(cfg.Capabilities, ","), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Set sets a config value by key
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "server":
		cfg.Server = value
	case "actor_id":
		cfg.ActorID = value
	case "namespace_id":
		cfg.NamespaceID = value
	case "capabilities":
		cfg.Capabilities = strings.Split(value, ",")
		for i := range cfg.Capabilities {
			cfg.Capabilities[i] = strings.TrimSpace(cfg.Capabilities[i])
		}
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".dot/config.json"
	}
	return filepath.Join(home, ".dot", "config.json")
}
