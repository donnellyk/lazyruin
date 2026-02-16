package gui

import (
	"testing"
	"time"
)

func TestCursorBytePos(t *testing.T) {
	tests := []struct {
		name    string
		content string
		cx, cy  int
		want    int
	}{
		{"empty string", "", 0, 0, 0},
		{"start of single line", "hello", 0, 0, 0},
		{"middle of single line", "hello", 3, 0, 3},
		{"end of single line", "hello", 5, 0, 5},
		{"start of second line", "hello\nworld", 0, 1, 6},
		{"middle of second line", "hello\nworld", 3, 1, 9},
		{"cx past end of line", "hello\nworld", 10, 0, 5},
		{"cy past end of content", "hello", 0, 5, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cursorBytePos(tc.content, tc.cx, tc.cy)
			if got != tc.want {
				t.Errorf("cursorBytePos(%q, %d, %d) = %d, want %d", tc.content, tc.cx, tc.cy, got, tc.want)
			}
		})
	}
}

func TestExtractTokenAtCursor(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		cursorPos int
		wantToken string
		wantStart int
	}{
		{"empty", "", 0, "", 0},
		{"single word", "hello", 5, "hello", 0},
		{"second word", "hello world", 11, "world", 6},
		{"mid word", "hello world", 8, "wo", 6},
		{"after space", "hello ", 6, "", 6},
		{"hash tag", "#project", 8, "#project", 0},
		{"hash tag after space", "search #project", 15, "#project", 7},
		{"created prefix", "created:today", 13, "created:today", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, start := extractTokenAtCursor(tc.content, tc.cursorPos)
			if token != tc.wantToken {
				t.Errorf("token = %q, want %q", token, tc.wantToken)
			}
			if start != tc.wantStart {
				t.Errorf("start = %d, want %d", start, tc.wantStart)
			}
		})
	}
}

func TestDetectTrigger(t *testing.T) {
	triggers := []CompletionTrigger{
		{Prefix: "#", Candidates: func(filter string) []CompletionItem { return nil }},
		{Prefix: "created:", Candidates: func(filter string) []CompletionItem { return nil }},
	}

	tests := []struct {
		name       string
		content    string
		cursorPos  int
		wantPrefix string
		wantFilter string
		wantNil    bool
	}{
		{"hash trigger", "#pro", 4, "#", "pro", false},
		{"hash trigger empty filter", "#", 1, "#", "", false},
		{"created trigger", "created:today", 13, "created:", "today", false},
		{"created trigger empty filter", "created:", 8, "created:", "", false},
		{"no trigger", "hello", 5, "", "", true},
		{"hash after space", "search #tag", 11, "#", "tag", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trigger, filter, _ := detectTrigger(tc.content, tc.cursorPos, triggers)
			if tc.wantNil {
				if trigger != nil {
					t.Errorf("expected nil trigger, got prefix %q", trigger.Prefix)
				}
				return
			}
			if trigger == nil {
				t.Fatal("expected non-nil trigger")
			}
			if trigger.Prefix != tc.wantPrefix {
				t.Errorf("prefix = %q, want %q", trigger.Prefix, tc.wantPrefix)
			}
			if filter != tc.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tc.wantFilter)
			}
		})
	}
}

func TestCompletionDown(t *testing.T) {
	state := &CompletionState{
		Active:        true,
		Items:         []CompletionItem{{Label: "a"}, {Label: "b"}, {Label: "c"}},
		SelectedIndex: 0,
	}

	completionDown(state)
	if state.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", state.SelectedIndex)
	}

	completionDown(state)
	if state.SelectedIndex != 2 {
		t.Errorf("SelectedIndex = %d, want 2", state.SelectedIndex)
	}

	// Should not go past the end
	completionDown(state)
	if state.SelectedIndex != 2 {
		t.Errorf("SelectedIndex = %d, want 2 (clamped)", state.SelectedIndex)
	}
}

func TestCompletionUp(t *testing.T) {
	state := &CompletionState{
		Active:        true,
		Items:         []CompletionItem{{Label: "a"}, {Label: "b"}, {Label: "c"}},
		SelectedIndex: 2,
	}

	completionUp(state)
	if state.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", state.SelectedIndex)
	}

	completionUp(state)
	if state.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", state.SelectedIndex)
	}

	// Should not go below 0
	completionUp(state)
	if state.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (clamped)", state.SelectedIndex)
	}
}

func TestCompletionDown_Inactive(t *testing.T) {
	state := &CompletionState{
		Active:        false,
		Items:         []CompletionItem{{Label: "a"}},
		SelectedIndex: 0,
	}

	completionDown(state)
	if state.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (should not change when inactive)", state.SelectedIndex)
	}
}

func TestExtractSort(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantQuery string
		wantSort  string
	}{
		{"no sort", "#log", "#log", ""},
		{"sort only", "sort:created:desc", "", "created:desc"},
		{"query with sort", "#log sort:order:asc", "#log", "order:asc"},
		{"sort in middle", "#log sort:title:asc #project", "#log #project", "title:asc"},
		{"multiple sorts last wins", "sort:created:asc sort:updated:desc", "", "updated:desc"},
		{"empty", "", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, sort := extractSort(tc.input)
			if query != tc.wantQuery {
				t.Errorf("query = %q, want %q", query, tc.wantQuery)
			}
			if sort != tc.wantSort {
				t.Errorf("sort = %q, want %q", sort, tc.wantSort)
			}
		})
	}
}

