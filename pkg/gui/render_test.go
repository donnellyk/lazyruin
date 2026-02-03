package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTruncate_Short(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Errorf("truncate(%q, 10) = %q, want %q", "hello", result, "hello")
	}
}

func TestTruncate_Exact(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("truncate(%q, 5) = %q, want %q", "hello", result, "hello")
	}
}

func TestTruncate_Long(t *testing.T) {
	result := truncate("hello world", 8)
	if result != "hello..." {
		t.Errorf("truncate(%q, 8) = %q, want %q", "hello world", result, "hello...")
	}
}

func TestTruncate_Empty(t *testing.T) {
	result := truncate("", 5)
	if result != "" {
		t.Errorf("truncate(%q, 5) = %q, want %q", "", result, "")
	}
}

func TestTruncate_ExactlyThreeOver(t *testing.T) {
	// "abcdefgh" is 8 chars, maxLen 8 should not truncate
	result := truncate("abcdefgh", 8)
	if result != "abcdefgh" {
		t.Errorf("truncate(%q, 8) = %q, want %q", "abcdefgh", result, "abcdefgh")
	}

	// "abcdefghi" is 9 chars, maxLen 8 should truncate to "abcde..."
	result = truncate("abcdefghi", 8)
	if result != "abcde..." {
		t.Errorf("truncate(%q, 8) = %q, want %q", "abcdefghi", result, "abcde...")
	}
}

func TestTruncate_MinLength(t *testing.T) {
	// Edge case: maxLen of 3 means we can only fit "..."
	result := truncate("hello", 3)
	if result != "..." {
		t.Errorf("truncate(%q, 3) = %q, want %q", "hello", result, "...")
	}
}

func TestTruncate_TableDriven(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"", 5, ""},
		{"ab", 5, "ab"},
		{"abcdefghij", 10, "abcdefghij"},
		{"abcdefghijk", 10, "abcdefg..."},
		{"test string here", 12, "test stri..."},
		{"a", 1, "a"},
		{"ab", 4, "ab"},
	}

	for _, tc := range tests {
		result := truncate(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q",
				tc.input, tc.maxLen, result, tc.expected)
		}
	}
}

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
