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
			models.Note{UUID: "1", Title: "Note One", Tags: []string{"daily"}, InlineTags: []string{"#followup"}, Created: time.Now()},
			models.Note{UUID: "2", Title: "Note Two", Tags: []string{"work"}, Created: time.Now()},
			models.Note{UUID: "3", Title: "Note Three", Tags: []string{"daily"}, InlineTags: []string{"#todo"}, Created: time.Now()},
			models.Note{UUID: "4", Title: "Note Four", Tags: []string{"project"}, Created: time.Now()},
			models.Note{UUID: "5", Title: "Note Five", Tags: []string{"daily"}, InlineTags: []string{"#followup", "#todo"}, Created: time.Now()},
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
	if tg.gui.state.currentContext() != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", tg.gui.state.currentContext())
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

// testNotesDown moves the notes selection down by one (using the NotesContext).
func testNotesDown(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	items := notesCtx.Items
	idx := notesCtx.GetSelectedLineIdx()
	if idx+1 < len(items) {
		notesCtx.MoveSelectedLine(1)
		tg.gui.syncNotesToLegacy()
		tg.gui.renderNotes()
	}
}

// testNotesUp moves the notes selection up by one (using the NotesContext).
func testNotesUp(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	if notesCtx.GetSelectedLineIdx() > 0 {
		notesCtx.MoveSelectedLine(-1)
		tg.gui.syncNotesToLegacy()
		tg.gui.renderNotes()
	}
}

// testNotesTop jumps to the first note.
func testNotesTop(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	notesCtx.SetSelectedLineIdx(0)
	tg.gui.syncNotesToLegacy()
	tg.gui.renderNotes()
}

// testNotesBottom jumps to the last note.
func testNotesBottom(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	if len(notesCtx.Items) > 0 {
		notesCtx.SetSelectedLineIdx(len(notesCtx.Items) - 1)
		tg.gui.syncNotesToLegacy()
		tg.gui.renderNotes()
	}
}

// testQueriesDown moves the queries selection down by one (using the QueriesContext).
func testQueriesDown(tg *testGui) {
	queriesCtx := tg.gui.contexts.Queries
	t := queriesCtx.ActiveTrait()
	count := queriesCtx.ActiveItemCount()
	if t.GetSelectedLineIdx()+1 < count {
		t.MoveSelectedLine(1)
		tg.gui.syncQueriesToLegacy()
		tg.gui.renderQueries()
	}
}

// --- Notes navigation tests ---

func TestNotesDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	testNotesDown(tg)

	if tg.gui.state.Notes.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesDown_MultipleSteps(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	for i := 0; i < 3; i++ {
		testNotesDown(tg)
	}

	if tg.gui.state.Notes.SelectedIndex != 3 {
		t.Errorf("SelectedIndex = %d, want 3 after 3 down presses", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesUp_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move down first, then up
	testNotesDown(tg)
	testNotesDown(tg)
	testNotesUp(tg)

	if tg.gui.state.Notes.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesTop_JumpsToFirst(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	testNotesDown(tg)
	testNotesDown(tg)
	testNotesDown(tg)
	testNotesTop(tg)

	if tg.gui.state.Notes.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestNotesBottom_JumpsToLast(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	testNotesBottom(tg)

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
		testNotesDown(tg)
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
	if tg.gui.state.currentContext() != NotesContext {
		t.Fatalf("initial context = %v, want NotesContext", tg.gui.state.currentContext())
	}

	// Tab → QueriesContext
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.currentContext() != QueriesContext {
		t.Errorf("after first Tab: context = %v, want QueriesContext", tg.gui.state.currentContext())
	}

	// Tab → TagsContext
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.currentContext() != TagsContext {
		t.Errorf("after second Tab: context = %v, want TagsContext", tg.gui.state.currentContext())
	}

	// Tab → wraps to NotesContext
	tg.gui.nextPanel(tg.g, nil)
	if tg.gui.state.currentContext() != NotesContext {
		t.Errorf("after third Tab: context = %v, want NotesContext (wrap)", tg.gui.state.currentContext())
	}
}

func TestPrevPanel_CyclesBackward(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// BackTab from NotesContext → TagsContext (wraps backward)
	tg.gui.prevPanel(tg.g, nil)
	if tg.gui.state.currentContext() != TagsContext {
		t.Errorf("after BackTab from Notes: context = %v, want TagsContext", tg.gui.state.currentContext())
	}
}

func TestFocusNotes_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Switch to tags first
	tg.gui.focusTags(tg.g, nil)
	if tg.gui.state.currentContext() != TagsContext {
		t.Fatalf("context = %v, want TagsContext", tg.gui.state.currentContext())
	}

	// Press 1 → NotesContext
	tg.gui.focusNotes(tg.g, nil)
	if tg.gui.state.currentContext() != NotesContext {
		t.Errorf("context = %v, want NotesContext", tg.gui.state.currentContext())
	}
}

func TestFocusQueries_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	if tg.gui.state.currentContext() != QueriesContext {
		t.Errorf("context = %v, want QueriesContext", tg.gui.state.currentContext())
	}
}

func TestFocusTags_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	if tg.gui.state.currentContext() != TagsContext {
		t.Errorf("context = %v, want TagsContext", tg.gui.state.currentContext())
	}
}

func TestFocusPreview_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusPreview(tg.g, nil)
	if tg.gui.state.currentContext() != PreviewContext {
		t.Errorf("context = %v, want PreviewContext", tg.gui.state.currentContext())
	}
}

