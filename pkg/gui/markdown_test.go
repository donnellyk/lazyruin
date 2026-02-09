package gui

import "testing"

func TestMarkdownContinuation_BulletList(t *testing.T) {
	cont := markdownContinuation("- some item")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "- " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "- ")
	}
	if cont.Empty {
		t.Error("should not be empty")
	}
}

func TestMarkdownContinuation_BulletListEmpty(t *testing.T) {
	cont := markdownContinuation("- ")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if !cont.Empty {
		t.Error("should be empty")
	}
}

func TestMarkdownContinuation_StarList(t *testing.T) {
	cont := markdownContinuation("* item")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "* " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "* ")
	}
}

func TestMarkdownContinuation_TaskList(t *testing.T) {
	cont := markdownContinuation("- [ ] buy milk")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "- [ ] " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "- [ ] ")
	}
	if cont.Empty {
		t.Error("should not be empty")
	}
}

func TestMarkdownContinuation_CheckedTask(t *testing.T) {
	cont := markdownContinuation("- [x] done item")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "- [ ] " {
		t.Errorf("Prefix = %q, want %q (should be unchecked)", cont.Prefix, "- [ ] ")
	}
}

func TestMarkdownContinuation_TaskEmpty(t *testing.T) {
	cont := markdownContinuation("- [ ] ")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if !cont.Empty {
		t.Error("should be empty")
	}
}

func TestMarkdownContinuation_OrderedList(t *testing.T) {
	cont := markdownContinuation("3. third item")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "4. " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "4. ")
	}
}

func TestMarkdownContinuation_OrderedListEmpty(t *testing.T) {
	cont := markdownContinuation("2. ")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if !cont.Empty {
		t.Error("should be empty")
	}
}

func TestMarkdownContinuation_Blockquote(t *testing.T) {
	cont := markdownContinuation("> some quote")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "> " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "> ")
	}
}

func TestMarkdownContinuation_BlockquoteEmpty(t *testing.T) {
	cont := markdownContinuation("> ")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if !cont.Empty {
		t.Error("should be empty")
	}
}

func TestMarkdownContinuation_Indented(t *testing.T) {
	cont := markdownContinuation("  - nested item")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "  - " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "  - ")
	}
}

func TestMarkdownContinuation_IndentedTask(t *testing.T) {
	cont := markdownContinuation("    - [ ] nested task")
	if cont == nil {
		t.Fatal("expected continuation")
	}
	if cont.Prefix != "    - [ ] " {
		t.Errorf("Prefix = %q, want %q", cont.Prefix, "    - [ ] ")
	}
}

func TestMarkdownContinuation_NoMatch(t *testing.T) {
	tests := []string{
		"hello world",
		"# heading",
		"",
		"---",
		"```code```",
	}
	for _, line := range tests {
		if cont := markdownContinuation(line); cont != nil {
			t.Errorf("line %q: expected nil, got prefix %q", line, cont.Prefix)
		}
	}
}

func TestCurrentLine(t *testing.T) {
	content := "line0\nline1\nline2"
	tests := []struct {
		cy   int
		want string
	}{
		{0, "line0"},
		{1, "line1"},
		{2, "line2"},
		{3, ""},
		{-1, ""},
	}
	for _, tc := range tests {
		got := currentLine(content, tc.cy)
		if got != tc.want {
			t.Errorf("currentLine(_, %d) = %q, want %q", tc.cy, got, tc.want)
		}
	}
}
