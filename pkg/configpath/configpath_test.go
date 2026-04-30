package configpath

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDir_UsesXDGWhenSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-cfg")
	if got := Dir(); got != "/tmp/xdg-cfg/lazyruin" {
		t.Errorf("Dir() = %q, want /tmp/xdg-cfg/lazyruin", got)
	}
}

func TestDir_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/Users/alice")
	if got := Dir(); !strings.HasSuffix(got, filepath.Join(".config", "lazyruin")) {
		t.Errorf("Dir() = %q, want path ending in .config/lazyruin", got)
	}
}

func TestVaultKey_StableAndDistinct(t *testing.T) {
	a := VaultKey("/x")
	if a != VaultKey("/x") {
		t.Error("vault key not stable across calls")
	}
	if a == VaultKey("/y") {
		t.Error("distinct paths produced the same key")
	}
}

func TestVaultFileName(t *testing.T) {
	got := VaultFileName("/x", "json")
	if !strings.HasSuffix(got, ".json") {
		t.Errorf("expected .json suffix, got %q", got)
	}
	if VaultFileName("/x", "json") == VaultFileName("/y", "json") {
		t.Error("distinct paths produced the same filename")
	}
}
