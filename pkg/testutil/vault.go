package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// noteCounter ensures unique note titles across tests
var noteCounter atomic.Int64

// TestVault manages a temporary vault for testing.
type TestVault struct {
	Path            string
	t               *testing.T
	origVaultEnv    string
	origVaultEnvSet bool
}

// NewTestVault creates and initializes a temporary vault.
// The vault is automatically cleaned up when the test completes.
// Also saves and restores the original LAZYRUIN_VAULT environment variable.
func NewTestVault(t *testing.T) *TestVault {
	t.Helper()

	// Check if ruin CLI is available
	if _, err := exec.LookPath("ruin"); err != nil {
		t.Skip("ruin CLI not found in PATH")
	}

	// Save original vault environment
	origVaultEnv, origVaultEnvSet := os.LookupEnv("LAZYRUIN_VAULT")

	// Create temp directory
	dir, err := os.MkdirTemp("", "lazyruin-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	vaultPath := filepath.Join(dir, "vault")

	// Initialize vault
	cmd := exec.Command("ruin", "init", vaultPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to init vault: %v\n%s", err, output)
	}

	tv := &TestVault{
		Path:            vaultPath,
		t:               t,
		origVaultEnv:    origVaultEnv,
		origVaultEnvSet: origVaultEnvSet,
	}

	// Register cleanup
	t.Cleanup(func() {
		tv.cleanup()
	})

	return tv
}

// cleanup restores the original vault environment and removes temp files.
func (tv *TestVault) cleanup() {
	// Restore original LAZYRUIN_VAULT environment variable
	if tv.origVaultEnvSet {
		os.Setenv("LAZYRUIN_VAULT", tv.origVaultEnv)
	} else {
		os.Unsetenv("LAZYRUIN_VAULT")
	}

	// Remove temp directory
	if tv.Path != "" {
		dir := filepath.Dir(tv.Path)
		os.RemoveAll(dir)
	}
}

// SetAsDefault sets this test vault as the default via environment variable.
// The original value is restored on cleanup.
func (tv *TestVault) SetAsDefault() {
	tv.t.Helper()
	os.Setenv("LAZYRUIN_VAULT", tv.Path)
}

// CreateNote creates a test note in the vault.
// Uses unique titles to avoid filename collisions.
func (tv *TestVault) CreateNote(content string, tags ...string) {
	tv.t.Helper()

	// Generate unique title to avoid timestamp collisions
	id := noteCounter.Add(1)
	title := fmt.Sprintf("test-note-%d", id)

	// Build content with tags
	fullContent := content
	for _, tag := range tags {
		fullContent += " #" + tag
	}

	cmd := exec.Command("ruin", "log", "-t", title, fullContent, "--vault", tv.Path)
	if output, err := cmd.CombinedOutput(); err != nil {
		tv.t.Fatalf("failed to create note: %v\n%s", err, output)
	}
}

// CreateNoteWithTitle creates a test note with a specific title.
func (tv *TestVault) CreateNoteWithTitle(title, content string, tags ...string) {
	tv.t.Helper()

	fullContent := content
	for _, tag := range tags {
		fullContent += " #" + tag
	}

	cmd := exec.Command("ruin", "log", "-t", title, fullContent, "--vault", tv.Path)
	if output, err := cmd.CombinedOutput(); err != nil {
		tv.t.Fatalf("failed to create note with title: %v\n%s", err, output)
	}
}

// SaveQuery creates a saved query in the vault.
func (tv *TestVault) SaveQuery(name, query string) {
	tv.t.Helper()

	cmd := exec.Command("ruin", "query", "save", name, query, "-f", "--vault", tv.Path)
	if output, err := cmd.CombinedOutput(); err != nil {
		tv.t.Fatalf("failed to save query: %v\n%s", err, output)
	}
}
