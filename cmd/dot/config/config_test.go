package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Set a temporary home directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server != "http://localhost:8080" {
		t.Errorf("Expected default server http://localhost:8080, got %s", cfg.Server)
	}

	if cfg.ActorID != "user:anonymous" {
		t.Errorf("Expected default actor user:anonymous, got %s", cfg.ActorID)
	}
}

func TestLoadFromFile(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".dot")
	os.MkdirAll(configDir, 0755)

	// Write config file
	configPath := filepath.Join(configDir, "config.json")
	configData := `{
  "server": "http://test:9090",
  "actor_id": "user:test",
  "namespace_id": "ProductTree:/Test",
  "capabilities": ["read", "write"]
}`
	os.WriteFile(configPath, []byte(configData), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server != "http://test:9090" {
		t.Errorf("Expected server http://test:9090, got %s", cfg.Server)
	}

	if cfg.ActorID != "user:test" {
		t.Errorf("Expected actor user:test, got %s", cfg.ActorID)
	}

	if cfg.NamespaceID != "ProductTree:/Test" {
		t.Errorf("Expected namespace ProductTree:/Test, got %s", cfg.NamespaceID)
	}

	if len(cfg.Capabilities) != 2 || cfg.Capabilities[0] != "read" || cfg.Capabilities[1] != "write" {
		t.Errorf("Expected capabilities [read write], got %v", cfg.Capabilities)
	}
}

func TestEnvOverrides(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalServer := os.Getenv("DOT_SERVER")
	originalActor := os.Getenv("DOT_ACTOR")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("DOT_SERVER", originalServer)
		os.Setenv("DOT_ACTOR", originalActor)
	}()

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("DOT_SERVER", "http://env:8080")
	os.Setenv("DOT_ACTOR", "user:env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server != "http://env:8080" {
		t.Errorf("Expected env server http://env:8080, got %s", cfg.Server)
	}

	if cfg.ActorID != "user:env" {
		t.Errorf("Expected env actor user:env, got %s", cfg.ActorID)
	}
}

func TestGetSet(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Test Set
	if err := Set("server", "http://test:8080"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Test Get
	value, err := Get("server")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if value != "http://test:8080" {
		t.Errorf("Expected http://test:8080, got %s", value)
	}

	// Test invalid key
	_, err = Get("invalid_key")
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}
