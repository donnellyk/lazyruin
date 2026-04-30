package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStore_LoadMissingFileIsEmpty(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "state.json"))
	if err := s.Load(); err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if entry, ok := s.VaultEntry("/some/vault"); ok {
		t.Errorf("expected no entry on empty store, got %+v", entry)
	}
}

func TestStore_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := NewStoreWithPath(path)

	curr := VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	s.RecordApplied("/v1", curr, []string{"mig-a"})
	s.RecordVersions("/v2", VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.0"})

	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	s2 := NewStoreWithPath(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	e1, ok := s2.VaultEntry("/v1")
	if !ok {
		t.Fatal("expected /v1 entry after reload")
	}
	if e1.LastLazyruinVersion != "0.2.0" || e1.LastRuinVersion != "0.3.1" {
		t.Errorf("/v1 versions = %q/%q, want 0.2.0/0.3.1", e1.LastLazyruinVersion, e1.LastRuinVersion)
	}
	if len(e1.AppliedMigrations) != 1 || e1.AppliedMigrations[0] != "mig-a" {
		t.Errorf("/v1 AppliedMigrations = %v, want [mig-a]", e1.AppliedMigrations)
	}
	if e1.VaultPath != "/v1" {
		t.Errorf("/v1 VaultPath = %q, want /v1", e1.VaultPath)
	}

	e2, ok := s2.VaultEntry("/v2")
	if !ok {
		t.Fatal("expected /v2 entry after reload")
	}
	if e2.LastRuinVersion != "0.3.0" {
		t.Errorf("/v2 ruin = %q", e2.LastRuinVersion)
	}
	if len(e2.AppliedMigrations) != 0 {
		t.Errorf("/v2 AppliedMigrations should be empty, got %v", e2.AppliedMigrations)
	}
}

func TestStore_PerVaultIsolation(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "state.json"))
	s.RecordApplied("/a", VersionPair{Lazyruin: "x", Ruin: "y"}, []string{"id-a"})

	if entry, ok := s.VaultEntry("/b"); ok {
		t.Errorf("expected /b empty after recording /a, got %+v", entry)
	}
}

func TestStore_RecordAppliedAccumulates(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "state.json"))
	curr := VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"}
	s.RecordApplied("/v", curr, []string{"a"})
	s.RecordApplied("/v", curr, []string{"b", "a"}) // "a" already there; should dedupe.
	entry, _ := s.VaultEntry("/v")
	if len(entry.AppliedMigrations) != 2 {
		t.Errorf("expected 2 unique IDs, got %v", entry.AppliedMigrations)
	}
}

func TestStore_MalformedFileIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("not json {"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	s := NewStoreWithPath(path)
	if err := s.Load(); err == nil {
		t.Fatal("expected error on malformed file")
	}
}

func TestStore_ForwardCompat_PreservesUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	// Pre-populate with a vault whose entry has an unknown field.
	raw := []byte(`{
  "vaults": {
    "abc": {
      "vault_path": "/abc",
      "last_lazyruin_version": "0.1.0",
      "last_ruin_version": "0.3.0",
      "applied_migrations": ["legacy-id"],
      "future_field": "hello"
    }
  }
}`)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	s := NewStoreWithPath(path)
	if err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Touch a different vault so we re-write the file.
	s.RecordVersions("/other", VersionPair{Lazyruin: "0.2.0", Ruin: "0.3.1"})
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	roundTripped, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	// The unknown field on the abc entry should still be present.
	var parsed struct {
		Vaults map[string]map[string]any `json:"vaults"`
	}
	if err := json.Unmarshal(roundTripped, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	abc, ok := parsed.Vaults["abc"]
	if !ok {
		t.Fatal("abc vault missing after round-trip")
	}
	if abc["future_field"] != "hello" {
		t.Errorf("future_field lost in round-trip: %+v", abc)
	}
}

func TestVaultKey_Stable(t *testing.T) {
	a := vaultKey("/some/path")
	b := vaultKey("/some/path")
	if a != b {
		t.Errorf("hash not stable: %q vs %q", a, b)
	}
	if a == vaultKey("/other/path") {
		t.Error("different paths hashed to the same key")
	}
}
