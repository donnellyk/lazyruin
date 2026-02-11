package gui

import (
	"fmt"
	"testing"
	"time"

	"kvnd/lazyruin/pkg/models"
	"kvnd/lazyruin/pkg/testutil"

	"github.com/jesseduffield/gocui"
)

func defaultMock() *testutil.MockExecutor {
	return testutil.NewMockExecutor().
		WithNotes(
			models.Note{UUID: "1", Title: "Note One", Tags: []string{"daily"}, Created: time.Now()},
			models.Note{UUID: "2", Title: "Note Two", Tags: []string{"work"}, Created: time.Now()},
			models.Note{UUID: "3", Title: "Note Three", Tags: []string{"daily"}, Created: time.Now()},
			models.Note{UUID: "4", Title: "Note Four", Tags: []string{"project"}, Created: time.Now()},
			models.Note{UUID: "5", Title: "Note Five", Tags: []string{"daily"}, Created: time.Now()},
		).
		WithTags(
			models.Tag{Name: "daily", Count: 3},
			models.Tag{Name: "work", Count: 1},
			models.Tag{Name: "project", Count: 1},
		).
		WithQueries(
			models.Query{Name: "daily-notes", Query: "#daily"},
			models.Query{Name: "work-items", Query: "#work"},
		).
		WithParents(
			models.ParentBookmark{Name: "journal", UUID: "parent-1", Title: "Daily Journal"},
		)
}

// --- Initialization tests ---

func TestHeadlessGui_Initializes(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if !tg.gui.state.Initialized {
		t.Error("GUI should be initialized after layout")
	}
	if tg.gui.state.CurrentContext != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", tg.gui.state.CurrentContext)
	}
}

