// Package configpath centralizes lazyruin's per-user config-dir
// resolution and vault-keyed file naming. Used by features that store
// per-vault state under ~/.config/lazyruin/ (scratchpad, migrations,
// etc.) so the directory and hashing rules stay consistent.
package configpath

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the lazyruin config directory: $XDG_CONFIG_HOME/lazyruin
// when set, otherwise ~/.config/lazyruin.
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lazyruin")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazyruin")
}

// VaultKey returns a stable hex SHA-256 of the vault path. Used as the
// map key in per-vault state files.
func VaultKey(vaultPath string) string {
	hash := sha256.Sum256([]byte(vaultPath))
	return hex.EncodeToString(hash[:])
}

// VaultFileName returns a short hex filename derived from the vault
// path, suffixed with the given extension (no leading dot). Useful for
// per-vault files placed alongside one another in a single directory
// (e.g., scratchpad/<short>.json).
func VaultFileName(vaultPath, ext string) string {
	hash := sha256.Sum256([]byte(vaultPath))
	return fmt.Sprintf("%x.%s", hash[:8], ext)
}
