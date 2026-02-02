//go:build integration

package commands

import (
	"testing"

	"kvnd/lazyruin/pkg/testutil"
)

func TestTagsCommand_List(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Create notes with different tags
	tv.CreateNote("Note with daily tag", "daily")
	tv.CreateNote("Another daily note", "daily")
	tv.CreateNote("Note with work tag", "work")

	ruin := NewRuinCommand(tv.Path)
	tags, err := ruin.Tags.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("List() returned %d tags, want 2", len(tags))
	}

	// Check that daily tag has count of 2
	for _, tag := range tags {
		if tag.Name == "daily" && tag.Count != 2 {
			t.Errorf("daily tag count = %d, want 2", tag.Count)
		}
		if tag.Name == "work" && tag.Count != 1 {
			t.Errorf("work tag count = %d, want 1", tag.Count)
		}
	}
}

func TestTagsCommand_List_Empty(t *testing.T) {
	tv := testutil.NewTestVault(t)

	// Empty vault - no tags
	ruin := NewRuinCommand(tv.Path)
	tags, err := ruin.Tags.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("List() returned %d tags, want 0", len(tags))
	}
}

func TestTagsCommand_Rename(t *testing.T) {
	tv := testutil.NewTestVault(t)

	tv.CreateNote("Note with old tag", "oldtag")

	ruin := NewRuinCommand(tv.Path)

	// Rename the tag
	err := ruin.Tags.Rename("oldtag", "newtag")
	if err != nil {
		t.Fatalf("Rename() error: %v", err)
	}

	// Verify old tag is gone and new tag exists
	tags, err := ruin.Tags.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	for _, tag := range tags {
		if tag.Name == "oldtag" {
			t.Error("oldtag should not exist after rename")
		}
		if tag.Name == "newtag" && tag.Count != 1 {
			t.Errorf("newtag count = %d, want 1", tag.Count)
		}
	}
}