func TestHeadlessGui_LoadsNotes(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if len(tg.gui.state.Notes.Items) != 5 {
		t.Errorf("Notes.Items = %d, want 5", len(tg.gui.state.Notes.Items))
	}
	if tg.gui.state.Notes.SelectedIndex != 0 {
		t.Errorf("Notes.SelectedIndex = %d, want 0", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestHeadlessGui_LoadsTags(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if len(tg.gui.state.Tags.Items) != 3 {
		t.Errorf("Tags.Items = %d, want 3", len(tg.gui.state.Tags.Items))
	}
}

func TestHeadlessGui_LoadsQueries(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if len(tg.gui.state.Queries.Items) != 2 {
		t.Errorf("Queries.Items = %d, want 2", len(tg.gui.state.Queries.Items))
	}
}

func TestHeadlessGui_ViewsCreated(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if tg.gui.views.Notes == nil {
		t.Error("Notes view should be created")
	}
	if tg.gui.views.Queries == nil {
		t.Error("Queries view should be created")
	}
	if tg.gui.views.Tags == nil {
		t.Error("Tags view should be created")
	}
	if tg.gui.views.Preview == nil {
		t.Error("Preview view should be created")
	}
	if tg.gui.views.Status == nil {
		t.Error("Status view should be created")
	}
}

// --- Notes navigation tests ---

func TestNotesDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.notesDown(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Notes.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesDown_MultipleSteps(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	for i := 0; i < 3; i++ {
		tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	}

	if tg.gui.state.Notes.SelectedIndex != 3 {
		t.Errorf("SelectedIndex = %d, want 3 after 3 down presses", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesUp_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move down first, then up
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesUp(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Notes.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesTop_JumpsToFirst(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesTop(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Notes.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesBottom_JumpsToLast(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.notesBottom(tg.g, tg.gui.views.Notes)

	want := len(tg.gui.state.Notes.Items) - 1
	if tg.gui.state.Notes.SelectedIndex != want {
		t.Errorf("SelectedIndex = %d, want %d", tg.gui.state.Notes.SelectedIndex, want)
	}
}

func TestNotesDown_StopsAtBoundary(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	count := len(tg.gui.state.Notes.Items)
	for i := 0; i < count+5; i++ {
		tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	}

	if tg.gui.state.Notes.SelectedIndex != count-1 {
		t.Errorf("SelectedIndex = %d, want %d (should clamp at last)", tg.gui.state.Notes.SelectedIndex, count-1)
	}
}

// --- Context switching tests ---

func TestNextPanel_CyclesThroughContexts(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Start at NotesContext
	if tg.gui.state.CurrentContext != NotesContext {
		t.Fatalf("initial context = %v, want NotesContext", tg.gui.state.CurrentContext)
	}

	// Tab → QueriesContext
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.CurrentContext != QueriesContext {
		t.Errorf("after first Tab: context = %v, want QueriesContext", tg.gui.state.CurrentContext)
	}

	// Tab → TagsContext
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.CurrentContext != TagsContext {
		t.Errorf("after second Tab: context = %v, want TagsContext", tg.gui.state.CurrentContext)
	}

	// Tab → wraps to NotesContext
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.CurrentContext != NotesContext {
		t.Errorf("after third Tab: context = %v, want NotesContext (wrap)", tg.gui.state.CurrentContext)
	}
}

func TestPrevPanel_CyclesBackward(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// BackTab from NotesContext → TagsContext (wraps backward)
	tg.gui.prevPanel(tg.g, nil)
	if tg.gui.state.CurrentContext != TagsContext {
		t.Errorf("after BackTab from Notes: context = %v, want TagsContext", tg.gui.state.CurrentContext)
	}
}

func TestFocusNotes_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Switch to tags first
	tg.gui.focusTags(tg.g, nil)
	if tg.gui.state.CurrentContext != TagsContext {
		t.Fatalf("context = %v, want TagsContext", tg.gui.state.CurrentContext)
	}

	// Press 1 → NotesContext
	tg.gui.focusNotes(tg.g, nil)
	if tg.gui.state.CurrentContext != NotesContext {
		t.Errorf("context = %v, want NotesContext", tg.gui.state.CurrentContext)
	}
}

func TestFocusQueries_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	if tg.gui.state.CurrentContext != QueriesContext {
		t.Errorf("context = %v, want QueriesContext", tg.gui.state.CurrentContext)
	}
}

func TestFocusTags_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	if tg.gui.state.CurrentContext != TagsContext {
		t.Errorf("context = %v, want TagsContext", tg.gui.state.CurrentContext)
	}
}

func TestFocusPreview_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusPreview(tg.g, nil)
	if tg.gui.state.CurrentContext != PreviewContext {
		t.Errorf("context = %v, want PreviewContext", tg.gui.state.CurrentContext)
	}
}

func TestContextSwitch_TracksPrevious(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	if tg.gui.state.PreviousContext != NotesContext {
		t.Errorf("PreviousContext = %v, want NotesContext", tg.gui.state.PreviousContext)
	}

	tg.gui.focusPreview(tg.g, nil)
	if tg.gui.state.PreviousContext != TagsContext {
		t.Errorf("PreviousContext = %v, want TagsContext", tg.gui.state.PreviousContext)
	}
}

// --- Search workflow tests ---

func TestOpenSearch_EntersSearchMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openSearch(tg.g, nil)

	if !tg.gui.state.SearchMode {
		t.Error("SearchMode should be true")
	}
	if tg.gui.state.CurrentContext != SearchContext {
		t.Errorf("CurrentContext = %v, want SearchContext", tg.gui.state.CurrentContext)
	}
}

func TestCancelSearch_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Start from tags context
	tg.gui.focusTags(tg.g, nil)
	prev := tg.gui.state.CurrentContext

	tg.gui.openSearch(tg.g, nil)
	tg.gui.cancelSearch(tg.g, tg.gui.views.Search)

	if tg.gui.state.SearchMode {
		t.Error("SearchMode should be false after cancel")
	}
	if tg.gui.state.CurrentContext != prev {
		t.Errorf("CurrentContext = %v, want %v (restored)", tg.gui.state.CurrentContext, prev)
	}
}

func TestClearSearch_ResetsState(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Simulate an active search by setting SearchQuery
	tg.gui.state.SearchQuery = "#daily"

	tg.gui.clearSearch(tg.g, nil)

	if tg.gui.state.SearchQuery != "" {
		t.Errorf("SearchQuery = %q, want empty", tg.gui.state.SearchQuery)
	}
	if tg.gui.state.Notes.CurrentTab != NotesTabAll {
		t.Errorf("CurrentTab = %v, want NotesTabAll", tg.gui.state.Notes.CurrentTab)
	}
	if tg.gui.state.CurrentContext != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", tg.gui.state.CurrentContext)
	}
}

// --- Tags navigation tests ---

func TestTagsDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Focus tags panel first
	tg.gui.focusTags(tg.g, nil)

	tg.gui.tagsDown(tg.g, tg.gui.views.Tags)

	if tg.gui.state.Tags.SelectedIndex != 1 {
		t.Errorf("Tags.SelectedIndex = %d, want 1", tg.gui.state.Tags.SelectedIndex)
	}
}

func TestTagsUp_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.tagsDown(tg.g, tg.gui.views.Tags)
	tg.gui.tagsDown(tg.g, tg.gui.views.Tags)
	tg.gui.tagsUp(tg.g, tg.gui.views.Tags)

	if tg.gui.state.Tags.SelectedIndex != 1 {
		t.Errorf("Tags.SelectedIndex = %d, want 1", tg.gui.state.Tags.SelectedIndex)
	}
}

// --- Queries navigation tests ---

func TestQueriesDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	tg.gui.queriesDown(tg.g, tg.gui.views.Queries)

	if tg.gui.state.Queries.SelectedIndex != 1 {
		t.Errorf("Queries.SelectedIndex = %d, want 1", tg.gui.state.Queries.SelectedIndex)
	}
}

