//go:build integration

package commands

import (
	"testing"

	"kvnd/lazyruin/pkg/testutil"
)

func TestQueriesCommand_List(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Create some saved queries
	tv.SaveQuery("daily-notes", "#daily")
	tv.SaveQuery("work-items", "#work")

	ruin := NewRuinCommand(tv.Path)
	queries, err := ruin.Queries.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(queries) != 2 {
		t.Errorf("List() returned %d queries, want 2", len(queries))
	}
}

func TestQueriesCommand_List_Empty(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Empty vault - no saved queries
	ruin := NewRuinCommand(tv.Path)
	queries, err := ruin.Queries.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(queries) != 0 {
		t.Errorf("List() returned %d queries, want 0", len(queries))
	}
}

func TestQueriesCommand_Run(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Create notes and a query
	tv.CreateNote("Daily standup notes", "daily")
	tv.CreateNote("Weekly review", "weekly")
	tv.SaveQuery("daily-notes", "#daily")

	ruin := NewRuinCommand(tv.Path)
	notes, err := ruin.Queries.Run("daily-notes")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Run() returned %d notes, want 1", len(notes))
	}
}

func TestQueriesCommand_Save(t *testing.T) {
	tv := testutil.NewTestVault(t)

	ruin := NewRuinCommand(tv.Path)

	// Save a new query
	err := ruin.Queries.Save("test-query", "#test")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify it was saved
	queries, err := ruin.Queries.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	found := false
	for _, q := range queries {
		if q.Name == "test-query" {
			found = true
			if q.Query != "#test" {
				t.Errorf("query string = %q, want %q", q.Query, "#test")
			}
		}
	}

	if !found {
		t.Error("saved query not found in list")
	}
}
