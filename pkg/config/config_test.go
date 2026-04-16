package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig_ViewOptionsRoundTrip verifies that ViewOptions.HideDone
// survives a Save() + Load() cycle via the on-disk YAML file.
func TestConfig_ViewOptionsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &Config{
		VaultPath: "/tmp/some-vault",
	}
	cfg.ViewOptions.HideDone = true
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reloaded.ViewOptions.HideDone {
		t.Error("reloaded ViewOptions.HideDone = false, want true")
	}
	if reloaded.VaultPath != "/tmp/some-vault" {
		t.Errorf("reloaded VaultPath = %q, want /tmp/some-vault", reloaded.VaultPath)
	}
}

// TestConfig_ViewOptionsMissing_DefaultsToFalse verifies that a YAML config
// written before ViewOptions existed loads with HideDone=false (zero value).
func TestConfig_ViewOptionsMissing_DefaultsToFalse(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	configPath := filepath.Join(tmp, "lazyruin", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Config file with no view_options key at all.
	legacy := "vault_path: /tmp/v\neditor: vim\n"
	if err := os.WriteFile(configPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ViewOptions.HideDone {
		t.Error("expected HideDone to default to false for legacy config without view_options")
	}
}