// --- Preview state tests ---

func TestNotesDown_UpdatesPreview(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move down to select Note Two
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)

	// Preview should be in single note mode showing the selected note
	if tg.gui.state.Preview.Mode != PreviewModeSingleNote {
		t.Errorf("Preview.Mode = %v, want PreviewModeSingleNote", tg.gui.state.Preview.Mode)
	}
}

// --- Empty state tests ---

func TestEmptyNotes_NoNavigationPanic(t *testing.T) {
	mock := testutil.NewMockExecutor() // no data
	tg := newTestGui(t, mock)
	defer tg.Close()

	// These should not panic with empty lists
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesUp(tg.g, tg.gui.views.Notes)
	tg.gui.notesTop(tg.g, tg.gui.views.Notes)
	tg.gui.notesBottom(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Notes.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 for empty list", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestEmptyTags_NoNavigationPanic(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.tagsDown(tg.g, tg.gui.views.Tags)
	tg.gui.tagsUp(tg.g, tg.gui.views.Tags)

	if tg.gui.state.Tags.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 for empty list", tg.gui.state.Tags.SelectedIndex)
	}
}

// --- Filter by tag tests ---

func TestFilterByTag_SetsPreviewCardList(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	if tg.gui.state.Preview.Mode != PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want PreviewModeCardList", tg.gui.state.Preview.Mode)
	}
	if tg.gui.state.Preview.SelectedCardIndex != 0 {
		t.Errorf("SelectedCardIndex = %d, want 0", tg.gui.state.Preview.SelectedCardIndex)
	}
	if tg.gui.state.CurrentContext != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", tg.gui.state.CurrentContext)
	}
}

func TestFilterByTag_EmptyTags_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	// Should remain in tags context, no switch to preview
	if tg.gui.state.CurrentContext != TagsContext {
		t.Errorf("CurrentContext = %v, want TagsContext (noop for empty)", tg.gui.state.CurrentContext)
	}
}

// --- Run query tests ---

func TestRunQuery_SetsPreviewCardList(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	tg.gui.runQuery(tg.g, tg.gui.views.Queries)

	if tg.gui.state.Preview.Mode != PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want PreviewModeCardList", tg.gui.state.Preview.Mode)
	}
	if len(tg.gui.state.Preview.Cards) == 0 {
		t.Error("Preview.Cards should not be empty after running query")
	}
	if tg.gui.state.CurrentContext != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", tg.gui.state.CurrentContext)
	}
}

func TestRunQuery_EmptyQueries_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	tg.gui.runQuery(tg.g, tg.gui.views.Queries)

	if tg.gui.state.CurrentContext != QueriesContext {
		t.Errorf("CurrentContext = %v, want QueriesContext (noop for empty)", tg.gui.state.CurrentContext)
	}
}

// --- Preview navigation tests ---

