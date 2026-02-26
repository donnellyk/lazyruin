package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestVaultAbbreviations_VaultSpecific(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"/vault/a": {"mtg": "## Meeting Notes\n"},
			"/vault/b": {"log": "## Daily Log\n"},
		},
	}

	got := cfg.VaultAbbreviations("/vault/a")
	if got["mtg"] != "## Meeting Notes\n" {
		t.Errorf("expected vault-specific abbreviation, got %v", got)
	}
}

func TestVaultAbbreviations_LegacyFallback(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"": {"mtg": "legacy meeting"},
		},
	}

	got := cfg.VaultAbbreviations("/some/other/vault")
	if got["mtg"] != "legacy meeting" {
		t.Errorf("expected legacy fallback, got %v", got)
	}
}

func TestVaultAbbreviations_VaultSpecificOverridesLegacy(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"":        {"mtg": "legacy"},
			"/vault":  {"mtg": "vault-specific"},
		},
	}

	got := cfg.VaultAbbreviations("/vault")
	if got["mtg"] != "vault-specific" {
		t.Errorf("expected vault-specific to take precedence, got %q", got["mtg"])
	}
}

func TestVaultAbbreviations_NilAbbreviations(t *testing.T) {
	cfg := &Config{}

	got := cfg.VaultAbbreviations("/vault")
	if got == nil {
		t.Error("expected non-nil map")
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestVaultAbbreviations_EmptyMaps(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"/vault": {},
			"":       {},
		},
	}

	got := cfg.VaultAbbreviations("/vault")
	if len(got) != 0 {
		t.Errorf("expected empty map for empty vault entries, got %v", got)
	}
}

func TestSetVaultAbbreviation(t *testing.T) {
	cfg := &Config{}

	cfg.SetVaultAbbreviation("/vault", "mtg", "## Meeting\n")

	if cfg.Abbreviations == nil {
		t.Fatal("expected Abbreviations to be initialized")
	}
	if cfg.Abbreviations["/vault"]["mtg"] != "## Meeting\n" {
		t.Errorf("expected abbreviation to be set, got %v", cfg.Abbreviations)
	}
}

func TestSetVaultAbbreviation_ExistingMap(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"/vault": {"old": "existing"},
		},
	}

	cfg.SetVaultAbbreviation("/vault", "new", "added")

	if cfg.Abbreviations["/vault"]["old"] != "existing" {
		t.Error("existing abbreviation was lost")
	}
	if cfg.Abbreviations["/vault"]["new"] != "added" {
		t.Error("new abbreviation was not set")
	}
}

func TestDeleteVaultAbbreviation(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"/vault": {"mtg": "meeting", "log": "daily log"},
		},
	}

	cfg.DeleteVaultAbbreviation("/vault", "mtg")

	if _, ok := cfg.Abbreviations["/vault"]["mtg"]; ok {
		t.Error("expected abbreviation to be deleted")
	}
	if cfg.Abbreviations["/vault"]["log"] != "daily log" {
		t.Error("other abbreviation should remain")
	}
}

func TestDeleteVaultAbbreviation_RemovesEmptyVault(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"/vault": {"only": "one"},
		},
	}

	cfg.DeleteVaultAbbreviation("/vault", "only")

	if _, ok := cfg.Abbreviations["/vault"]; ok {
		t.Error("expected empty vault map to be cleaned up")
	}
}

func TestDeleteVaultAbbreviation_NilAbbreviations(t *testing.T) {
	cfg := &Config{}
	// Should not panic
	cfg.DeleteVaultAbbreviation("/vault", "nope")
}

func TestDeleteVaultAbbreviation_MissingVault(t *testing.T) {
	cfg := &Config{
		Abbreviations: map[string]map[string]string{
			"/other": {"mtg": "meeting"},
		},
	}
	// Should not panic
	cfg.DeleteVaultAbbreviation("/vault", "mtg")
}

func TestGetConfigPath_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := getConfigPath()
	want := "/custom/config/lazyruin/config.yml"
	if got != want {
		t.Errorf("getConfigPath() = %q, want %q", got, want)
	}
}

func TestGetConfigPath_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got := getConfigPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "lazyruin", "config.yml")
	if got != want {
		t.Errorf("getConfigPath() = %q, want %q", got, want)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.VaultPath != "" {
		t.Errorf("expected empty VaultPath, got %q", cfg.VaultPath)
	}
}

func TestLoad_NewFormat(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "lazyruin")
	os.MkdirAll(cfgDir, 0o755)

	data := []byte(`vault_path: /my/vault
editor: vim
chroma_theme: monokai
abbreviations:
  /my/vault:
    mtg: "## Meeting\n"
`)
	os.WriteFile(filepath.Join(cfgDir, "config.yml"), data, 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.VaultPath != "/my/vault" {
		t.Errorf("VaultPath = %q, want /my/vault", cfg.VaultPath)
	}
	if cfg.Editor != "vim" {
		t.Errorf("Editor = %q, want vim", cfg.Editor)
	}
	if cfg.Abbreviations["/my/vault"]["mtg"] != "## Meeting\n" {
		t.Errorf("expected abbreviation, got %v", cfg.Abbreviations)
	}
}

func TestLoad_OldFlatFormat_Migrates(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "lazyruin")
	os.MkdirAll(cfgDir, 0o755)

	data := []byte(`vault_path: /my/vault
abbreviations:
  mtg: "## Meeting"
  log: "## Daily Log"
`)
	cfgPath := filepath.Join(cfgDir, "config.yml")
	os.WriteFile(cfgPath, data, 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.VaultPath != "/my/vault" {
		t.Errorf("VaultPath = %q, want /my/vault", cfg.VaultPath)
	}
	// Old flat format should be migrated to nested under "" key
	legacy := cfg.Abbreviations[""]
	if legacy == nil {
		t.Fatal("expected legacy abbreviations under empty key")
	}
	if legacy["mtg"] != "## Meeting" {
		t.Errorf("expected mtg abbreviation, got %v", legacy)
	}

	// Verify the file was re-saved in new format
	reloaded, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("could not re-read config: %v", err)
	}
	var newCfg Config
	if err := yaml.Unmarshal(reloaded, &newCfg); err != nil {
		t.Fatalf("could not parse re-saved config: %v", err)
	}
	if newCfg.Abbreviations[""]["mtg"] != "## Meeting" {
		t.Errorf("re-saved config lost abbreviations: %v", newCfg.Abbreviations)
	}
}
