package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// TestBuildSeparatorLine_LongTitle_TrimsWithEllipsis is a regression guard:
// when a card's title is longer than the viewport width, the separator line
// must be trimmed with "…" rather than allowed to overflow and wrap the
// card frame to the next row.
func TestBuildSeparatorLine_LongTitle_TrimsWithEllipsis(t *testing.T) {
	gui := &Gui{}
	width := 40
	longTitle := " This is a really ridiculously long note title that cannot possibly fit "

	line := gui.buildSeparatorLine(true, longTitle, "", width, false)
	if w := visibleWidth(line); w != width {
		t.Errorf("visible width = %d, want %d (line wrapped or truncated wrong)\nline: %q", w, width, stripAnsi(line))
	}
	if !strings.Contains(stripAnsi(line), "…") {
		t.Errorf("trimmed title should include '…' marker, got %q", stripAnsi(line))
	}
}

// TestBuildSeparatorLine_LongTitleWithRight checks trimming still fits when
// the right-side metadata (date, tags, parent) is also populated.
func TestBuildSeparatorLine_LongTitleWithRight(t *testing.T) {
	gui := &Gui{}
	width := 40
	longTitle := " Really really really really long title "
	right := " Apr 18 · #tag "

	line := gui.buildSeparatorLine(true, longTitle, right, width, false)
	if w := visibleWidth(line); w != width {
		t.Errorf("visible width = %d, want %d\nline: %q", w, width, stripAnsi(line))
	}
	// The right-side metadata must still appear intact.
	if !strings.Contains(stripAnsi(line), strings.TrimSpace(right)) {
		t.Errorf("right-side metadata lost from trimmed line: %q", stripAnsi(line))
	}
}

