package gui

import "testing"

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
		{"created prefix", "created:7d", 10, "created:7d", 0},
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
		{"created trigger", "created:7d", 10, "created:", "7d", false},
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