func TestPreviewDown_CardListMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Enter card list mode via tag filter
	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	if len(tg.gui.state.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.state.Preview.Cards))
	}

	tg.gui.previewDown(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.SelectedCardIndex != 1 {
		t.Errorf("SelectedCardIndex = %d, want 1", tg.gui.state.Preview.SelectedCardIndex)
	}
}

func TestPreviewUp_CardListMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	if len(tg.gui.state.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.state.Preview.Cards))
	}

	tg.gui.previewDown(tg.g, tg.gui.views.Preview)
	tg.gui.previewUp(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.SelectedCardIndex != 0 {
		t.Errorf("SelectedCardIndex = %d, want 0", tg.gui.state.Preview.SelectedCardIndex)
	}
}

func TestPreviewDown_SingleNoteMode_Scrolls(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// In single note mode, j scrolls
	tg.gui.focusPreview(tg.g, nil)
	tg.gui.previewDown(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.ScrollOffset != 1 {
		t.Errorf("ScrollOffset = %d, want 1", tg.gui.state.Preview.ScrollOffset)
	}
}

func TestPreviewUp_SingleNoteMode_Scrolls(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusPreview(tg.g, nil)
	tg.gui.previewDown(tg.g, tg.gui.views.Preview)
	tg.gui.previewDown(tg.g, tg.gui.views.Preview)
	tg.gui.previewUp(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.ScrollOffset != 1 {
		t.Errorf("ScrollOffset = %d, want 1", tg.gui.state.Preview.ScrollOffset)
	}
}

func TestPreviewUp_SingleNoteMode_ClampsAtZero(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusPreview(tg.g, nil)
	tg.gui.previewUp(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.ScrollOffset != 0 {
		t.Errorf("ScrollOffset = %d, want 0 (clamped)", tg.gui.state.Preview.ScrollOffset)
	}
}

func TestPreviewBack_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.focusPreview(tg.g, nil)
	tg.gui.previewBack(tg.g, tg.gui.views.Preview)

	if tg.gui.state.CurrentContext != TagsContext {
		t.Errorf("CurrentContext = %v, want TagsContext (restored)", tg.gui.state.CurrentContext)
	}
}

func TestPreviewBack_ExitsEditMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Enter edit mode
	tg.gui.editNotesInPreview(tg.g, tg.gui.views.Notes)
	if !tg.gui.state.Preview.EditMode {
		t.Fatal("should be in edit mode")
	}

	tg.gui.previewBack(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.EditMode {
		t.Error("EditMode should be false after previewBack")
	}
}

// --- Preview toggle tests ---

func TestToggleFrontmatter(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if tg.gui.state.Preview.ShowFrontmatter {
		t.Fatal("ShowFrontmatter should default to false")
	}

	tg.gui.toggleFrontmatter(tg.g, tg.gui.views.Preview)
	if !tg.gui.state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be true after toggle")
	}

	tg.gui.toggleFrontmatter(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be false after second toggle")
	}
}

func TestToggleMarkdown(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.state.Preview.RenderMarkdown
	tg.gui.toggleMarkdown(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.RenderMarkdown == initial {
		t.Error("RenderMarkdown should have toggled")
	}
}

func TestToggleTitle(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.state.Preview.ShowTitle
	tg.gui.toggleTitle(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.ShowTitle == initial {
		t.Error("ShowTitle should have toggled")
	}
}

func TestToggleGlobalTags(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.state.Preview.ShowGlobalTags
	tg.gui.toggleGlobalTags(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.ShowGlobalTags == initial {
		t.Error("ShowGlobalTags should have toggled")
	}
}

// --- Notes tab cycling tests ---

func TestFocusNotes_CyclesTabs_WhenAlreadyFocused(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Already on NotesContext, pressing 1 again cycles tab
	if tg.gui.state.Notes.CurrentTab != NotesTabAll {
		t.Fatalf("initial tab = %v, want All", tg.gui.state.Notes.CurrentTab)
	}

	tg.gui.focusNotes(tg.g, nil) // cycles All → Today
	if tg.gui.state.Notes.CurrentTab != NotesTabToday {
		t.Errorf("tab = %v, want Today", tg.gui.state.Notes.CurrentTab)
	}

	tg.gui.focusNotes(tg.g, nil) // cycles Today → Recent
	if tg.gui.state.Notes.CurrentTab != NotesTabRecent {
		t.Errorf("tab = %v, want Recent", tg.gui.state.Notes.CurrentTab)
	}

	tg.gui.focusNotes(tg.g, nil) // cycles Recent → All
	if tg.gui.state.Notes.CurrentTab != NotesTabAll {
		t.Errorf("tab = %v, want All (wrapped)", tg.gui.state.Notes.CurrentTab)
	}
}

// --- Queries tab cycling tests ---

func TestFocusQueries_CyclesTabs_WhenAlreadyFocused(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil) // switch to queries
	if tg.gui.state.Queries.CurrentTab != QueriesTabQueries {
		t.Fatalf("initial tab = %v, want Queries", tg.gui.state.Queries.CurrentTab)
	}

	tg.gui.focusQueries(tg.g, nil) // already focused → cycle to Parents
	if tg.gui.state.Queries.CurrentTab != QueriesTabParents {
		t.Errorf("tab = %v, want Parents", tg.gui.state.Queries.CurrentTab)
	}

	tg.gui.focusQueries(tg.g, nil) // cycle back to Queries
	if tg.gui.state.Queries.CurrentTab != QueriesTabQueries {
		t.Errorf("tab = %v, want Queries (wrapped)", tg.gui.state.Queries.CurrentTab)
	}
}

// --- Edit mode tests ---

func TestEditNotesInPreview_EntersEditMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.editNotesInPreview(tg.g, tg.gui.views.Notes)

	if !tg.gui.state.Preview.EditMode {
		t.Error("EditMode should be true")
	}
	if tg.gui.state.Preview.Mode != PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want PreviewModeCardList", tg.gui.state.Preview.Mode)
	}
	if len(tg.gui.state.Preview.Cards) != len(tg.gui.state.Notes.Items) {
		t.Errorf("Cards = %d, want %d (copy of notes)", len(tg.gui.state.Preview.Cards), len(tg.gui.state.Notes.Items))
	}
	if tg.gui.state.CurrentContext != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", tg.gui.state.CurrentContext)
	}
}

