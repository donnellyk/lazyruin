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
		UUID            string `json:"uuid"`
		Title           string `json:"title"`
		Path            string `json:"path"`
		ComposedContent string `json:"composed_content"`
		SourceMap       []struct {
			UUID      string `json:"uuid"`
			Path      string `json:"path"`
			Title     string `json:"title"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		} `json:"source_map"`
	}{
		UUID:            "root-1",
		Title:           "Root Note",
		Path:            "/vault/root.md",
		ComposedContent: "Root content\n\n## Child A\n\nChild A content\n\n## Child B\n\nChild B content",
		SourceMap: []struct {
			UUID      string `json:"uuid"`
			Path      string `json:"path"`
			Title     string `json:"title"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		}{
			{UUID: "child-a", Path: "/vault/child-a.md", Title: "Child A", StartLine: 1, EndLine: 5},
			{UUID: "child-b", Path: "/vault/child-b.md", Title: "Child B", StartLine: 7, EndLine: 9},
		},
	}

	data, _ := json.Marshal(result)
	mock := NewMockExecutor().WithCompose(data)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	note, sourceMap, err := ruin.Parent.ComposeFlat("root-1", "My Bookmark")
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
	if len(sourceMap) != 2 {
		t.Errorf("SourceMap length = %d, want 2", len(sourceMap))
	}
	if len(sourceMap) > 0 && sourceMap[0].UUID != "child-a" {
		t.Errorf("SourceMap[0].UUID = %q, want %q", sourceMap[0].UUID, "child-a")
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
