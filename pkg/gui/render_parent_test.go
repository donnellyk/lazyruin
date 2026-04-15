package gui

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

// TestResolveParentLabel_UsesTitleCache verifies that when a note's parent
// UUID is neither a loaded bookmark nor currently in Notes.Items,
// resolveParentLabel consults the title cache before falling back to a
// truncated UUID.
//
// Regression: notes whose parent is an older note (not in the current tab)
// rendered as a truncated UUID instead of the parent's title.
func TestResolveParentLabel_UsesTitleCache(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Drop both fast-path sources so only the cache can resolve the label.
	tg.gui.contexts.Queries.Parents = nil
	tg.gui.contexts.Notes.Items = nil

	tg.gui.helpers.TitleCache().Put("offscreen-parent", "Old Parent Note")

	got := tg.gui.resolveParentLabel("offscreen-parent")
	if got != "Old Parent Note" {
		t.Errorf("resolveParentLabel = %q, want %q", got, "Old Parent Note")
	}
}

// TestResolveParentLabel_EagerFetchOnNotesLoad verifies that when the Notes
// panel loads, parent UUIDs referenced by the loaded notes but not otherwise
// known are fetched via `ruin get --uuid` and cached, so subsequent renders
// resolve parent labels to titles rather than truncated UUIDs.
func TestResolveParentLabel_EagerFetchOnNotesLoad(t *testing.T) {
	parentNote := models.Note{UUID: "hidden-parent", Title: "Hidden Parent Note"}
	childNote := models.Note{UUID: "child-1", Title: "Child One", Parent: "hidden-parent"}

	mock := testutil.NewMockExecutor().WithNotes(parentNote, childNote)
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.helpers.Notes().FetchNotesForCurrentTab(false)

	// Simulate a tab where only the child is visible in the panel.
	tg.gui.contexts.Notes.Items = []models.Note{childNote}
	tg.gui.contexts.Queries.Parents = nil

	got := tg.gui.resolveParentLabel("hidden-parent")
	if got != "Hidden Parent Note" {
		t.Errorf("resolveParentLabel = %q, want %q", got, "Hidden Parent Note")
	}
}
