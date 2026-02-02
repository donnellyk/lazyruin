package commands

import (
	"testing"

	"kvnd/lazyruin/pkg/models"
)

func TestTagsCommand_List_Unit(t *testing.T) {
	mock := NewMockExecutor().WithTags(
		models.Tag{Name: "daily", Count: 5},
		models.Tag{Name: "work", Count: 3},
	)

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	tags, err := ruin.Tags.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("List() returned %d tags, want 2", len(tags))
	}

	// Check specific tag
	found := false
	for _, tag := range tags {
		if tag.Name == "daily" && tag.Count == 5 {
			found = true
		}
	}
	if !found {
		t.Error("expected tag 'daily' with count 5")
	}
}

func TestTagsCommand_List_Empty_Unit(t *testing.T) {
	mock := NewMockExecutor() // No tags

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	tags, err := ruin.Tags.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("List() returned %d tags, want 0", len(tags))
	}
}

func TestTagsCommand_Rename_Unit(t *testing.T) {
	mock := NewMockExecutor()

	ruin := NewRuinCommandWithExecutor(mock, mock.VaultPath())
	err := ruin.Tags.Rename("oldtag", "newtag")
	if err != nil {
		t.Fatalf("Rename() error: %v", err)
	}
	// Success - no error means the command was executed
}
