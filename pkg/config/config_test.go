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

// TestConfig_NotesPane_RoundTrip verifies that a notes_pane block with custom
// sections survives Save/Load via YAML.
func TestConfig_NotesPane_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &Config{VaultPath: "/tmp/some-vault"}
	cfg.NotesPane.SectionsMode = true
	cfg.NotesPane.CustomSections = []NotesPaneSection{
		{
			Title: "Reading Queue",
			Items: []NotesPaneSectionItem{
				{Title: "Articles to read", Embed: "![[search: #article !#done | limit=20]]"},
				{Title: "Long-form", Embed: "![[search: #article #longform !#done]]"},
			},
		},
		{
			// untitled — should round-trip with empty title
			Items: []NotesPaneSectionItem{
				{Title: "Followups this week", Embed: "![[search: #followup between:today-7,today]]"},
			},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reloaded.NotesPane.SectionsMode {
		t.Error("reloaded SectionsMode = false, want true")
	}
	got := reloaded.NotesPane.CustomSections
	if len(got) != 2 {
		t.Fatalf("len(CustomSections) = %d, want 2", len(got))
	}
	if got[0].Title != "Reading Queue" || len(got[0].Items) != 2 {
		t.Errorf("section[0] = %+v, want Title=Reading Queue with 2 items", got[0])
	}
	if got[0].Items[0].Embed != "![[search: #article !#done | limit=20]]" {
		t.Errorf("section[0].items[0].Embed = %q", got[0].Items[0].Embed)
	}
	if got[1].Title != "" || len(got[1].Items) != 1 {
		t.Errorf("section[1] = %+v, want untitled with 1 item", got[1])
	}
}

// TestConfig_NotesPane_DefaultsOff verifies that a config without any
// notes_pane block defaults to SectionsMode=false (legacy behavior).
func TestConfig_NotesPane_DefaultsOff(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	configPath := filepath.Join(tmp, "lazyruin", "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := "vault_path: /tmp/v\neditor: vim\n"
	if err := os.WriteFile(configPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.NotesPane.SectionsMode {
		t.Error("expected SectionsMode to default to false for legacy config")
	}
	if cfg.NotesPane.CustomSections != nil {
		t.Errorf("expected CustomSections nil, got %+v", cfg.NotesPane.CustomSections)
	}
}
