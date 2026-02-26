package commands

import (
	"fmt"
	"testing"
)

func TestSearchCommand_Error(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("connection refused"))
	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	_, err := ruin.Search.Today()
	if err == nil {
		t.Fatal("expected error from Today()")
	}
	if err.Error() != "connection refused" {
		t.Errorf("error = %q, want %q", err.Error(), "connection refused")
	}
}

func TestSearchCommand_SearchError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("vault locked"))
	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	_, err := ruin.Search.Search("test", SearchOptions{})
	if err == nil {
		t.Fatal("expected error from Search()")
	}
}

func TestTagsCommand_ListError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("timeout"))
	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	_, err := ruin.Tags.List()
	if err == nil {
		t.Fatal("expected error from Tags.List()")
	}
}

func TestQueriesCommand_ListError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("disk full"))
	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	_, err := ruin.Queries.List()
	if err == nil {
		t.Fatal("expected error from Queries.List()")
	}
}

func TestQueriesCommand_RunError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("invalid query"))
	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	_, err := ruin.Queries.Run("bad-query", SearchOptions{})
	if err == nil {
		t.Fatal("expected error from Queries.Run()")
	}
}

func TestParentCommand_ListError(t *testing.T) {
	mock := NewMockExecutor().WithError(fmt.Errorf("not found"))
	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())

	_, err := ruin.Parent.List()
	if err == nil {
		t.Fatal("expected error from Parent.List()")
	}
}

func TestUnmarshalJSON_InvalidJSON(t *testing.T) {
	_, err := unmarshalJSON[[]string]([]byte("not json"))
	if err == nil {
		t.Fatal("expected error from unmarshalJSON with invalid JSON")
	}
}

func TestUnmarshalJSON_EmptyArray(t *testing.T) {
	result, err := unmarshalJSON[[]string]([]byte("[]"))
	if err != nil {
		t.Fatalf("unmarshalJSON error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestUnmarshalJSON_ValidJSON(t *testing.T) {
	type simple struct {
		Name string `json:"name"`
	}
	result, err := unmarshalJSON[simple]([]byte(`{"name":"test"}`))
	if err != nil {
		t.Fatalf("unmarshalJSON error: %v", err)
	}
	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}
}

func TestExecute_UsesInjectedExecutor(t *testing.T) {
	mock := NewMockExecutor()
	ruin := NewRuinCommandWithExecutor(mock, "/test/vault")

	if ruin.VaultPath() != "/test/vault" {
		t.Errorf("VaultPath() = %q, want %q", ruin.VaultPath(), "/test/vault")
	}

	// Verify the executor is used (search calls through Execute)
	notes, err := ruin.Search.Today()
	if err != nil {
		t.Fatalf("Today() error: %v", err)
	}
	// Empty mock returns empty JSON array, which unmarshals to empty (possibly nil) slice
	if len(notes) != 0 {
		t.Errorf("expected 0 notes from empty mock, got %d", len(notes))
	}
}