func TestExtractHeaders(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []headerInfo
	}{
		{
			"basic headings",
			"# Title\n## Section\n### Sub",
			[]headerInfo{{1, "Title"}, {2, "Section"}, {3, "Sub"}},
		},
		{
			"skips code blocks",
			"# Real\n```\n# Not a heading\n```\n## Also Real",
			[]headerInfo{{1, "Real"}, {2, "Also Real"}},
		},
		{
			"skips empty headings",
			"# \n## Valid",
			[]headerInfo{{2, "Valid"}},
		},
		{
			"all levels",
			"# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6",
			[]headerInfo{{1, "H1"}, {2, "H2"}, {3, "H3"}, {4, "H4"}, {5, "H5"}, {6, "H6"}},
		},
		{
			"non-heading hash lines",
			"not a heading\n#tag is not a heading",
			nil,
		},
		{
			"empty content",
			"",
			nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractHeaders(tc.content)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d headers, want %d: %v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("header[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestExtractParentPath(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantRemain string
		wantPath   string
	}{
		{"no parent path", "#ruin, #log", "#ruin, #log", ""},
		{"parent path only", ">log/daily", "", "log/daily"},
		{"parent path with text", "#ruin, #log >log/daily", "#ruin, #log", "log/daily"},
		{"parent path at start", ">projects #work", "#work", "projects"},
		{"multiple tokens", "#a #b >foo/bar #c", "#a #b #c", "foo/bar"},
		{"empty text", "", "", ""},
		{"bare >", ">", ">", ""},
		{"parent with spaces", ">My Project", "", "My Project"},
		{"parent with spaces and text", "#ruin >My Project/Daily Log #work", "#ruin #work", "My Project/Daily Log"},
		{"parent with spaces before hash", "#a >Parent Name #b", "#a #b", "Parent Name"},
		{"double arrow with spaces", ">>All Notes/My Note", "", "All Notes/My Note"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			remain, path := extractParentPath(tc.text)
			if remain != tc.wantRemain {
				t.Errorf("remaining = %q, want %q", remain, tc.wantRemain)
			}
			if path != tc.wantPath {
				t.Errorf("parentPath = %q, want %q", path, tc.wantPath)
			}
		})
	}
}

func TestDetectTrigger_SpaceAwareDates(t *testing.T) {
	triggers := []CompletionTrigger{
		{Prefix: "#", Candidates: func(filter string) []CompletionItem { return nil }},
		{Prefix: "created:", Candidates: func(filter string) []CompletionItem { return nil }},
		{Prefix: "before:", Candidates: func(filter string) []CompletionItem { return nil }},
		{Prefix: "after:", Candidates: func(filter string) []CompletionItem { return nil }},
		{Prefix: "between:", Candidates: func(filter string) []CompletionItem { return nil }},
		{Prefix: "@", Candidates: func(filter string) []CompletionItem { return nil }},
	}

	tests := []struct {
		name       string
		content    string
		cursorPos  int
		wantPrefix string
		wantFilter string
	}{
		{"created with spaces", "created:next week", 17, "created:", "next week"},
		{"created with spaces after other token", "#tag created:last monday", 24, "created:", "last monday"},
		{"before with spaces", "before:next friday", 18, "before:", "next friday"},
		{"after with spaces", "after:last month", 16, "after:", "last month"},
		{"between with spaces", "between:last week", 17, "between:", "last week"},
		// Single-word still works via extractTokenAtCursor (no fallback needed)
		{"created single word", "created:today", 13, "created:", "today"},
		// @ trigger with spaces
		{"at with spaces", "@next friday", 12, "@", "next friday"},
		{"at with spaces after text", "meeting @last monday", 20, "@", "last monday"},
		{"at single word", "@today", 6, "@", "today"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trigger, filter, _ := detectTrigger(tc.content, tc.cursorPos, triggers)
			if trigger == nil {
				t.Fatal("expected non-nil trigger")
			}
			if trigger.Prefix != tc.wantPrefix {
				t.Errorf("prefix = %q, want %q", trigger.Prefix, tc.wantPrefix)
			}
			if filter != tc.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tc.wantFilter)
			}
		})
	}
}

func TestDateCandidates(t *testing.T) {
	// Use a fixed time for deterministic tests
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		filter     string
		wantCount  int
		wantLabels []string // subset check
	}{
		{"empty filter shows literals", "", 3, []string{"created:today", "created:yesterday", "created:tomorrow"}},
		{"to prefix matches today and tomorrow", "to", 2, []string{"created:today", "created:tomorrow"}},
		{"exact literal no anytime", "today", 1, []string{"created:today"}},
		{"next friday parses", "next friday", 1, []string{"created:2026-02-20"}},
		{"nonsense returns nothing", "xyzzy123", 0, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			items := dateCandidatesAt("created:", tc.filter, now)
			if len(items) != tc.wantCount {
				labels := make([]string, len(items))
				for i, it := range items {
					labels[i] = it.Label
				}
				t.Fatalf("got %d items %v, want %d", len(items), labels, tc.wantCount)
			}
			for i, wantLabel := range tc.wantLabels {
				if i >= len(items) {
					break
				}
				if items[i].Label != wantLabel {
					t.Errorf("item[%d].Label = %q, want %q", i, items[i].Label, wantLabel)
				}
			}
		})
	}
}