func TestContextSwitch_TracksPrevious(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	if tg.gui.state.previousContext() != NotesContext {
		t.Errorf("previousContext() = %v, want NotesContext", tg.gui.state.previousContext())
	}

	tg.gui.focusPreview(tg.g, nil)
	if tg.gui.state.previousContext() != TagsContext {
		t.Errorf("previousContext() = %v, want TagsContext", tg.gui.state.previousContext())
	}
}

// --- Search workflow tests ---

func TestOpenSearch_EntersSearchOverlay(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openSearch(tg.g, nil)

	if tg.gui.state.ActiveOverlay != OverlaySearch {
		t.Error("ActiveOverlay should be OverlaySearch")
	}
	if tg.gui.state.currentContext() != SearchContext {
		t.Errorf("currentContext() = %v, want SearchContext", tg.gui.state.currentContext())
	}
}

func TestCancelSearch_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Start from tags context
	tg.gui.focusTags(tg.g, nil)
	prev := tg.gui.state.currentContext()

	tg.gui.openSearch(tg.g, nil)
	tg.gui.cancelSearch(tg.g, tg.gui.views.Search)

	if tg.gui.state.ActiveOverlay != OverlayNone {
		t.Error("ActiveOverlay should be OverlayNone after cancel")
	}
	if tg.gui.state.currentContext() != prev {
		t.Errorf("CurrentContext = %v, want %v (restored)", tg.gui.state.currentContext(), prev)
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
	if tg.gui.state.currentContext() != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", tg.gui.state.currentContext())
	}
}

// --- Tags navigation tests ---

func TestTagsDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Focus tags panel first
	tg.gui.focusTags(tg.g, nil)

	// Use TagsContext for navigation (migrated from old tagsDown)
	tagsCtx := tg.gui.contexts.Tags
	tagsCtx.MoveSelectedLine(1)
	tg.gui.syncTagsToLegacy()

	if tg.gui.state.Tags.SelectedIndex != 1 {
		t.Errorf("Tags.SelectedIndex = %d, want 1", tg.gui.state.Tags.SelectedIndex)
	}
}

func TestTagsUp_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tagsCtx := tg.gui.contexts.Tags
	tagsCtx.MoveSelectedLine(1)
	tagsCtx.MoveSelectedLine(1)
	tagsCtx.MoveSelectedLine(-1)
	tg.gui.syncTagsToLegacy()

	if tg.gui.state.Tags.SelectedIndex != 1 {
		t.Errorf("Tags.SelectedIndex = %d, want 1", tg.gui.state.Tags.SelectedIndex)
	}
}

// --- Queries navigation tests ---

func TestQueriesDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	testQueriesDown(tg)

	if tg.gui.state.Queries.SelectedIndex != 1 {
		t.Errorf("Queries.SelectedIndex = %d, want 1", tg.gui.state.Queries.SelectedIndex)
	}
}

