package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsOnlyNewConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	newPath := filepath.Join(home, defaultConfigPath)
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte(`{"app_id":"new","app_secret":"new-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AppID != "new" {
		t.Fatalf("expected new path config, got %q", cfg.AppID)
	}
}

func TestLoadDoesNotFallbackToLegacyPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	oldPath := filepath.Join(home, ".config/feishu-docs/config.json")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, []byte(`{"app_id":"legacy","app_secret":"legacy-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("expected Load to fail when only legacy config exists")
	}
}

func TestEnsureConfigFileCreatesNewPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	newPath := filepath.Join(home, defaultConfigPath)

	gotPath, err := EnsureConfigFile()
	if err != nil {
		t.Fatalf("EnsureConfigFile returned error: %v", err)
	}
	if gotPath != newPath {
		t.Fatalf("expected path %q, got %q", newPath, gotPath)
	}

	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}