func TestBetweenCandidates(t *testing.T) {
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)

	// Empty filter returns 3 preset ranges
	items := betweenCandidatesAt("", now)
	if len(items) != 3 {
		t.Fatalf("empty filter: got %d items, want 3", len(items))
	}
	// First item should be 7-day range
	want := "between:2026-02-09,2026-02-16"
	if items[0].InsertText != want {
		t.Errorf("item[0].InsertText = %q, want %q", items[0].InsertText, want)
	}

	// Natural language filter
	items = betweenCandidatesAt("last monday", now)
	if len(items) != 1 {
		t.Fatalf("'last monday' filter: got %d items, want 1", len(items))
	}
	if items[0].InsertText == "" {
		t.Error("expected non-empty InsertText for anytime parse")
	}
}

func TestAtDateCandidates(t *testing.T) {
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		filter     string
		wantCount  int
		wantLabels []string
		wantInsert []string
	}{
		{"empty filter shows all suggestions", "", len(atDateSuggestions),
			[]string{"@today", "@yesterday", "@tomorrow"}, // check first 3
			[]string{"@today", "@yesterday", "@tomorrow"}},
		{"to prefix matches today and tomorrow", "to", 2,
			[]string{"@today", "@tomorrow"},
			[]string{"@today", "@tomorrow"}},
		{"today exact", "today", 1, []string{"@today"}, []string{"@today"}},
		{"next friday from suggestions", "next friday", 1,
			[]string{"@next friday"}, []string{"@2026-02-20"}},
		{"next week expands to 7 days", "next week", 7,
			[]string{"@next sunday", "@next monday"}, nil},
		{"last week expands to 7 days", "last week", 7,
			[]string{"@last sunday", "@last monday"}, nil},
		{"month name", "jan", 1, []string{"@january"}, []string{"@2027-01-01"}},
		{"freeform fallback", "3 days ago", 1, nil, nil},
		{"US date MM/DD/YYYY", "02/20/2026", 1, []string{"@2026-02-20"}, []string{"@2026-02-20"}},
		{"US date MM/DD/YY", "2/20/26", 1, []string{"@2026-02-20"}, []string{"@2026-02-20"}},
		{"nonsense returns nothing", "xyzzy123", 0, nil, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			items := atDateCandidatesAt(tc.filter, now)
			if len(items) != tc.wantCount {
				labels := make([]string, len(items))
				for i, it := range items {
					labels[i] = it.Label
				}
				t.Fatalf("got %d items %v, want %d", len(items), labels, tc.wantCount)
			}
			for i, wantLabel := range tc.wantLabels {
				if items[i].Label != wantLabel {
					t.Errorf("item[%d].Label = %q, want %q", i, items[i].Label, wantLabel)
				}
			}
			for i, wantInsert := range tc.wantInsert {
				if items[i].InsertText != wantInsert {
					t.Errorf("item[%d].InsertText = %q, want %q", i, items[i].InsertText, wantInsert)
				}
			}
		})
	}
}

func TestLineContainsAt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		cursor  int
		want    bool
	}{
		{"at on same line", "@friday meeting", 15, true},
		{"at on previous line", "@friday\nmeeting", 15, false},
		{"no at", "meeting friday", 14, false},
		{"at after cursor", "meeting @friday", 7, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := lineContainsAt(tc.content, tc.cursor)
			if got != tc.want {
				t.Errorf("lineContainsAt(%q, %d) = %v, want %v", tc.content, tc.cursor, got, tc.want)
			}
		})
	}
}

func TestAmbientDateCandidates(t *testing.T) {
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)

	// Valid natural language date returns 3 suggestions (created, after, before)
	items := ambientDateCandidatesAt("friday", now)
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	if items[0].Label != "created:2026-02-20" {
		t.Errorf("item[0].Label = %q, want created:2026-02-20", items[0].Label)
	}
	if items[1].Label != "after:2026-02-20" {
		t.Errorf("item[1].Label = %q, want after:2026-02-20", items[1].Label)
	}
	if items[2].Label != "before:2026-02-20" {
		t.Errorf("item[2].Label = %q, want before:2026-02-20", items[2].Label)
	}

	// Nonsense returns nil
	items = ambientDateCandidatesAt("xyzzy", now)
	if items != nil {
		t.Errorf("expected nil for nonsense, got %d items", len(items))
	}
}

func TestNewCompletionState(t *testing.T) {
	state := NewCompletionState()
	if state.Active {
		t.Error("new state should not be active")
	}
	if state.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", state.SelectedIndex)
	}
	if state.Items != nil {
		t.Error("Items should be nil")
	}
}
