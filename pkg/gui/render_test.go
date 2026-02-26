package gui

import (
	"os"
	"path/filepath"
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"
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

// TestBuildCardContent_ComposeSourceIdentity verifies that in compose mode,
// each SourceLine carries the child note's UUID/Path/LineNum — not the
// composed parent's. This is the contract that resolveTarget() depends on
// for line-level edits (todo toggle, tag add, etc.) to target the correct
// child note.
func TestBuildCardContent_ComposeSourceIdentity(t *testing.T) {
	dir := t.TempDir()

	// Create two child note files with frontmatter
	childAPath := filepath.Join(dir, "child-a.md")
	os.WriteFile(childAPath, []byte(
		"---\nuuid: child-a-uuid\ntags: [alpha]\n---\n"+
			"# Child A Title\n"+
			"\n"+
			"#alpha\n"+
			"\n"+
			"Content line A1\n"+
			"- Task A #todo\n"+
			"Content line A3\n",
	), 0644)

	childBPath := filepath.Join(dir, "child-b.md")
	os.WriteFile(childBPath, []byte(
		"---\nuuid: child-b-uuid\ntags: [beta]\n---\n"+
			"# Child B Title\n"+
			"\n"+
			"#beta\n"+
			"\n"+
			"Content line B1\n"+
			"- Task B #todo\n",
	), 0644)

	// Composed content as produced by `ruin compose --strip-title --strip-global-tags --normalize-headers`
	// Child A: title stripped, global tag (#alpha) stripped
	// Child B: title re-inserted as H2 by normalize-headers, global tag stripped
	composedContent := "" +
		"Content line A1\n" +
		"- Task A #todo\n" +
		"Content line A3\n" +
		"\n" +
		"## Child B Title\n" +
		"\n" +
		"Content line B1\n" +
		"- Task B #todo"

	sourceMap := []models.SourceMapEntry{
		{UUID: "child-a-uuid", Path: childAPath, Title: "Child A Title", StartLine: 1, EndLine: 3},
		{UUID: "child-b-uuid", Path: childBPath, Title: "Child B Title", StartLine: 5, EndLine: 8},
	}

	navHistory := context.NewSharedNavHistory()
	gui := &Gui{
		state:      NewGuiState(),
		contextMgr: NewContextMgr(),
		contexts: &context.ContextTree{
			Compose:          context.NewComposeContext(navHistory),
			CardList:         context.NewCardListContext(navHistory),
			ActivePreviewKey: "compose",
		},
	}
	gui.contexts.Compose.Note = models.Note{
		UUID:    "parent-uuid",
		Path:    filepath.Join(dir, "parent.md"),
		Content: composedContent,
	}
	gui.contexts.Compose.SourceMap = sourceMap
	// Disable markdown rendering to avoid needing chroma initialization
	gui.contexts.Compose.RenderMarkdown = false

	lines := gui.BuildCardContent(gui.contexts.Compose.Note, 80)

	// Find SourceLines for specific content and verify child identity.
	// LineNum values are 1-indexed content line numbers in the child's raw
	// file (after frontmatter), matching the ruin CLI --line convention.
	type expect struct {
		text    string
		uuid    string
		lineNum int
		path    string
	}
	expectations := []expect{
		// child-a raw content (after frontmatter):
		//   L1: # Child A Title
		//   L2: (blank)
		//   L3: #alpha
		//   L4: (blank)
		//   L5: Content line A1
		//   L6: - Task A #todo
		//   L7: Content line A3
		{"Content line A1", "child-a-uuid", 5, childAPath},
		{"- Task A #todo", "child-a-uuid", 6, childAPath},
		{"Content line A3", "child-a-uuid", 7, childAPath},
		// child-b raw content (after frontmatter):
		//   L1: # Child B Title   <-- composed has ## Child B Title (normalized)
		//   L2: (blank)
		//   L3: #beta
		//   L4: (blank)
		//   L5: Content line B1
		//   L6: - Task B #todo
		{"Content line B1", "child-b-uuid", 5, childBPath},
		{"- Task B #todo", "child-b-uuid", 6, childBPath},
	}

	for _, exp := range expectations {
		found := false
		for _, sl := range lines {
			trimmed := stripAnsi(sl.Text)
			if len(trimmed) > 0 {
				trimmed = trimmed[1:] // strip leading space added by BuildCardContent
			}
			if trimmed == exp.text {
				found = true
				if sl.UUID != exp.uuid {
					t.Errorf("line %q: UUID = %q, want %q", exp.text, sl.UUID, exp.uuid)
				}
				if sl.LineNum != exp.lineNum {
					t.Errorf("line %q: LineNum = %d, want %d", exp.text, sl.LineNum, exp.lineNum)
				}
				if sl.Path != exp.path {
					t.Errorf("line %q: Path = %q, want %q", exp.text, sl.Path, exp.path)
				}
				break
			}
		}
		if !found {
			t.Errorf("line %q not found in SourceLines", exp.text)
		}
	}
}

func TestIsHeaderLine(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		// Actual markdown headers
		{"# Header", true},
		{"## Sub Header", true},
		{"### Deep Header", true},
		{" # Indented Header", true},
		// With ANSI codes (chroma output)
		{"\x1b[38;5;37m# \x1b[0m\x1b[38;5;147mHeader\x1b[0m", true},
		// Tag lines — NOT headers
		{"#tag", false},
		{"#runner, #log", false},
		{"#followup #done", false},
		{" #ruin #log", false},
		// Empty / no hash
		{"", false},
		{"No header here", false},
		{"Some text #inline-tag", false},
	}
	for _, tt := range tests {
		got := isHeaderLine(tt.line)
		if got != tt.want {
			t.Errorf("isHeaderLine(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}
