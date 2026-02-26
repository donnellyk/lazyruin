package helpers

import (
	"strings"
	"testing"
	"time"
)

func TestAtDateCandidatesAt_NoFilter(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("", now)

	if len(items) == 0 {
		t.Fatal("expected candidates with empty filter")
	}

	// Should include common suggestions
	labels := make(map[string]bool)
	for _, item := range items {
		labels[item.Label] = true
	}
	for _, want := range []string{"@today", "@yesterday", "@tomorrow"} {
		if !labels[want] {
			t.Errorf("missing expected candidate %q", want)
		}
	}
}

func TestAtDateCandidatesAt_FilterMatches(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("to", now)

	if len(items) == 0 {
		t.Fatal("expected candidates matching 'to'")
	}

	for _, item := range items {
		label := strings.TrimPrefix(item.Label, "@")
		if !strings.HasPrefix(label, "to") {
			t.Errorf("candidate %q doesn't match filter 'to'", item.Label)
		}
	}
}

func TestAtDateCandidatesAt_NextWeekExpandsDays(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("next week", now)

	if len(items) != 7 {
		t.Errorf("expected 7 day-of-week candidates for 'next week', got %d", len(items))
	}

	for _, item := range items {
		if !strings.HasPrefix(item.Label, "@next ") {
			t.Errorf("expected '@next *' label, got %q", item.Label)
		}
	}
}

func TestAtDateCandidatesAt_LastWeekExpandsDays(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("last week", now)

	if len(items) != 7 {
		t.Errorf("expected 7 day-of-week candidates for 'last week', got %d", len(items))
	}

	for _, item := range items {
		if !strings.HasPrefix(item.Label, "@last ") {
			t.Errorf("expected '@last *' label, got %q", item.Label)
		}
	}
}

func TestAtDateCandidatesAt_NativeDate_KeepsNaturalInsertText(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("today", now)

	if len(items) == 0 {
		t.Fatal("expected candidates for 'today'")
	}

	found := false
	for _, item := range items {
		if item.Label == "@today" {
			found = true
			if item.InsertText != "@today" {
				t.Errorf("native date 'today' should keep natural insert text, got %q", item.InsertText)
			}
		}
	}
	if !found {
		t.Error("@today not found in candidates")
	}
}

func TestAtDateCandidatesAt_NonNativeDate_ConvertsToISO(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("monday", now)

	found := false
	for _, item := range items {
		if item.Label == "@monday" {
			found = true
			if !strings.HasPrefix(item.InsertText, "@2026-") {
				t.Errorf("non-native date 'monday' should convert to ISO, got %q", item.InsertText)
			}
		}
	}
	if !found {
		t.Error("@monday not found in candidates")
	}
}

func TestAtDateCandidatesAt_USDateFallback(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("2/26/2026", now)

	if len(items) != 1 {
		t.Fatalf("expected 1 candidate for US date, got %d", len(items))
	}
	if items[0].InsertText != "@2026-02-26" {
		t.Errorf("US date insert text = %q, want @2026-02-26", items[0].InsertText)
	}
}

func TestAtDateCandidatesAt_NoMatch(t *testing.T) {
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	items := atDateCandidatesAt("zzzznotadate", now)

	if len(items) != 0 {
		t.Errorf("expected 0 candidates for nonsense, got %d", len(items))
	}
}

func TestParseUSDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"2/26/2026", "2026-02-26", true},
		{"12/31/2025", "2025-12-31", true},
		{"1/1/26", "2026-01-01", true},
		{"not-a-date", "", false},
		{"2026-02-26", "", false}, // ISO format not US
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			parsed, ok := parseUSDate(tc.input)
			if ok != tc.ok {
				t.Errorf("parseUSDate(%q) ok = %v, want %v", tc.input, ok, tc.ok)
				return
			}
			if ok {
				got := parsed.Format("2006-01-02")
				if got != tc.want {
					t.Errorf("parseUSDate(%q) = %q, want %q", tc.input, got, tc.want)
				}
			}
		})
	}
}

func TestAmbientDateCandidates_ValidToken(t *testing.T) {
	fn := AmbientDateCandidates()
	items := fn("today")

	if len(items) != 3 {
		t.Fatalf("expected 3 candidates (created/after/before), got %d", len(items))
	}

	prefixes := []string{"created:", "after:", "before:"}
	for i, item := range items {
		if !strings.HasPrefix(item.Label, prefixes[i]) {
			t.Errorf("item[%d].Label = %q, want prefix %q", i, item.Label, prefixes[i])
		}
	}
}

func TestAmbientDateCandidates_InvalidToken(t *testing.T) {
	fn := AmbientDateCandidates()
	items := fn("notadate_xyz")

	if len(items) != 0 {
		t.Errorf("expected 0 candidates for invalid token, got %d", len(items))
	}
}
