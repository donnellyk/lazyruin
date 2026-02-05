package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNoteContent(t *testing.T) {
	// Create a temp file with known content
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.md")
	content := "# Test Note\n\nThis is test content."

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gui := &Gui{}
	result, err := gui.loadNoteContent(testFile)
	if err != nil {
		t.Fatalf("loadNoteContent() error: %v", err)
	}

	if result != content {
		t.Errorf("loadNoteContent() = %q, want %q", result, content)
	}
}

func TestLoadNoteContent_NotFound(t *testing.T) {
	gui := &Gui{}
	_, err := gui.loadNoteContent("/nonexistent/path/file.md")

	if err == nil {
		t.Error("loadNoteContent() should return error for nonexistent file")
	}
}

func TestLoadNoteContent_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "empty.md")

	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gui := &Gui{}
	result, err := gui.loadNoteContent(testFile)
	if err != nil {
		t.Fatalf("loadNoteContent() error: %v", err)
	}

	if result != "" {
		t.Errorf("loadNoteContent() = %q, want empty string", result)
	}
}

func TestLoadNoteContent_WithUnicode(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "unicode.md")
	content := "# 日本語タイトル\n\nEmoji: 🎉 and special chars: äöü"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gui := &Gui{}
	result, err := gui.loadNoteContent(testFile)
	if err != nil {
		t.Fatalf("loadNoteContent() error: %v", err)
	}

	if result != content {
		t.Errorf("loadNoteContent() = %q, want %q", result, content)
	}
}
