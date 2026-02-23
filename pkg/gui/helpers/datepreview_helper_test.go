package helpers

import (
	"testing"
	"time"

	"kvnd/lazyruin/pkg/models"
)

func TestISOWeekday(t *testing.T) {
	tests := []struct {
		day  time.Weekday
		want int
	}{
		{time.Monday, 1},
		{time.Tuesday, 2},
		{time.Wednesday, 3},
		{time.Thursday, 4},
		{time.Friday, 5},
		{time.Saturday, 6},
		{time.Sunday, 7},
	}
	for _, tt := range tests {
		if got := ISOWeekday(tt.day); got != tt.want {
			t.Errorf("ISOWeekday(%v) = %d, want %d", tt.day, got, tt.want)
		}
	}
}

func TestCurrentWeekday(t *testing.T) {
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	got := CurrentWeekday(today.Weekday())
	if got != todayStr {
		t.Errorf("CurrentWeekday(today's weekday) = %s, want %s", got, todayStr)
	}

	for _, wd := range []time.Weekday{
		time.Monday, time.Tuesday, time.Wednesday, time.Thursday,
		time.Friday, time.Saturday, time.Sunday,
	} {
		result := CurrentWeekday(wd)
		parsed, err := time.Parse("2006-01-02", result)
		if err != nil {
			t.Errorf("CurrentWeekday(%v) returned invalid date: %s", wd, result)
			continue
		}
		if parsed.Weekday() != wd {
			t.Errorf("CurrentWeekday(%v) = %s (weekday %v), wrong weekday", wd, result, parsed.Weekday())
		}
		diff := parsed.Sub(today).Hours() / 24
		if diff < -6 || diff > 6 {
			t.Errorf("CurrentWeekday(%v) = %s, more than 6 days from today", wd, result)
		}
	}
}

func TestDeduplicateNotes(t *testing.T) {
	created := []models.Note{
		{UUID: "a", Title: "Note A"},
		{UUID: "b", Title: "Note B"},
	}
	updated := []models.Note{
		{UUID: "b", Title: "Note B (updated)"},
		{UUID: "c", Title: "Note C"},
	}
	result := DeduplicateNotes(created, updated)
	if len(result) != 3 {
		t.Fatalf("expected 3 notes, got %d", len(result))
	}
	if result[0].UUID != "a" || result[1].UUID != "b" || result[2].UUID != "c" {
		t.Errorf("wrong UUIDs: %v, %v, %v", result[0].UUID, result[1].UUID, result[2].UUID)
	}
	if result[1].Title != "Note B" {
		t.Errorf("duplicate should keep created version, got title=%s", result[1].Title)
	}
}

func TestDeduplicateNotes_Empty(t *testing.T) {
	result := DeduplicateNotes(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected 0 notes, got %d", len(result))
	}
}

func TestFilterOutTodoLines(t *testing.T) {
	results := []models.PickResult{
		{
			UUID:  "note1",
			Title: "Note 1",
			Matches: []models.PickMatch{
				{Line: 1, Content: "#followup @2026-02-23"},
				{Line: 2, Content: "- [ ] todo @2026-02-23"},
				{Line: 3, Content: "- [x] done todo @2026-02-23"},
				{Line: 4, Content: "another line #tag @2026-02-23"},
			},
		},
		{
			UUID:  "note2",
			Title: "Note 2",
			Matches: []models.PickMatch{
				{Line: 1, Content: "- [ ] only todo"},
			},
		},
	}
	filtered := filterOutTodoLines(results)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 result group after filtering, got %d", len(filtered))
	}
	if len(filtered[0].Matches) != 2 {
		t.Fatalf("expected 2 matches after filtering, got %d", len(filtered[0].Matches))
	}
	if filtered[0].Matches[0].Line != 1 || filtered[0].Matches[1].Line != 4 {
		t.Errorf("wrong matches kept: lines %d and %d", filtered[0].Matches[0].Line, filtered[0].Matches[1].Line)
	}
}

func TestFilterOutTodoLines_WithWhitespace(t *testing.T) {
	results := []models.PickResult{
		{
			UUID: "note1",
			Matches: []models.PickMatch{
				{Line: 1, Content: "  - [ ] indented todo"},
				{Line: 2, Content: "not a todo"},
			},
		},
	}
	filtered := filterOutTodoLines(results)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 result, got %d", len(filtered))
	}
	if len(filtered[0].Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(filtered[0].Matches))
	}
	if filtered[0].Matches[0].Line != 2 {
		t.Errorf("expected line 2 kept, got %d", filtered[0].Matches[0].Line)
	}
}

func TestFilterOutTodoLines_AllTodos(t *testing.T) {
	results := []models.PickResult{
		{
			UUID: "note1",
			Matches: []models.PickMatch{
				{Line: 1, Content: "- [ ] todo 1"},
				{Line: 2, Content: "- [x] todo 2"},
			},
		},
	}
	filtered := filterOutTodoLines(results)
	if len(filtered) != 0 {
		t.Errorf("expected 0 results when all matches are todos, got %d", len(filtered))
	}
}

func TestSortDonePicksLast(t *testing.T) {
	results := []models.PickResult{
		{
			UUID: "note1",
			Matches: []models.PickMatch{
				{Line: 1, Content: "active line", Done: false},
				{Line: 2, Content: "done line #done", Done: true},
			},
		},
		{
			UUID: "note2",
			Matches: []models.PickMatch{
				{Line: 5, Content: "all done #done", Done: true},
			},
		},
		{
			UUID: "note3",
			Matches: []models.PickMatch{
				{Line: 3, Content: "all active", Done: false},
			},
		},
	}
	sorted := sortDonePicksLast(results)
	if len(sorted) != 4 {
		t.Fatalf("expected 4 result groups, got %d", len(sorted))
	}
	// Active groups first: note1 (active only), note3
	if sorted[0].UUID != "note1" || len(sorted[0].Matches) != 1 || sorted[0].Matches[0].Done {
		t.Errorf("sorted[0] should be note1 active match, got UUID=%s done=%v", sorted[0].UUID, sorted[0].Matches[0].Done)
	}
	if sorted[1].UUID != "note3" {
		t.Errorf("sorted[1] should be note3, got UUID=%s", sorted[1].UUID)
	}
	// Done groups last: note1 (done), note2
	if sorted[2].UUID != "note1" || !sorted[2].Matches[0].Done {
		t.Errorf("sorted[2] should be note1 done match, got UUID=%s done=%v", sorted[2].UUID, sorted[2].Matches[0].Done)
	}
	if sorted[3].UUID != "note2" || !sorted[3].Matches[0].Done {
		t.Errorf("sorted[3] should be note2 done match, got UUID=%s done=%v", sorted[3].UUID, sorted[3].Matches[0].Done)
	}
}

func TestSortDonePicksLast_AllActive(t *testing.T) {
	results := []models.PickResult{
		{UUID: "a", Matches: []models.PickMatch{{Line: 1, Done: false}}},
	}
	sorted := sortDonePicksLast(results)
	if len(sorted) != 1 || sorted[0].UUID != "a" {
		t.Errorf("all-active should pass through unchanged")
	}
}

func TestSortDonePicksLast_AllDone(t *testing.T) {
	results := []models.PickResult{
		{UUID: "a", Matches: []models.PickMatch{{Line: 1, Done: true}}},
	}
	sorted := sortDonePicksLast(results)
	if len(sorted) != 1 || sorted[0].UUID != "a" {
		t.Errorf("all-done should pass through unchanged")
	}
}
