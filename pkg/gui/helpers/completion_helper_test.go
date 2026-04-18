package helpers

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"

	"github.com/jesseduffield/gocui"
)

// mockGuiCommon implements IGuiCommon with a stubbed ContextTree for tests.
type mockGuiCommon struct {
	contexts *context.ContextTree
}

func (m *mockGuiCommon) Contexts() *context.ContextTree { return m.contexts }

// types.IGuiCommon stubs
func (m *mockGuiCommon) Update(func() error)                                  {}
func (m *mockGuiCommon) RenderNotes()                                         {}
func (m *mockGuiCommon) RenderTags()                                          {}
func (m *mockGuiCommon) RenderQueries()                                       {}
func (m *mockGuiCommon) RenderPreview()                                       {}
func (m *mockGuiCommon) RenderAll()                                           {}
func (m *mockGuiCommon) UpdateNotesTab()                                      {}
func (m *mockGuiCommon) UpdateTagsTab()                                       {}
func (m *mockGuiCommon) UpdateQueriesTab()                                    {}
func (m *mockGuiCommon) UpdateStatusBar()                                     {}
func (m *mockGuiCommon) CurrentContext() types.Context                        { return nil }
func (m *mockGuiCommon) CurrentContextKey() types.ContextKey                  { return "" }
func (m *mockGuiCommon) PushContext(types.Context, types.OnFocusOpts)         {}
func (m *mockGuiCommon) PushContextByKey(types.ContextKey)                    {}
func (m *mockGuiCommon) PopContext()                                          {}
func (m *mockGuiCommon) ReplaceContext(types.Context)                         {}
func (m *mockGuiCommon) ReplaceContextByKey(types.ContextKey)                 {}
func (m *mockGuiCommon) ContextByKey(types.ContextKey) types.Context          { return nil }
func (m *mockGuiCommon) PopupActive() bool                                    { return false }
func (m *mockGuiCommon) SearchQueryActive() bool                              { return false }
func (m *mockGuiCommon) ShowConfirm(string, string, func() error)             {}
func (m *mockGuiCommon) ShowInput(string, string, func(string) error)         {}
func (m *mockGuiCommon) ShowError(error)                                      {}
func (m *mockGuiCommon) ShowMenuDialog(string, []types.MenuItem)              {}
func (m *mockGuiCommon) ShowAbout()                                           {}
func (m *mockGuiCommon) SetCursorEnabled(bool)                                {}
func (m *mockGuiCommon) Suspend() error                                       { return nil }
func (m *mockGuiCommon) Resume() error                                        { return nil }
func (m *mockGuiCommon) GetView(string) *gocui.View                           { return nil }
func (m *mockGuiCommon) DeleteView(string)                                    {}
func (m *mockGuiCommon) BuildCardContent(models.Note, int) []types.SourceLine { return nil }
func (m *mockGuiCommon) RenderPickDialog()                                    {}
func (m *mockGuiCommon) PreviousContextKey() types.ContextKey                 { return "" }

func newTestCompletionHelper(mock *testutil.MockExecutor, gui *mockGuiCommon) *CompletionHelper {
	ruinCmd := commands.NewRuinCommandWithExecutor(mock, "/mock")
	common := NewHelperCommon(ruinCmd, nil, gui)
	return NewCompletionHelper(common)
}

func TestParentCandidates_BookmarkTopLevel(t *testing.T) {
	// > mode, no drill: should return Bookmarks section with bookmarked parents
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
	// expect: Bookmarks header + alpha + beta
	if len(items) != 3 {
		t.Fatalf("expected 3 items (1 header + 2 bookmarks), got %d", len(items))
	}
	if !items[0].IsHeader || items[0].Label != "Bookmarks" {
		t.Errorf("first item = %+v, want Bookmarks header", items[0])
	}
	if items[1].Label != "alpha" || items[1].Value != "uuid-1" {
		t.Errorf("second item = %+v, want alpha/uuid-1", items[1])
	}
}

func TestParentCandidates_TopLevelShowsBookmarksAndNotes(t *testing.T) {
	// > mode, no drill: should show both Bookmarks and Notes sections
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
	gui.contexts.Notes.Items = []models.Note{
		{UUID: "uuid-1", Title: "Alpha Note"}, // bookmarked — should be skipped in Notes
		{UUID: "n2", Title: "Note B"},
	}

	state := types.NewCompletionState()
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates("")
	// expect: Bookmarks header + alpha + Notes header + Note B
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d: %+v", len(items), items)
	}
	if !items[0].IsHeader || items[0].Label != "Bookmarks" {
		t.Errorf("items[0] = %+v, want Bookmarks header", items[0])
	}
	if items[1].Label != "alpha" {
		t.Errorf("items[1] = %+v, want alpha", items[1])
	}
	if !items[2].IsHeader || items[2].Label != "Notes" {
		t.Errorf("items[2] = %+v, want Notes header", items[2])
	}
	if items[3].Label != "Note B" {
		t.Errorf("items[3] = %+v, want Note B", items[3])
	}
}

func TestParentCandidates_HidesEmptyBookmarksSection(t *testing.T) {
	// No bookmarks, only notes: should omit the Bookmarks header
	mock := testutil.NewMockExecutor()
	gui := &mockGuiCommon{
		contexts: &context.ContextTree{
			Queries: context.NewQueriesContext(noop, noop, noop, noop),
			Notes:   context.NewNotesContext(noop, noop),
		},
	}
	gui.contexts.Notes.Items = []models.Note{
		{UUID: "n1", Title: "Note A"},
	}

	state := types.NewCompletionState()
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates("")
	if len(items) != 2 {
		t.Fatalf("expected 2 items (Notes header + Note A), got %d: %+v", len(items), items)
	}
	if !items[0].IsHeader || items[0].Label != "Notes" {
		t.Errorf("items[0] = %+v, want Notes header", items[0])
	}
	if items[1].Label != "Note A" {
		t.Errorf("items[1] = %+v, want Note A", items[1])
	}
}

func TestParentCandidates_HidesEmptyNotesSection(t *testing.T) {
	// Only bookmarks, no notes: should omit the Notes header
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
	candidates := newTestCompletionHelper(mock, gui).ParentCandidatesFor(state)

	items := candidates("")
	if len(items) != 2 {
		t.Fatalf("expected 2 items (Bookmarks header + alpha), got %d: %+v", len(items), items)
	}
	if !items[0].IsHeader || items[0].Label != "Bookmarks" {
		t.Errorf("items[0] = %+v, want Bookmarks header", items[0])
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

func TestParentCandidates_DrillStackSyncOnBackspace(t *testing.T) {
	// User backspaced past the slash: filter has no slashes but drill stack has an entry.
	// Should truncate drill stack and return top-level sections.
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
	// Should return Bookmarks section (header + alpha), not children
	if len(items) != 2 || items[1].Label != "alpha" {
		t.Errorf("expected [Bookmarks header, alpha], got %+v", items)
	}
}

func noop() {}