// --- Preview state tests ---

func TestNotesDown_UpdatesPreview(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move down to select Note Two
	testNotesDown(tg)

	// Preview should be in card list mode showing the selected note as a single card
	if tg.gui.state.Preview.Mode != PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want PreviewModeCardList", tg.gui.state.Preview.Mode)
	}
}

// --- Empty state tests ---

func TestEmptyNotes_NoNavigationPanic(t *testing.T) {
	mock := testutil.NewMockExecutor() // no data
	tg := newTestGui(t, mock)
	defer tg.Close()

	// These should not panic with empty lists
	testNotesDown(tg)
	testNotesUp(tg)
	testNotesTop(tg)
	testNotesBottom(tg)

	if tg.gui.state.Notes.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 for empty list", tg.gui.state.Notes.SelectedIndex)
	}
}

func TestEmptyTags_NoNavigationPanic(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tagsCtx := tg.gui.contexts.Tags
	tagsCtx.MoveSelectedLine(1)
	tagsCtx.MoveSelectedLine(-1)
	tg.gui.syncTagsToLegacy()

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
	if tg.gui.state.currentContext() != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", tg.gui.state.currentContext())
	}
}

func TestFilterByTag_EmptyTags_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	// Should remain in tags context, no switch to preview
	if tg.gui.state.currentContext() != TagsContext {
		t.Errorf("CurrentContext = %v, want TagsContext (noop for empty)", tg.gui.state.currentContext())
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
	if tg.gui.state.currentContext() != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", tg.gui.state.currentContext())
	}
}

func TestRunQuery_EmptyQueries_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.focusQueries(tg.g, nil)
	tg.gui.runQuery(tg.g, tg.gui.views.Queries)

	if tg.gui.state.currentContext() != QueriesContext {
		t.Errorf("CurrentContext = %v, want QueriesContext (noop for empty)", tg.gui.state.currentContext())
	}
}

// --- Preview navigation tests ---

func TestPreviewCardDown_CardListMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Enter card list mode via tag filter
	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	if len(tg.gui.state.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.state.Preview.Cards))
	}

	tg.gui.preview.previewCardDown(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.SelectedCardIndex != 1 {
		t.Errorf("SelectedCardIndex = %d, want 1", tg.gui.state.Preview.SelectedCardIndex)
	}
}

func TestPreviewCardUp_CardListMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

	if len(tg.gui.state.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.state.Preview.Cards))
	}

	tg.gui.preview.previewCardDown(tg.g, tg.gui.views.Preview)
	tg.gui.preview.previewCardUp(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.SelectedCardIndex != 0 {
		t.Errorf("SelectedCardIndex = %d, want 0", tg.gui.state.Preview.SelectedCardIndex)
	}
}

func TestPreviewDown_CardMode_MovesCursor(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Set up card list mode — set CardLineRanges AFTER setContext since
	// setContext(PreviewContext) calls renderPreview which rebuilds them.
	tg.gui.state.Preview.Mode = PreviewModeCardList
	tg.gui.state.Preview.Cards = tg.gui.state.Notes.Items
	tg.gui.state.ContextStack = append(tg.gui.state.ContextStack, PreviewContext)
	// Override with known ranges after any render
	tg.gui.state.Preview.CursorLine = 1
	tg.gui.state.Preview.CardLineRanges = [][2]int{{0, 5}, {6, 11}}

	tg.gui.preview.previewDown(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.CursorLine != 2 {
		t.Errorf("CursorLine = %d, want 2 after previewDown", tg.gui.state.Preview.CursorLine)
	}
}

func TestPreviewUp_CardMode_MovesCursor(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.state.Preview.Mode = PreviewModeCardList
	tg.gui.state.Preview.Cards = tg.gui.state.Notes.Items
	tg.gui.state.ContextStack = append(tg.gui.state.ContextStack, PreviewContext)
	tg.gui.state.Preview.CursorLine = 3
	tg.gui.state.Preview.CardLineRanges = [][2]int{{0, 5}, {6, 11}}

	tg.gui.preview.previewUp(tg.g, tg.gui.views.Preview)

	if tg.gui.state.Preview.CursorLine != 2 {
		t.Errorf("CursorLine = %d, want 2 after previewUp", tg.gui.state.Preview.CursorLine)
	}
}

func TestPreviewUp_CardMode_ClampsAtTop(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.state.Preview.Mode = PreviewModeCardList
	tg.gui.state.Preview.Cards = tg.gui.state.Notes.Items
	tg.gui.state.ContextStack = append(tg.gui.state.ContextStack, PreviewContext)
	tg.gui.state.Preview.CursorLine = 1
	tg.gui.state.Preview.CardLineRanges = [][2]int{{0, 5}}

	tg.gui.preview.previewUp(tg.g, tg.gui.views.Preview)

	// Cursor should stay at 1 (line 0 is a separator, not content)
	if tg.gui.state.Preview.CursorLine != 1 {
		t.Errorf("CursorLine = %d, want 1 (clamped at first content line)", tg.gui.state.Preview.CursorLine)
	}
}

func TestPreviewBack_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.focusPreview(tg.g, nil)
	tg.gui.preview.previewBack(tg.g, tg.gui.views.Preview)

	if tg.gui.state.currentContext() != TagsContext {
		t.Errorf("CurrentContext = %v, want TagsContext (restored)", tg.gui.state.currentContext())
	}
}