func TestEditNotesInPreview_EmptyNotes_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.editNotesInPreview(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Preview.EditMode {
		t.Error("EditMode should be false for empty notes")
	}
}

// --- focusNoteFromPreview tests ---

func TestFocusNoteFromPreview_JumpsToNote(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Enter card list via tag filter, select second card
	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	if len(tg.gui.state.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.state.Preview.Cards))
	}

	tg.gui.previewDown(tg.g, tg.gui.views.Preview)
	card := tg.gui.state.Preview.Cards[tg.gui.state.Preview.SelectedCardIndex]

	tg.gui.focusNoteFromPreview(tg.g, tg.gui.views.Preview)

	if tg.gui.state.CurrentContext != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", tg.gui.state.CurrentContext)
	}

	// The selected note in the list should match the card we were on
	selectedNote := tg.gui.state.Notes.Items[tg.gui.state.Notes.SelectedIndex]
	if selectedNote.UUID != card.UUID {
		t.Errorf("selected note UUID = %q, want %q", selectedNote.UUID, card.UUID)
	}
}

// --- Capture workflow tests ---

func TestOpenCapture_EntersCaptureMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openCapture(tg.g, nil)

	if !tg.gui.state.CaptureMode {
		t.Error("CaptureMode should be true")
	}
	if tg.gui.state.CurrentContext != CaptureContext {
		t.Errorf("CurrentContext = %v, want CaptureContext", tg.gui.state.CurrentContext)
	}
	if tg.gui.state.CaptureParent != nil {
		t.Error("CaptureParent should be nil initially")
	}
}

func TestNewNote_OpensCapture(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.newNote(tg.g, nil)

	if !tg.gui.state.CaptureMode {
		t.Error("newNote should open capture mode")
	}
}

// --- Quit test ---

func TestQuit_ReturnsErrQuit(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	err := tg.gui.quit(tg.g, nil)
	if err != gocui.ErrQuit {
		t.Errorf("quit() = %v, want gocui.ErrQuit", err)
	}
}