// TestBuildSeparatorLine_ShortTitleUnchanged verifies the non-truncating
// path still produces a perfectly-sized separator.
func TestBuildSeparatorLine_ShortTitleUnchanged(t *testing.T) {
	gui := &Gui{}
	width := 40

	line := gui.buildSeparatorLine(true, " Short ", "", width, false)
	if w := visibleWidth(line); w != width {
		t.Errorf("visible width = %d, want %d", w, width)
	}
	if strings.Contains(stripAnsi(line), "…") {
		t.Errorf("short title should not be truncated: %q", stripAnsi(line))
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

	gui := &Gui{
		state:      NewGuiState(),
		contextMgr: NewContextMgr(),
		contexts: &context.ContextTree{
			Compose:          context.NewComposeContext(),
			CardList:         context.NewCardListContext(),
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

// TestBuildCardContent_ComposePickEmbedIdentity covers the pick-embed
// pattern where the CLI prefixes every extracted line with "- " as a
// list item. Forward-scan must strip that decoration to match the source
// line; otherwise Enter/e/E silently no-op on the extracted content.
func TestBuildCardContent_ComposePickEmbedIdentity(t *testing.T) {
	dir := t.TempDir()

	// Source note — a meeting note with an extractable #work line.
	sourcePath := filepath.Join(dir, "meeting-053.md")
	os.WriteFile(sourcePath, []byte(
		"---\nuuid: test-uuid-053\n---\n"+
			"# Meeting - Feature flag rollout plan\n"+
			"\n"+
			"Discussed timelines and priorities. #work\n",
	), 0644)

	// Composed output as pick produces it: H2 header injected, content
	// line prefixed with "- " list marker.
	composedContent := "## Meeting - Feature flag rollout plan\n" +
		"- Discussed timelines and priorities. #work"

	sourceMap := []models.SourceMapEntry{
		{UUID: "test-uuid-053", Path: sourcePath, Title: "Meeting", StartLine: 1, EndLine: 1},
		{UUID: "test-uuid-053", Path: sourcePath, Title: "Meeting", StartLine: 2, EndLine: 2},
	}

	gui := &Gui{
		state:      NewGuiState(),
		contextMgr: NewContextMgr(),
		contexts: &context.ContextTree{
			Compose:          context.NewComposeContext(),
			CardList:         context.NewCardListContext(),
			ActivePreviewKey: "compose",
		},
	}
	gui.contexts.Compose.Note = models.Note{
		UUID:    "parent-uuid",
		Path:    filepath.Join(dir, "parent.md"),
		Content: composedContent,
	}
	gui.contexts.Compose.SourceMap = sourceMap
	gui.contexts.Compose.RenderMarkdown = false

	lines := gui.BuildCardContent(gui.contexts.Compose.Note, 80)

	// Source file content lines (after frontmatter):
	//   L1: # Meeting - Feature flag rollout plan
	//   L2: (blank)
	//   L3: Discussed timelines and priorities. #work
	want := map[string]int{
		"## Meeting - Feature flag rollout plan":      1,
		"- Discussed timelines and priorities. #work": 3,
	}
	for text, wantLine := range want {
		found := false
		for _, sl := range lines {
			trimmed := stripAnsi(sl.Text)
			if len(trimmed) > 0 {
				trimmed = trimmed[1:] // strip leading space added by BuildCardContent
			}
			if trimmed == text {
				found = true
				if sl.LineNum != wantLine {
					t.Errorf("line %q: LineNum = %d, want %d", text, sl.LineNum, wantLine)
				}
				if sl.UUID != "test-uuid-053" {
					t.Errorf("line %q: UUID = %q, want test-uuid-053", text, sl.UUID)
				}
				break
			}
		}
		if !found {
			t.Errorf("line %q not found in SourceLines", text)
		}
	}
}

func TestNormalizeLineForMatch(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		// Header decoration
		{"# Foo", "Foo"},
		{"### Foo", "Foo"},
		// List bullet decoration
		{"- Foo", "Foo"},
		{"* Foo", "Foo"},
		{"+ Foo", "Foo"},
		// Blockquote decoration
		{"> Foo", "Foo"},
		// Task checkbox decoration
		{"[ ] Foo", "Foo"},
		{"[x] Foo", "Foo"},
		{"[X] Foo", "Foo"},
		// Header + list decoration both strip (both directions symmetric)
		{"## - Foo", "Foo"},
		// Identity preservation
		{"Foo bar", "Foo bar"},
		{"", ""},
		// Non-decoration: single-char lines, ordered-list markers
		{"-", "-"},
		{"1. Foo", "1. Foo"},
		// stripHeaderPrefix strips leading '#' regardless of trailing space
		// (so "#tag" → "tag") — the forward-scan matcher applies this
		// symmetrically to both sides, so match stability is preserved.
		{"#tag", "tag"},
	}
	for _, tt := range tests {
		got := normalizeLineForMatch(tt.in)
		if got != tt.want {
			t.Errorf("normalizeLineForMatch(%q) = %q, want %q", tt.in, got, tt.want)
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

// buildCardContentFixture constructs a minimal Gui suitable for calling
// BuildCardContent directly, without any of the GUI lifecycle. Returns the
// gui and the note whose content is the provided `content` string.
func buildCardContentFixture(t *testing.T, content string) (*Gui, models.Note) {
	t.Helper()
	dir := t.TempDir()
	notePath := filepath.Join(dir, "note.md")
	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	gui := &Gui{
		state:      NewGuiState(),
		contextMgr: NewContextMgr(),
		contexts: &context.ContextTree{
			CardList:         context.NewCardListContext(),
			Compose:          context.NewComposeContext(),
			ActivePreviewKey: "cardList",
		},
	}
	// Disable markdown rendering so tests don't require chroma init.
	gui.contexts.CardList.RenderMarkdown = false

	note := models.Note{
		UUID:    "test-uuid",
		Path:    notePath,
		Title:   "Test",
		Content: content,
	}
	return gui, note
}

func extractTexts(lines []types.SourceLine) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimSpace(stripAnsi(l.Text)))
	}
	return out
}

func TestBuildCardContent_DimDoneSection(t *testing.T) {
	content := strings.Join([]string{
		"# Wrap up #done",
		"still linked",
		"more stuff",
	}, "\n")
	gui, note := buildCardContentFixture(t, content)
	ds := gui.contexts.ActivePreview().DisplayState()
	ds.DimDone = true
	ds.HideDone = false

	lines := gui.BuildCardContent(note, 80)
	if len(lines) == 0 {
		t.Fatal("expected at least one rendered line")
	}
	for _, l := range lines {
		if !strings.Contains(l.Text, AnsiDim) {
			t.Errorf("expected every line in done section to contain AnsiDim, got %q", l.Text)
		}
	}
}

func TestBuildCardContent_HideDoneSection(t *testing.T) {
	content := strings.Join([]string{
		"# Done bucket #done",
		"finished item",
		"# Active",
		"keep working",
	}, "\n")
	gui, note := buildCardContentFixture(t, content)
	ds := gui.contexts.ActivePreview().DisplayState()
	ds.HideDone = true

	lines := gui.BuildCardContent(note, 80)
	texts := extractTexts(lines)

	for _, want := range []string{"# Active", "keep working"} {
		found := false
		for _, got := range texts {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to see %q in output, got %v", want, texts)
		}
	}
	for _, banned := range []string{"# Done bucket #done", "finished item"} {
		for _, got := range texts {
			if got == banned {
				t.Errorf("expected %q to be hidden, still appeared in %v", banned, texts)
			}
		}
	}
}

func TestBuildCardContent_HideDoneCollapsesWhitespace(t *testing.T) {
	content := strings.Join([]string{
		"# title",
		"",
		"hidden line #done",
		"",
		"Second line",
	}, "\n")
	gui, note := buildCardContentFixture(t, content)
	ds := gui.contexts.ActivePreview().DisplayState()
	ds.HideDone = true

	lines := gui.BuildCardContent(note, 80)
	texts := extractTexts(lines)

	// The surviving body should be: "# title", one blank, "Second line".
	// Without collapse we'd have two blanks between them, but leading
	// blanks are trimmed by BuildCardContent — the regression is the
	// trailing gap between "# title" and "Second line".
	want := []string{"# title", "", "Second line"}
	if len(texts) != len(want) {
		t.Fatalf("text count = %d, want %d. got: %v", len(texts), len(want), texts)
	}
	for i, w := range want {
		if texts[i] != w {
			t.Errorf("line %d = %q, want %q (full: %v)", i, texts[i], w, texts)
		}
	}
}

func TestBuildCardContent_HideDoneSafety_PreservesFirstLine(t *testing.T) {
	// Entire body is #done — without safety the body would collapse to nothing.
	content := strings.Join([]string{
		"only line #done",
	}, "\n")
	gui, note := buildCardContentFixture(t, content)
	ds := gui.contexts.ActivePreview().DisplayState()
	ds.HideDone = true

	lines := gui.BuildCardContent(note, 80)
	if len(lines) == 0 {
		t.Fatal("hide safety should preserve at least one line")
	}
	first := strings.TrimSpace(stripAnsi(lines[0].Text))
	if first != "only line #done" {
		t.Errorf("first preserved line = %q, want %q", first, "only line #done")
	}
	if !strings.Contains(lines[0].Text, AnsiDim) {
		t.Errorf("safety-preserved line should be dimmed, got %q", lines[0].Text)
	}
}

func TestDimLine(t *testing.T) {
	// Plain text gets wrapped in dim.
	got := dimLine(" hello")
	if got != AnsiDim+" hello"+AnsiReset {
		t.Errorf("dimLine plain: got %q", got)
	}

	// Mid-line resets are patched so dim survives chroma highlighting.
	input := AnsiGreen + "- " + AnsiReset + "task"
	got = dimLine(input)
	want := AnsiDim + AnsiGreen + "- " + AnsiReset + AnsiDim + "task" + AnsiReset
	if got != want {
		t.Errorf("dimLine with reset:\ngot  %q\nwant %q", got, want)
	}
}
