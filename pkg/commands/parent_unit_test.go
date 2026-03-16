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

func TestParentCommand_Compose_UUIDBased_Unit(t *testing.T) {
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
	bm := models.ParentBookmark{Name: "mybookmark", UUID: "root-1", Title: "My Bookmark"}
	note, sourceMap, err := ruin.Parent.Compose(bm)
	if err != nil {
		t.Fatalf("Compose() error: %v", err)
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

func TestParentCommand_Compose_PassesExpandEmbeds(t *testing.T) {
	result := struct {
		UUID            string `json:"uuid"`
		Title           string `json:"title"`
		Path            string `json:"path"`
		ComposedContent string `json:"composed_content"`
	}{
		UUID:            "root-1",
		Title:           "Root",
		Path:            "/vault/root.md",
		ComposedContent: "content",
	}
	data, _ := json.Marshal(result)
	mock := NewMockExecutor().WithCompose(data)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	bm := models.ParentBookmark{Name: "mybookmark", UUID: "root-1", Title: "My Bookmark"}
	_, _, err := ruin.Parent.Compose(bm)
	if err != nil {
		t.Fatalf("Compose() error: %v", err)
	}

	composeCall := findCall(mock.Calls, "compose")
	if composeCall == nil {
		t.Fatal("no compose call recorded")
	}
	if !containsArg(composeCall, "--expand-embeds") {
		t.Errorf("compose call missing --expand-embeds, got: %v", composeCall)
	}
}

func TestParentCommand_Compose_FileBased_Unit(t *testing.T) {
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
		ComposedContent: "composed content here",
		SourceMap: []struct {
			UUID      string `json:"uuid"`
			Path      string `json:"path"`
			Title     string `json:"title"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		}{
			{UUID: "child-a", Path: "/vault/child-a.md", Title: "Child A", StartLine: 1, EndLine: 5},
		},
	}

	data, _ := json.Marshal(result)
	mock := NewMockExecutor().WithCompose(data)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	bm := models.ParentBookmark{Name: "alpha", File: "project.yml", Title: "Alpha"}
	note, sourceMap, err := ruin.Parent.Compose(bm)
	if err != nil {
		t.Fatalf("Compose() error: %v", err)
	}

	if note.Title != "Alpha" {
		t.Errorf("Title = %q, want %q", note.Title, "Alpha")
	}
	if note.UUID != "root-1" {
		t.Errorf("UUID = %q, want %q", note.UUID, "root-1")
	}
	if note.Content != "composed content here" {
		t.Errorf("Content = %q, want %q", note.Content, "composed content here")
	}
	if len(sourceMap) != 1 {
		t.Errorf("SourceMap length = %d, want 1", len(sourceMap))
	}

	composeCall := findCall(mock.Calls, "compose")
	if composeCall == nil {
		t.Fatal("no compose call recorded")
	}
	if !containsArg(composeCall, "--file") {
		t.Errorf("compose call missing --file, got: %v", composeCall)
	}
	if !containsArg(composeCall, "project.yml") {
		t.Errorf("compose call missing file path, got: %v", composeCall)
	}
	if !containsArg(composeCall, "--expand-embeds") {
		t.Errorf("compose call missing --expand-embeds, got: %v", composeCall)
	}
	if containsArg(composeCall, "--sort") {
		t.Errorf("compose call should not include --sort for file-based compose, got: %v", composeCall)
	}
}

func TestParentCommand_List_FileBased_Unit(t *testing.T) {
	mock := NewMockExecutor().WithParents(
		models.ParentBookmark{Name: "docs", UUID: "abc-123", Title: "Documentation Root"},
		models.ParentBookmark{Name: "alpha", File: "project.yml", Title: "Project Alpha Hub"},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	parents, err := ruin.Parent.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("List() returned %d parents, want 2", len(parents))
	}
	if parents[1].IsFileBased() != true {
		t.Error("parents[1] should be file-based")
	}
	if parents[0].IsFileBased() != false {
		t.Error("parents[0] should not be file-based")
	}
	if parents[1].File != "project.yml" {
		t.Errorf("parents[1].File = %q, want %q", parents[1].File, "project.yml")
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

func findCall(calls [][]string, cmd string) []string {
	for _, call := range calls {
		if len(call) > 0 && call[0] == cmd {
			return call
		}
	}
	return nil
}

func containsArg(args []string, target string) bool {
	for _, a := range args {
		if a == target {
			return true
		}
	}
	return false
}
