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

func TestNote_FirstLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "empty content", content: "", want: ""},
		{name: "single line", content: "hello world", want: "hello world"},
		{name: "multiline", content: "first\nsecond\nthird", want: "first"},
		{name: "leading blank lines", content: "\n\n  \nactual first", want: "actual first"},
		{name: "all blank lines", content: "\n  \n\t\n", want: ""},
		{name: "whitespace trimmed", content: "  padded  ", want: "padded"},
		{name: "only whitespace line then content", content: "   \ncontent here", want: "content here"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			note := Note{Content: tc.content}
			got := note.FirstLine()
			if got != tc.want {
				t.Errorf("FirstLine() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestJoinDot(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{name: "no parts", parts: nil, want: ""},
		{name: "single part", parts: []string{"hello"}, want: "hello"},
		{name: "two parts", parts: []string{"a", "b"}, want: "a · b"},
		{name: "three parts", parts: []string{"a", "b", "c"}, want: "a · b · c"},
		{name: "empty parts filtered", parts: []string{"a", "", "b", "", "c"}, want: "a · b · c"},
		{name: "all empty", parts: []string{"", "", ""}, want: ""},
		{name: "single non-empty among empties", parts: []string{"", "only", ""}, want: "only"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := JoinDot(tc.parts...)
			if got != tc.want {
				t.Errorf("JoinDot(%v) = %q, want %q", tc.parts, got, tc.want)
			}
		})
	}
}

func TestNote_GlobalTagsString(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{name: "empty tags", tags: nil, want: ""},
		{name: "single tag without hash", tags: []string{"daily"}, want: "#daily"},
		{name: "single tag with hash", tags: []string{"#daily"}, want: "#daily"},
		{name: "multiple tags", tags: []string{"daily", "work"}, want: "#daily, #work"},
		{name: "mixed hash prefixes", tags: []string{"#meeting", "work"}, want: "#meeting, #work"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			note := Note{Tags: tc.tags}
			got := note.GlobalTagsString()
			if got != tc.want {
				t.Errorf("GlobalTagsString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNote_ShortDate_Zero(t *testing.T) {
	note := Note{}
	if got := note.ShortDate(); got != "" {
		t.Errorf("ShortDate() on zero time = %q, want empty", got)
	}
}
