package context

import "testing"

func TestFilterActive(t *testing.T) {
	s := &CardListState{}
	if s.FilterActive() {
		t.Fatal("expected FilterActive false on zero value")
	}
	s.FilterText = "#bug"
	if !s.FilterActive() {
		t.Fatal("expected FilterActive true when FilterText is set")
	}
}

func TestClearFilter(t *testing.T) {
	s := &CardListState{
		FilterText:      "#bug",
		UnfilteredCount: 10,
	}
	s.ClearFilter()
	if s.FilterText != "" {
		t.Fatalf("expected FilterText empty, got %q", s.FilterText)
	}
	if s.UnfilteredCount != 0 {
		t.Fatalf("expected UnfilteredCount 0, got %d", s.UnfilteredCount)
	}
}

func TestPickResultsFilterActive(t *testing.T) {
	s := &PickResultsState{}
	if s.FilterActive() {
		t.Fatal("expected FilterActive false on zero value")
	}
	s.FilterText = "#bug"
	if !s.FilterActive() {
		t.Fatal("expected FilterActive true when FilterText is set")
	}
}

func TestPickResultsClearFilter(t *testing.T) {
	s := &PickResultsState{
		FilterText:      "#bug",
		UnfilteredCount: 10,
	}
	s.ClearFilter()
	if s.FilterText != "" {
		t.Fatalf("expected FilterText empty, got %q", s.FilterText)
	}
	if s.UnfilteredCount != 0 {
		t.Fatalf("expected UnfilteredCount 0, got %d", s.UnfilteredCount)
	}
}
