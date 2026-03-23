package inbox

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddAppendsItem(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "inbox.json"))
	s.Add("test item")

	if s.Len() != 1 {
		t.Fatalf("expected 1 item, got %d", s.Len())
	}
	items := s.Items()
	if items[0].Text != "test item" {
		t.Fatalf("expected text 'test item', got %q", items[0].Text)
	}
	if items[0].ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestDeleteRemovesItem(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "inbox.json"))
	s.Add("a")
	s.Add("b")

	id := s.Items()[0].ID
	s.Delete(id)

	if s.Len() != 1 {
		t.Fatalf("expected 1 item after delete, got %d", s.Len())
	}
}

func TestDeleteUnknownIDIsNoop(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "inbox.json"))
	s.Add("a")
	s.Delete("nonexistent")

	if s.Len() != 1 {
		t.Fatalf("expected 1 item, got %d", s.Len())
	}
}

func TestItemsNewestFirst(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "inbox.json"))

	// Manually inject items with distinct timestamps to test ordering.
	now := time.Now()
	s.items = []Item{
		{ID: "a", Text: "first", Created: now.Add(-2 * time.Second)},
		{ID: "b", Text: "second", Created: now.Add(-1 * time.Second)},
		{ID: "c", Text: "third", Created: now},
	}

	items := s.Items()
	if items[0].Text != "third" {
		t.Fatalf("expected newest first, got %q", items[0].Text)
	}
	if items[2].Text != "first" {
		t.Fatalf("expected oldest last, got %q", items[2].Text)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.json")
	s := NewStoreWithPath(path)
	s.Add("hello")
	s.Add("world")

	if err := s.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	s2 := NewStoreWithPath(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if s2.Len() != 2 {
		t.Fatalf("expected 2 items after load, got %d", s2.Len())
	}
	texts := map[string]bool{}
	for _, item := range s2.Items() {
		texts[item.Text] = true
	}
	if !texts["hello"] || !texts["world"] {
		t.Fatalf("expected both items, got %v", texts)
	}
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	s := NewStoreWithPath(filepath.Join(t.TempDir(), "nonexistent", "inbox.json"))
	if err := s.Load(); err != nil {
		t.Fatalf("load on missing file should not error, got: %v", err)
	}
	if s.Len() != 0 {
		t.Fatalf("expected 0 items, got %d", s.Len())
	}
}

func TestPathForVaultDeterministic(t *testing.T) {
	a := PathForVault("/tmp/vault-a")
	b := PathForVault("/tmp/vault-a")
	if a != b {
		t.Fatalf("same vault path should produce same inbox path: %q != %q", a, b)
	}
}

func TestPathForVaultDiffersPerVault(t *testing.T) {
	a := PathForVault("/tmp/vault-a")
	b := PathForVault("/tmp/vault-b")
	if a == b {
		t.Fatalf("different vault paths should produce different inbox paths: %q", a)
	}
}

func TestPathForVaultInConfigDir(t *testing.T) {
	p := PathForVault("/some/vault")
	if dir := filepath.Base(filepath.Dir(p)); dir != "inboxes" {
		t.Fatalf("expected parent dir 'inboxes', got %q", dir)
	}
	if ext := filepath.Ext(p); ext != ".json" {
		t.Fatalf("expected .json extension, got %q", ext)
	}
}

func TestNewStoreForVaultUsesVaultPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	s := NewStoreForVault("/tmp/test-vault")
	s.Add("test item")
	if err := s.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	s2 := NewStoreForVault("/tmp/test-vault")
	if err := s2.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if s2.Len() != 1 {
		t.Fatalf("expected 1 item, got %d", s2.Len())
	}

	// Different vault should be empty.
	s3 := NewStoreForVault("/tmp/other-vault")
	if err := s3.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if s3.Len() != 0 {
		t.Fatalf("expected 0 items for different vault, got %d", s3.Len())
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	path := filepath.Join(dir, "inbox.json")
	s := NewStoreWithPath(path)
	s.Add("item")

	if err := s.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist after save")
	}
}