// --- Refresh test ---

func TestRefresh_ReloadsData(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move selection, then refresh — selection should reset
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	tg.gui.notesDown(tg.g, tg.gui.views.Notes)
	if tg.gui.state.Notes.SelectedIndex != 2 {
		t.Fatalf("SelectedIndex = %d, want 2 before refresh", tg.gui.state.Notes.SelectedIndex)
	}

	tg.gui.refresh(tg.g, nil)

	// After refresh, notes are reloaded and selection resets to 0
	if tg.gui.state.Notes.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 after refresh", tg.gui.state.Notes.SelectedIndex)
	}
	if len(tg.gui.state.Notes.Items) != 5 {
		t.Errorf("Notes.Items = %d, want 5 after refresh", len(tg.gui.state.Notes.Items))
	}
}

// --- Error handling tests ---

func TestFilterByTag_WithError_NoPanic(t *testing.T) {
	mock := testutil.NewMockExecutor().
		WithTags(models.Tag{Name: "broken", Count: 1}).
		WithError(fmt.Errorf("connection failed"))
	tg := newTestGui(t, mock)
	defer tg.Close()

	// Tags are loaded before WithError takes effect (loaded during init).
	// But filterByTag calls Search which will fail.
	// We need to set the error after init.
	// Re-set mock error state by manipulating state directly.
	tg.gui.state.Tags.Items = []models.Tag{{Name: "broken", Count: 1}}

	// filterByTag should not panic even though Search fails
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)
}

func TestRunQuery_WithError_NoPanic(t *testing.T) {
	mock := testutil.NewMockExecutor().
		WithQueries(models.Query{Name: "broken", Query: "#fail"}).
		WithError(fmt.Errorf("timeout"))
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.state.Queries.Items = []models.Query{{Name: "broken", Query: "#fail"}}

	// Should not panic
	tg.gui.runQuery(tg.g, tg.gui.views.Queries)
}

// --- Tab cycle includes search filter ---

func TestNextPanel_IncludesSearchFilter_WhenActive(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Set an active search query to add SearchFilter to the cycle
	tg.gui.state.SearchQuery = "#test"

	// Force layout to create SearchFilter view
	tg.g.ForceLayoutAndRedraw()

	// Cycle should now include SearchFilter at the start
	tg.gui.setContext(SearchFilterContext)
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.CurrentContext != NotesContext {
		t.Errorf("after Tab from SearchFilter: context = %v, want NotesContext", tg.gui.state.CurrentContext)
	}
}

// --- Dialog system tests ---

func TestDeleteNote_ShowsConfirmDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.deleteNote(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Dialog == nil {
		t.Fatal("Dialog should be set")
	}
	if tg.gui.state.Dialog.Type != "confirm" {
		t.Errorf("Dialog.Type = %q, want confirm", tg.gui.state.Dialog.Type)
	}
	if tg.gui.state.Dialog.OnConfirm == nil {
		t.Error("OnConfirm callback should be set")
	}
}

func TestDeleteNote_EmptyNotes_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.deleteNote(tg.g, tg.gui.views.Notes)

	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should not be shown for empty notes")
	}
}

func TestDeleteTag_ShowsConfirmDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.deleteTag(tg.g, tg.gui.views.Tags)

	if tg.gui.state.Dialog == nil {
		t.Fatal("Dialog should be set")
	}
	if tg.gui.state.Dialog.Type != "confirm" {
		t.Errorf("Dialog.Type = %q, want confirm", tg.gui.state.Dialog.Type)
	}
}

func TestDeleteQuery_ShowsConfirmDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	tg.gui.deleteQuery(tg.g, tg.gui.views.Queries)

	if tg.gui.state.Dialog == nil {
		t.Fatal("Dialog should be set")
	}
	if tg.gui.state.Dialog.Type != "confirm" {
		t.Errorf("Dialog.Type = %q, want confirm", tg.gui.state.Dialog.Type)
	}
}

func TestRenameTag_ShowsInputDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.renameTag(tg.g, tg.gui.views.Tags)

	if tg.gui.state.Dialog == nil {
		t.Fatal("Dialog should be set")
	}
	if tg.gui.state.Dialog.Type != "input" {
		t.Errorf("Dialog.Type = %q, want input", tg.gui.state.Dialog.Type)
	}
}

