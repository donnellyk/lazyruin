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

func TestParentCommand_Compose_Unit(t *testing.T) {
	tree := struct {
		UUID     string `json:"uuid"`
		Title    string `json:"title"`
		Path     string `json:"path"`
		Content  string `json:"content"`
		Children []struct {
			UUID     string        `json:"uuid"`
			Title    string        `json:"title"`
			Path     string        `json:"path"`
			Content  string        `json:"content"`
			Children []interface{} `json:"children"`
		} `json:"children"`
	}{
		UUID:    "root-1",
		Title:   "Root Note",
		Path:    "/vault/root.md",
		Content: "Root content",
		Children: []struct {
			UUID     string        `json:"uuid"`
			Title    string        `json:"title"`
			Path     string        `json:"path"`
			Content  string        `json:"content"`
			Children []interface{} `json:"children"`
		}{
			{UUID: "child-1", Title: "Child A", Path: "/vault/a.md", Content: "Child A content", Children: nil},
			{UUID: "child-2", Title: "Child B", Path: "/vault/b.md", Content: "Child B content", Children: nil},
		},
	}

	data, _ := json.Marshal(tree)
	mock := NewMockExecutor().WithCompose(data)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	notes, err := ruin.Parent.Compose("root-1")
	if err != nil {
		t.Fatalf("Compose() error: %v", err)
	}

	if len(notes) != 3 {
		t.Fatalf("Compose() returned %d notes, want 3", len(notes))
	}

	if notes[0].Title != "Root Note" {
		t.Errorf("notes[0].Title = %q, want %q", notes[0].Title, "Root Note")
	}
	if notes[1].Title != "Child A" {
		t.Errorf("notes[1].Title = %q, want %q", notes[1].Title, "Child A")
	}
	if notes[2].Title != "Child B" {
		t.Errorf("notes[2].Title = %q, want %q", notes[2].Title, "Child B")
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
