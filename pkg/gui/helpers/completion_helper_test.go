package helpers

import (
	"testing"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
	"kvnd/lazyruin/pkg/testutil"

	"github.com/jesseduffield/gocui"
)

// mockGuiCommon implements IGuiCommon with a stubbed ContextTree for tests.
type mockGuiCommon struct {
	contexts *context.ContextTree
}

func (m *mockGuiCommon) Contexts() *context.ContextTree { return m.contexts }

// types.IGuiCommon stubs
func (m *mockGuiCommon) Update(func() error)                          {}
func (m *mockGuiCommon) RenderNotes()                                 {}
func (m *mockGuiCommon) RenderTags()                                  {}
func (m *mockGuiCommon) RenderQueries()                               {}
func (m *mockGuiCommon) RenderPreview()                               {}
func (m *mockGuiCommon) RenderAll()                                   {}
func (m *mockGuiCommon) UpdateNotesTab()                              {}
func (m *mockGuiCommon) UpdateTagsTab()                               {}
func (m *mockGuiCommon) UpdateQueriesTab()                            {}
func (m *mockGuiCommon) UpdateStatusBar()                             {}
func (m *mockGuiCommon) CurrentContext() types.Context                { return nil }
func (m *mockGuiCommon) CurrentContextKey() types.ContextKey          { return "" }
func (m *mockGuiCommon) PushContext(types.Context, types.OnFocusOpts) {}
func (m *mockGuiCommon) PushContextByKey(types.ContextKey)            {}
func (m *mockGuiCommon) PopContext()                                  {}
func (m *mockGuiCommon) ReplaceContext(types.Context)                 {}
func (m *mockGuiCommon) ReplaceContextByKey(types.ContextKey)         {}
func (m *mockGuiCommon) ContextByKey(types.ContextKey) types.Context  { return nil }
func (m *mockGuiCommon) PopupActive() bool                            { return false }
func (m *mockGuiCommon) SearchQueryActive() bool                      { return false }
func (m *mockGuiCommon) ShowConfirm(string, string, func() error)     {}
func (m *mockGuiCommon) ShowInput(string, string, func(string) error) {}
func (m *mockGuiCommon) ShowError(error)                              {}
func (m *mockGuiCommon) ShowMenuDialog(string, []types.MenuItem)      {}
func (m *mockGuiCommon) SetCursorEnabled(bool)                        {}
func (m *mockGuiCommon) Suspend() error                               { return nil }
func (m *mockGuiCommon) Resume() error                                { return nil }
func (m *mockGuiCommon) GetView(string) *gocui.View                   { return nil }
func (m *mockGuiCommon) DeleteView(string)                            {}
func (m *mockGuiCommon) BuildCardContent(models.Note, int) []string   { return nil }
func (m *mockGuiCommon) PreviousContextKey() types.ContextKey         { return "" }

func newTestCompletionHelper(mock *testutil.MockExecutor, gui *mockGuiCommon) *CompletionHelper {
	ruinCmd := commands.NewRuinCommandWithExecutor(mock, "/mock")
	common := NewHelperCommon(ruinCmd, nil, gui)
	return NewCompletionHelper(common)
}

func TestParentCandidates_BookmarkTopLevel(t *testing.T) {
	// > mode, no drill: should return bookmarked parents
	mock := testutil.NewMockExecutor()
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			Queries: context.NewQueriesContext(noop, noop, noop, noop),
			Notes:   context.NewNotesContext(noop, noop),
		},
	}
	gui.contexts.Queries.Parents = []models.ParentBookmark{
		{Name: "alpha", UUID: "uuid-1", Title: "Alpha Note"},
		{Name: "beta", UUID: "uuid-2", Title: "Beta Note"},
	}

	state := types.NewCompletionState()
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates("")
	if len(items) != 2 {
		t.Fatalf("expected 2 bookmarks, got %d", len(items))
	}
	if items[0].Label != "alpha" || items[0].Value != "uuid-1" {
		t.Errorf("first item = %+v, want alpha/uuid-1", items[0])
	}
}

