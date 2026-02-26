package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSourceLine(t *testing.T) {
	dir := t.TempDir()

	// Note with frontmatter
	withFrontmatter := filepath.Join(dir, "with_fm.md")
	os.WriteFile(withFrontmatter, []byte("---\ntitle: Test\n---\nline one\nline two\nline three\n"), 0o644)

	// Note without frontmatter
	noFrontmatter := filepath.Join(dir, "no_fm.md")
	os.WriteFile(noFrontmatter, []byte("line one\nline two\nline three\n"), 0o644)

	t.Run("first content line with frontmatter", func(t *testing.T) {
		line, _, _ := readSourceLine(withFrontmatter, 1)
		if line != "line one" {
			t.Errorf("readSourceLine(_, 1) = %q, want %q", line, "line one")
		}
	})

	t.Run("second content line with frontmatter", func(t *testing.T) {
		line, _, _ := readSourceLine(withFrontmatter, 2)
		if line != "line two" {
			t.Errorf("readSourceLine(_, 2) = %q, want %q", line, "line two")
		}
	})

	t.Run("third content line with frontmatter", func(t *testing.T) {
		line, _, _ := readSourceLine(withFrontmatter, 3)
		if line != "line three" {
			t.Errorf("readSourceLine(_, 3) = %q, want %q", line, "line three")
		}
	})

	t.Run("first line without frontmatter", func(t *testing.T) {
		line, _, _ := readSourceLine(noFrontmatter, 1)
		if line != "line one" {
			t.Errorf("readSourceLine(_, 1) = %q, want %q", line, "line one")
		}
	})

	t.Run("out of range line", func(t *testing.T) {
		line, _, _ := readSourceLine(withFrontmatter, 100)
		if line != "" {
			t.Errorf("readSourceLine(_, 100) = %q, want empty", line)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		line, _, _ := readSourceLine("/nonexistent/file.md", 1)
		if line != "" {
			t.Errorf("readSourceLine(nonexistent) = %q, want empty", line)
		}
	})
}

func TestInlineTagRegex(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"#todo #done", []string{"#todo", "#done"}},
		{"no tags here", nil},
		{"#single-tag", []string{"#single-tag"}},
		{"before #middle after", []string{"#middle"}},
		{"#a #b #c", []string{"#a", "#b", "#c"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := inlineTagRe.FindAllString(tc.input, -1)
			if len(got) != len(tc.want) {
				t.Errorf("inlineTagRe matches = %v, want %v", got, tc.want)
				return
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("match[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestInlineDateRegex(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"@2026-02-23", []string{"@2026-02-23"}},
		{"before @2026-01-01 after @2026-12-31", []string{"@2026-01-01", "@2026-12-31"}},
		{"no dates here", nil},
		{"@today", nil}, // only ISO dates match
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := inlineDateRe.FindAllString(tc.input, -1)
			if len(got) != len(tc.want) {
				t.Errorf("inlineDateRe matches = %v, want %v", got, tc.want)
				return
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("match[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
