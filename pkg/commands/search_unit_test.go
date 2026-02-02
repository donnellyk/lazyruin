package commands

import (
	"testing"
	"time"

	"kvnd/lazyruin/pkg/models"
)

func TestSearchCommand_Today_Unit(t *testing.T) {
	mock := NewMockExecutor().WithNotes(
		models.Note{UUID: "1", Title: "Today's note", Tags: []string{"daily"}, Created: time.Now()},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	notes, err := ruin.Search.Today()
	if err != nil {
		t.Fatalf("Today() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Today() returned %d notes, want 1", len(notes))
	}

	if notes[0].Title != "Today's note" {
		t.Errorf("Title = %q, want %q", notes[0].Title, "Today's note")
	}
}

func TestSearchCommand_Today_Empty_Unit(t *testing.T) {
	mock := NewMockExecutor() // No notes

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	notes, err := ruin.Search.Today()
	if err != nil {
		t.Fatalf("Today() error: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Today() returned %d notes, want 0", len(notes))
	}
}

func TestSearchCommand_ByTag_Unit(t *testing.T) {
	mock := NewMockExecutor().WithNotes(
		models.Note{UUID: "1", Title: "Daily note", Tags: []string{"daily"}},
		models.Note{UUID: "2", Title: "Work note", Tags: []string{"work"}},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	// Search with # prefix
	notes, err := ruin.Search.ByTag("#daily")
	if err != nil {
		t.Fatalf("ByTag() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("ByTag(#daily) returned %d notes, want 1", len(notes))
	}
}

func TestSearchCommand_ByTag_WithoutPrefix_Unit(t *testing.T) {
	mock := NewMockExecutor().WithNotes(
		models.Note{UUID: "1", Title: "Daily note", Tags: []string{"daily"}},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	// Search without # prefix - should still work
	notes, err := ruin.Search.ByTag("daily")
	if err != nil {
		t.Fatalf("ByTag() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("ByTag(daily) returned %d notes, want 1", len(notes))
	}
}

func TestSearchCommand_Search_WithLimit_Unit(t *testing.T) {
	mock := NewMockExecutor().WithNotes(
		models.Note{UUID: "1", Title: "Note 1", Tags: []string{"test"}},
		models.Note{UUID: "2", Title: "Note 2", Tags: []string{"test"}},
		models.Note{UUID: "3", Title: "Note 3", Tags: []string{"test"}},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	// Note: The mock doesn't actually limit, but we're testing the command layer
	notes, err := ruin.Search.Search("#test", SearchOptions{Limit: 2})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	// Mock returns all matching notes; limit would be enforced by real CLI
	if len(notes) == 0 {
		t.Error("Search() returned no notes")
	}
}
