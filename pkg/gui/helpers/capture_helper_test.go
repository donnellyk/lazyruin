package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "standard frontmatter with blank line after",
			in:   "---\nuuid: abc\n---\n\nbody line",
			want: "---\nuuid: abc\n---\n\n",
		},
		{
			name: "frontmatter without trailing blank",
			in:   "---\nuuid: abc\n---\nbody line",
			want: "---\nuuid: abc\n---\n",
		},
		{
			name: "no frontmatter",
			in:   "hello world\n",
			want: "",
		},
		{
			name: "empty file",
			in:   "",
			want: "",
		},
		{
			name: "malformed frontmatter (no closing ---)",
			in:   "---\nuuid: abc\nbody",
			want: "",
		},
		{
			name: "CRLF frontmatter",
			in:   "---\r\nuuid: abc\r\n---\r\n\r\nbody",
			want: "---\r\nuuid: abc\r\n---\r\n\r\n",
		},
		{
			name: "mixed LF open, CRLF close",
			in:   "---\nuuid: abc\n---\r\nbody",
			want: "---\nuuid: abc\n---\r\n",
		},
		{
			name: "frontmatter-only file with trailing newline after closing ---",
			in:   "---\nuuid: abc\n---\n",
			want: "---\nuuid: abc\n---\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(extractFrontmatter([]byte(tt.in)))
			if got != tt.want {
				t.Errorf("extractFrontmatter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadNoteBodyContent(t *testing.T) {
	tmp := t.TempDir()
	write := func(t *testing.T, name, content string) string {
		p := filepath.Join(tmp, name)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("strips frontmatter and leading blank", func(t *testing.T) {
		p := write(t, "a.md", "---\nuuid: abc\n---\n\n# Heading\n\nbody\n")
		got, err := readNoteBodyContent(p)
		if err != nil {
			t.Fatal(err)
		}
		want := "# Heading\n\nbody\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("no frontmatter returns full content", func(t *testing.T) {
		p := write(t, "b.md", "hello\nworld\n")
		got, err := readNoteBodyContent(p)
		if err != nil {
			t.Fatal(err)
		}
		if got != "hello\nworld\n" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("CRLF file strips frontmatter correctly", func(t *testing.T) {
		p := write(t, "crlf.md", "---\r\nuuid: abc\r\n---\r\n\r\nbody line\r\n")
		got, err := readNoteBodyContent(p)
		if err != nil {
			t.Fatal(err)
		}
		if got != "body line\r\n" {
			t.Errorf("got %q, want %q", got, "body line\r\n")
		}
	})

	t.Run("missing file errors", func(t *testing.T) {
		_, err := readNoteBodyContent(filepath.Join(tmp, "nope.md"))
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestReadNoteBodyAndMtime(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "note.md")
	if err := os.WriteFile(p, []byte("---\nuuid: abc\n---\n\nbody\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, mtime, err := readNoteBodyAndMtime(p)
	if err != nil {
		t.Fatal(err)
	}
	if mtime.IsZero() {
		t.Error("expected non-zero mtime")
	}
}
