package commands

import (
	"testing"

	"kvnd/lazyruin/pkg/testutil"
)

func TestSearchCommand_Today(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Create a note (will be dated today)
	tv.CreateNote("Test note for today", "daily", "test")

	ruin := NewRuinCommand(tv.Path)
	notes, err := ruin.Search.Today()
	if err != nil {
		t.Fatalf("Today() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Today() returned %d notes, want 1", len(notes))
	}
}

func TestSearchCommand_Today_Empty(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Empty vault - no notes
	ruin := NewRuinCommand(tv.Path)
	notes, err := ruin.Search.Today()
	if err != nil {
		t.Fatalf("Today() error: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Today() returned %d notes, want 0", len(notes))
	}
}

func TestSearchCommand_ByTag_WithPrefix(t *testing.T) {
	tv := testutil.NewTestVault(t)

	tv.CreateNote("Note with daily tag", "daily")
	tv.CreateNote("Note with work tag", "work")

	ruin := NewRuinCommand(tv.Path)
	notes, err := ruin.Search.ByTag("#daily")
	if err != nil {
		t.Fatalf("ByTag() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("ByTag(#daily) returned %d notes, want 1", len(notes))
	}
}

func TestSearchCommand_ByTag_WithoutPrefix(t *testing.T) {
	tv := testutil.NewTestVault(t)

	tv.CreateNote("Note with daily tag", "daily")
	tv.CreateNote("Note with work tag", "work")

	ruin := NewRuinCommand(tv.Path)
	notes, err := ruin.Search.ByTag("daily")
	if err != nil {
		t.Fatalf("ByTag() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("ByTag(daily) returned %d notes, want 1", len(notes))
	}
}

func TestSearchCommand_Search_WithOptions(t *testing.T) {
	tv := testutil.NewTestVault(t)

	tv.CreateNote("First note", "test")
	tv.CreateNote("Second note", "test")
	tv.CreateNote("Third note", "test")

	ruin := NewRuinCommand(tv.Path)

	// Test with limit
	notes, err := ruin.Search.Search("#test", SearchOptions{Limit: 2})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Search() with limit=2 returned %d notes, want 2", len(notes))
	}
}
