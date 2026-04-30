package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/donnellyk/lazyruin/pkg/configpath"
)

type VaultEntry struct {
	VaultPath           string   `json:"vault_path"`
	LastLazyruinVersion string   `json:"last_lazyruin_version"`
	LastRuinVersion     string   `json:"last_ruin_version"`
	AppliedMigrations   []string `json:"applied_migrations"`
}

// stateFile uses json.RawMessage values so unknown fields in entries
// (added by future schema bumps) round-trip untouched.
type stateFile struct {
	Vaults map[string]json.RawMessage `json:"vaults"`
}

type Store struct {
	path  string
	state stateFile
}

func NewStore() *Store                    { return &Store{path: defaultStatePath()} }
func NewStoreWithPath(path string) *Store { return &Store{path: path} }

// Load reads state.json into memory. A missing file is not an error —
// the Store stays empty so first launches work without setup.
func (s *Store) Load() error {
	s.state = stateFile{Vaults: map[string]json.RawMessage{}}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &s.state); err != nil {
		return fmt.Errorf("parse %s: %w", s.path, err)
	}
	if s.state.Vaults == nil {
		s.state.Vaults = map[string]json.RawMessage{}
	}
	return nil
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	if s.state.Vaults == nil {
		s.state.Vaults = map[string]json.RawMessage{}
	}
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// VaultEntry returns the recorded entry for a vault path, or zero values
// (non-nil entry, empty fields) when the vault hasn't been seen before.
// The boolean reports whether the vault was already present.
func (s *Store) VaultEntry(vaultPath string) (VaultEntry, bool) {
	raw, ok := s.state.Vaults[vaultKey(vaultPath)]
	if !ok {
		return VaultEntry{}, false
	}
	var entry VaultEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		// Malformed entry — return zero values so callers proceed as if
		// first launch. The bad raw bytes survive in state.Vaults until
		// we re-record this vault.
		return VaultEntry{}, false
	}
	return entry, true
}

// RecordVersions writes the current version pair into the entry for the
// given vault, preserving any previously-applied migration IDs. Used on
// every launch where no migrations were pending so the next launch sees
// an accurate "previous" version.
func (s *Store) RecordVersions(vaultPath string, curr VersionPair) {
	entry, _ := s.VaultEntry(vaultPath)
	entry.VaultPath = vaultPath
	entry.LastLazyruinVersion = curr.Lazyruin
	entry.LastRuinVersion = curr.Ruin
	s.put(vaultPath, entry)
}

// RecordApplied marks the given migration IDs as applied for the vault
// and updates the version pair to the current values. Called only after
// a successful run.
func (s *Store) RecordApplied(vaultPath string, curr VersionPair, ids []string) {
	entry, _ := s.VaultEntry(vaultPath)
	entry.VaultPath = vaultPath
	entry.LastLazyruinVersion = curr.Lazyruin
	entry.LastRuinVersion = curr.Ruin
	for _, id := range ids {
		if !contains(entry.AppliedMigrations, id) {
			entry.AppliedMigrations = append(entry.AppliedMigrations, id)
		}
	}
	s.put(vaultPath, entry)
}

func (s *Store) put(vaultPath string, entry VaultEntry) {
	if s.state.Vaults == nil {
		s.state.Vaults = map[string]json.RawMessage{}
	}
	data, _ := json.Marshal(entry)
	s.state.Vaults[vaultKey(vaultPath)] = data
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func vaultKey(vaultPath string) string { return configpath.VaultKey(vaultPath) }

func defaultStatePath() string {
	return filepath.Join(configpath.Dir(), "state.json")
}
