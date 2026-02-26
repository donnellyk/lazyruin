package helpers

import "testing"

func TestExtractSort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantQ    string
		wantSort string
	}{
		{name: "no sort token", input: "#todo today", wantQ: "#todo today", wantSort: ""},
		{name: "sort at end", input: "#todo sort:created", wantQ: "#todo", wantSort: "created"},
		{name: "sort at beginning", input: "sort:updated #meeting", wantQ: "#meeting", wantSort: "updated"},
		{name: "sort in middle", input: "hello sort:title world", wantQ: "hello world", wantSort: "title"},
		{name: "only sort", input: "sort:created", wantQ: "", wantSort: "created"},
		{name: "empty input", input: "", wantQ: "", wantSort: ""},
		{name: "multiple sort tokens takes last", input: "sort:a sort:b", wantQ: "", wantSort: "b"},
		{name: "sort with no value", input: "sort:", wantQ: "", wantSort: ""},
		{name: "sort-like but not prefix", input: "nosort:created", wantQ: "nosort:created", wantSort: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotQ, gotSort := extractSort(tc.input)
			if gotQ != tc.wantQ {
				t.Errorf("extractSort(%q) query = %q, want %q", tc.input, gotQ, tc.wantQ)
			}
			if gotSort != tc.wantSort {
				t.Errorf("extractSort(%q) sort = %q, want %q", tc.input, gotSort, tc.wantSort)
			}
		})
	}
}
