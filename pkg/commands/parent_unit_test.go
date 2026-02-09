package commands

import (
	"encoding/json"
	"testing"

	"kvnd/lazyruin/pkg/models"
)

func TestParentCommand_List_Unit(t *testing.T) {
	mock := NewMockExecutor().WithParents(
		models.ParentBookmark{Name: "project", UUID: "abc-123", Title: "Project Notes"},
		models.ParentBookmark{Name: "journal", UUID: "def-456", Title: "Daily Journal"},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	parents, err := ruin.Parent.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("List() returned %d parents, want 2", len(parents))
	}
	if parents[0].Name != "project" {
		t.Errorf("parents[0].Name = %q, want %q", parents[0].Name, "project")
	}
}

func TestParentCommand_List_Empty_Unit(t *testing.T) {
	mock := NewMockExecutor()

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	parents, err := ruin.Parent.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(parents) != 0 {
		t.Errorf("List() returned %d parents, want 0", len(parents))
	}
}

func TestParentCommand_ComposeFlat_Unit(t *testing.T) {
	result := struct {
		UUID    string `json:"uuid"`
		Title   string `json:"title"`
		Path    string `json:"path"`
		Content string `json:"content"`
	}{
		UUID:    "root-1",
		Title:   "Root Note",
		Path:    "/vault/root.md",
		Content: "Root content\n\n## Child A\n\nChild A content\n\n## Child B\n\nChild B content",
	}

	data, _ := json.Marshal(result)
	mock := NewMockExecutor().WithCompose(data)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	note, err := ruin.Parent.ComposeFlat("root-1", "My Bookmark")
	if err != nil {
		t.Fatalf("ComposeFlat() error: %v", err)
	}

	if note.Title != "My Bookmark" {
		t.Errorf("Title = %q, want %q", note.Title, "My Bookmark")
	}
	if note.UUID != "root-1" {
		t.Errorf("UUID = %q, want %q", note.UUID, "root-1")
	}
	if note.Content == "" {
		t.Error("Content is empty")
	}
}

func TestParentCommand_Delete_Unit(t *testing.T) {
	mock := NewMockExecutor()

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	err := ruin.Parent.Delete("project")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
}