func TestParentCandidates_BookmarkDrilled(t *testing.T) {
	// > mode, drilled into alpha: should return children of uuid-1, not bookmarks
	child := models.Note{UUID: "child-1", Title: "Child One", Parent: "uuid-1"}
	unrelated := models.Note{UUID: "other-1", Title: "Other", Parent: "uuid-99"}

	mock := testutil.NewMockExecutor().WithNotes(child, unrelated)
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			Queries: context.NewQueriesContext(noop, noop, noop, noop),
			Notes:   context.NewNotesContext(noop, noop),
		},
	}
	gui.contexts.Queries.Parents = []models.ParentBookmark{
		{Name: "alpha", UUID: "uuid-1", Title: "Alpha Note"},
	}

	state := types.NewCompletionState()
	state.ParentDrill = []types.ParentDrillEntry{{Name: "alpha", UUID: "uuid-1"}}
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates("alpha/")
	if len(items) != 1 {
		t.Fatalf("expected 1 child, got %d", len(items))
	}
	if items[0].Label != "Child One" || items[0].Value != "child-1" {
		t.Errorf("item = %+v, want Child One/child-1", items[0])
	}
}

func TestParentCandidates_AllNotesTopLevel(t *testing.T) {
	// >> mode, no drill: should return all notes
	mock := testutil.NewMockExecutor()
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			Queries: context.NewQueriesContext(noop, noop, noop, noop),
			Notes:   context.NewNotesContext(noop, noop),
		},
	}
	gui.contexts.Notes.Items = []models.Note{
		{UUID: "n1", Title: "Note A"},
		{UUID: "n2", Title: "Note B"},
	}

	state := types.NewCompletionState()
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates(">")
	if len(items) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(items))
	}
}

func TestParentCandidates_AllNotesDrilled(t *testing.T) {
	// >> mode, drilled into alpha: should return children of uuid-1, NOT all notes
	child := models.Note{UUID: "child-1", Title: "Child One", Parent: "uuid-1"}
	other := models.Note{UUID: "n2", Title: "Other Note", Parent: "uuid-99"}

	mock := testutil.NewMockExecutor().WithNotes(child, other)
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			Queries: context.NewQueriesContext(noop, noop, noop, noop),
			Notes:   context.NewNotesContext(noop, noop),
		},
	}
	gui.contexts.Notes.Items = []models.Note{
		{UUID: "n1", Title: "Note A"},
		{UUID: "n2", Title: "Note B"},
	}

	state := types.NewCompletionState()
	state.ParentDrill = []types.ParentDrillEntry{{Name: "alpha", UUID: "uuid-1"}}
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates(">alpha/")
	if len(items) != 1 {
		t.Fatalf("expected 1 child, got %d", len(items))
	}
	if items[0].Label != "Child One" {
		t.Errorf("item label = %q, want Child One", items[0].Label)
	}
}

func TestParentCandidates_DrillStackSyncOnBackspace(t *testing.T) {
	// User backspaced past the slash: filter has no slashes but drill stack has an entry.
	// Should truncate drill stack and return bookmarks.
	mock := testutil.NewMockExecutor()
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			Queries: context.NewQueriesContext(noop, noop, noop, noop),
			Notes:   context.NewNotesContext(noop, noop),
		},
	}
	gui.contexts.Queries.Parents = []models.ParentBookmark{
		{Name: "alpha", UUID: "uuid-1", Title: "Alpha Note"},
	}

	state := types.NewCompletionState()
	state.ParentDrill = []types.ParentDrillEntry{{Name: "alpha", UUID: "uuid-1"}}
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates("")
	// Drill stack should have been truncated to empty
	if len(state.ParentDrill) != 0 {
		t.Errorf("ParentDrill len = %d, want 0", len(state.ParentDrill))
	}
	// Should return bookmarks, not children
	if len(items) != 1 || items[0].Label != "alpha" {
		t.Errorf("expected bookmark [alpha], got %+v", items)
	}
}

func noop() {}