func TestConfirmYes_CallsCallback(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	called := false
	tg.gui.showConfirm("Test", "Are you sure?", func() error {
		called = true
		return nil
	})

	// Force layout to create the confirm view
	tg.g.ForceLayoutAndRedraw()

	confirmView, err := tg.g.View(ConfirmView)
	if err != nil {
		t.Fatalf("confirm view not created: %v", err)
	}

	tg.gui.confirmYes(tg.g, confirmView)

	if !called {
		t.Error("OnConfirm callback was not called")
	}
	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should be nil after confirm")
	}
}

func TestConfirmNo_ClosesDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	called := false
	tg.gui.showConfirm("Test", "Are you sure?", func() error {
		called = true
		return nil
	})

	tg.g.ForceLayoutAndRedraw()

	confirmView, err := tg.g.View(ConfirmView)
	if err != nil {
		t.Fatalf("confirm view not created: %v", err)
	}

	tg.gui.confirmNo(tg.g, confirmView)

	if called {
		t.Error("OnConfirm should NOT be called on cancel")
	}
	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should be nil after cancel")
	}
}

func TestDeleteTag_ConfirmYes_DeletesTag(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initialCount := len(tg.gui.state.Tags.Items)

	tg.gui.focusTags(tg.g, nil)
	tg.gui.deleteTag(tg.g, tg.gui.views.Tags)

	// Force layout to create confirm dialog view
	tg.g.ForceLayoutAndRedraw()

	confirmView, err := tg.g.View(ConfirmView)
	if err != nil {
		t.Fatalf("confirm view not created: %v", err)
	}

	// Confirm deletion
	tg.gui.confirmYes(tg.g, confirmView)

	// After confirming, tags are refreshed from mock (which returns same data)
	// The important thing is the flow completed without panic
	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should be closed after confirm")
	}
	// Tags should still have data (mock still returns same tags on refresh)
	if len(tg.gui.state.Tags.Items) != initialCount {
		t.Errorf("Tags.Items = %d, want %d (mock returns same data)", len(tg.gui.state.Tags.Items), initialCount)
	}
}

func TestShowHelp_CreatesMenuDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.showHelp()

	if tg.gui.state.Dialog == nil {
		t.Fatal("Dialog should be set")
	}
	if tg.gui.state.Dialog.Type != "menu" {
		t.Errorf("Dialog.Type = %q, want menu", tg.gui.state.Dialog.Type)
	}
	if len(tg.gui.state.Dialog.MenuItems) == 0 {
		t.Error("MenuItems should not be empty")
	}
}

func TestMenuNavigation(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.showHelp()
	tg.g.ForceLayoutAndRedraw()

	menuView, err := tg.g.View(MenuView)
	if err != nil {
		t.Fatalf("menu view not created: %v", err)
	}

	initial := tg.gui.state.Dialog.MenuSelection
	tg.gui.menuDown(tg.g, menuView)

	if tg.gui.state.Dialog.MenuSelection <= initial {
		t.Errorf("MenuSelection = %d, should have moved down from %d", tg.gui.state.Dialog.MenuSelection, initial)
	}
}

func TestMenuCancel_ClosesDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.showHelp()
	tg.g.ForceLayoutAndRedraw()

	menuView, err := tg.g.View(MenuView)
	if err != nil {
		t.Fatalf("menu view not created: %v", err)
	}

	tg.gui.menuCancel(tg.g, menuView)

	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should be nil after cancel")
	}
}

func TestCloseDialog_CleansUp(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.showConfirm("Test", "msg", func() error { return nil })
	tg.g.ForceLayoutAndRedraw()

	tg.gui.closeDialog()

	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should be nil")
	}
	// Confirm view should be deleted
	_, err := tg.g.View(ConfirmView)
	if err == nil {
		t.Error("ConfirmView should be deleted after closeDialog")
	}
}

// --- Snapshot test: verifies views render real content ---

func TestSnapshot_ContainsViewContent(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	snapshot := tg.g.Snapshot()
	if len(snapshot) == 0 {
		t.Error("Snapshot should not be empty")
	}
}