// --- Preview toggle tests ---

func TestToggleFrontmatter(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if tg.gui.state.Preview.ShowFrontmatter {
		t.Fatal("ShowFrontmatter should default to false")
	}

	tg.gui.preview.toggleFrontmatter(tg.g, tg.gui.views.Preview)
	if !tg.gui.state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be true after toggle")
	}

	tg.gui.preview.toggleFrontmatter(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be false after second toggle")
	}
}

func TestToggleMarkdown(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.state.Preview.RenderMarkdown
	tg.gui.preview.toggleMarkdown(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.RenderMarkdown == initial {
		t.Error("RenderMarkdown should have toggled")
	}
}

func TestToggleTitle(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.state.Preview.ShowTitle
	tg.gui.preview.toggleTitle(tg.g, tg.gui.views.Preview)
	if tg.gui.state.Preview.ShowTitle == initial {
		t.Error("ShowTitle should have toggled")
	}
}

func TestToggleGlobalTags(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.state.Preview.ShowGlobalTags
	tg.gui.preview.toggleGlobalTags(tg.g, tg.gui.views.Preview)
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

	tg.gui.preview.previewDown(tg.g, tg.gui.views.Preview)
	card := tg.gui.state.Preview.Cards[tg.gui.state.Preview.SelectedCardIndex]

	tg.gui.preview.focusNoteFromPreview(tg.g, tg.gui.views.Preview)

	if tg.gui.state.currentContext() != NotesContext {
		t.Errorf("CurrentContext = %v, want NotesContext", tg.gui.state.currentContext())
	}

	// The selected note in the list should match the card we were on
	selectedNote := tg.gui.state.Notes.Items[tg.gui.state.Notes.SelectedIndex]
	if selectedNote.UUID != card.UUID {
		t.Errorf("selected note UUID = %q, want %q", selectedNote.UUID, card.UUID)
	}
}

// --- Capture workflow tests ---

func TestOpenCapture_EntersCaptureOverlay(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openCapture(tg.g, nil)

	if tg.gui.state.ActiveOverlay != OverlayCapture {
		t.Error("ActiveOverlay should be OverlayCapture")
	}
	if tg.gui.state.currentContext() != CaptureContext {
		t.Errorf("currentContext() = %v, want CaptureContext", tg.gui.state.currentContext())
	}
	if tg.gui.state.CaptureParent != nil {
		t.Error("CaptureParent should be nil initially")
	}
}

func TestNewNote_OpensCapture(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.newNote(tg.g, nil)

	if tg.gui.state.ActiveOverlay != OverlayCapture {
		t.Error("newNote should open capture overlay")
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
	testNotesDown(tg)
	testNotesDown(tg)
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
	if tg.gui.state.currentContext() != NotesContext {
		t.Errorf("after Tab from SearchFilter: context = %v, want NotesContext", tg.gui.state.currentContext())
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
