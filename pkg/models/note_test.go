package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNote_ShortDate(t *testing.T) {
	tests := []struct {
		name     string
		created  time.Time
		expected string
	}{
		{
			name:     "January single digit",
			created:  time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: "Jan 05",
		},
		{
			name:     "December double digit",
			created:  time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: "Dec 25",
		},
		{
			name:     "February leap year",
			created:  time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expected: "Feb 29",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			note := Note{Created: tc.created}
			result := note.ShortDate()
			if result != tc.expected {
				t.Errorf("ShortDate() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestNote_TagsString(t *testing.T) {
	tests := []struct {
		name       string
		tags       []string
		inlineTags []string
		expected   string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: "",
		},
		{
			name:     "single tag",
			tags:     []string{"daily"},
			expected: "#daily",
		},
		{
			name:     "multiple tags",
			tags:     []string{"daily", "work", "meeting"},
			expected: "#daily, #work, #meeting",
		},
		{
			name:     "nil tags",
			tags:     nil,
			expected: "",
		},
		{
			name:       "global and inline tags",
			tags:       []string{"meeting", "work"},
			inlineTags: []string{"#followup"},
			expected:   "#meeting, #work, #followup",
		},
		{
			name:       "only inline tags",
			inlineTags: []string{"#todo", "#followup"},
			expected:   "#todo, #followup",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			note := Note{Tags: tc.tags, InlineTags: tc.inlineTags}
			result := note.TagsString()
			if result != tc.expected {
				t.Errorf("TagsString() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestNote_JSONInlineTags(t *testing.T) {
	input := `{"uuid":"abc","title":"Test","tags":["#meeting"],"inline_tags":["#followup","#todo"]}`
	var note Note
	if err := json.Unmarshal([]byte(input), &note); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if len(note.Tags) != 1 || note.Tags[0] != "#meeting" {
		t.Errorf("Tags = %v, want [#meeting]", note.Tags)
	}
	if len(note.InlineTags) != 2 {
		t.Fatalf("InlineTags length = %d, want 2", len(note.InlineTags))
	}
	if note.InlineTags[0] != "#followup" || note.InlineTags[1] != "#todo" {
		t.Errorf("InlineTags = %v, want [#followup #todo]", note.InlineTags)
	}
}

func TestHasDoneTag(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"some text #done", true},
		{"- task #done", true},
		{"- [ ] todo #done", true},
		{"- [x] checked #done", true},
		{"#done at start", true},
		{"#DONE uppercase", true},
		{"#Done mixed case", true},
		{"no done tag", false},
		{"#doneForToday is not done", false},
		{"", false},
		{"#doing something", false},
		{"undone", false},
		// notetext integration: tags inside code spans and markdown links
		// are no longer counted, matching what `ruin log` actually records.
		{"see `#done` in code span", false},
		{"[link text #done](https://example.com)", false},
		{"plain #done beats code `#nope`", true},
		// CLI tag semantics allow `-` and `_` inside tag words, so
		// `#done-later` and `#done_soon` are distinct tags — neither
		// matches `#done`. Match the CLI verbatim.
		{"follow up #done-later", false},
		{"revisit #done_soon maybe", false},
		// A trailing `-` is stripped by the CLI, so `#done-` extracts
		// as `#done`.
		{"wrap it up #done-", true},
	}
	for _, tt := range tests {
		got := HasDoneTag(tt.line)
		if got != tt.want {
			t.Errorf("HasDoneTag(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestIsCheckedTodo(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		// Checked forms — matches the ruin CLI's todos.go regex
		// (`^(\s*[-*]\s+)\[([ xX])\]\s+(.*)`).
		{"- [x] done task", true},
		{"- [X] upper-case check", true},
		{"* [x] star bullet", true},
		{"  - [x] indented", true},
		{"\t- [x] tab-indented", true},
		// Unchecked and non-checkbox
		{"- [ ] todo", false},
		{"- [] malformed", false},
		{"- [ x] stray space", false},
		{"plain text", false},
		{"", false},
		// Not a list item
		{"[x] no bullet", false},
		// Preview dim should not fire for code-span checkboxes, but those
		// wouldn't pass the "start of line" anchor, so they're excluded
		// naturally.
		{"see `- [x] inline`", false},
	}
	for _, tt := range tests {
		got := IsCheckedTodo(tt.line)
		if got != tt.want {
			t.Errorf("IsCheckedTodo(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestNote_IsLink(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want bool
	}{
		{"hash-prefixed #link tag", []string{"#link"}, true},
		{"bare link tag", []string{"link"}, true},
		{"uppercase #LINK tag", []string{"#LINK"}, true},
		{"mixed case #Link tag", []string{"#Link"}, true},
		{"no link tag", []string{"daily", "work"}, false},
		{"empty tags", []string{}, false},
		{"nil tags", nil, false},
		{"link among other tags", []string{"daily", "#link", "work"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Note{Tags: tt.tags}
			if got := n.IsLink(); got != tt.want {
				t.Errorf("IsLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNote_URL(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		content string
		want    string
	}{
		{
			name:    "link note with https URL",
			tags:    []string{"link"},
			content: "https://example.com\nsome description",
			want:    "https://example.com",
		},
		{
			name:    "link note with http URL",
			tags:    []string{"#link"},
			content: "http://example.com",
			want:    "http://example.com",
		},
		{
			name:    "link note with non-URL content",
			tags:    []string{"link"},
			content: "just some text",
			want:    "",
		},
		{
			name:    "non-link note with URL content",
			tags:    []string{"daily"},
			content: "https://example.com",
			want:    "",
		},
		{
			name:    "link note with empty content",
			tags:    []string{"link"},
			content: "",
			want:    "",
		},
		{
			name:    "link note with leading blank lines then URL",
			tags:    []string{"link"},
			content: "\n\nhttps://example.com",
			want:    "https://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Note{Tags: tt.tags, Content: tt.content}
			if got := n.URL(); got != tt.want {
				t.Errorf("URL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNote_JSONNoInlineTags(t *testing.T) {
	input := `{"uuid":"abc","title":"Test","tags":["#meeting"]}`
	var note Note
	if err := json.Unmarshal([]byte(input), &note); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if note.InlineTags != nil {
		t.Errorf("InlineTags = %v, want nil", note.InlineTags)
	}
}
