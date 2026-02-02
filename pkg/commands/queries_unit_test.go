package commands

import (
	"testing"

	"kvnd/lazyruin/pkg/models"
)

func TestQueriesCommand_List_Unit(t *testing.T) {
	mock := NewMockExecutor().WithQueries(
		models.Query{Name: "daily-notes", Query: "#daily"},
		models.Query{Name: "work-items", Query: "#work"},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	queries, err := ruin.Queries.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(queries) != 2 {
		t.Errorf("List() returned %d queries, want 2", len(queries))
	}
}

func TestQueriesCommand_List_Empty_Unit(t *testing.T) {
	mock := NewMockExecutor() // No queries

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	queries, err := ruin.Queries.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(queries) != 0 {
		t.Errorf("List() returned %d queries, want 0", len(queries))
	}
}

func TestQueriesCommand_Run_Unit(t *testing.T) {
	mock := NewMockExecutor().WithNotes(
		models.Note{UUID: "1", Title: "Matching note", Tags: []string{"daily"}},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	notes, err := ruin.Queries.Run("daily-notes")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Run() returned %d notes, want 1", len(notes))
	}
}

func TestQueriesCommand_Save_Unit(t *testing.T) {
	mock := NewMockExecutor()

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	err := ruin.Queries.Save("test-query", "#test")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	// Success - no error means the command was executed
}
