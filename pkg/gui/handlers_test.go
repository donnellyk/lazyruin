package gui

import (
	"fmt"
	"testing"
	"time"

	"kvnd/lazyruin/pkg/gui/context"
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
	if tg.gui.contextMgr.Current() != "notes" {
		t.Errorf("CurrentContext = %v, want notes", tg.gui.contextMgr.Current())
	}
}

func TestHeadlessGui_LoadsNotes(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if len(tg.gui.contexts.Notes.Items) != 5 {
		t.Errorf("Notes.Items = %d, want 5", len(tg.gui.contexts.Notes.Items))
	}
	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 0 {
		t.Errorf("Notes.SelectedIndex = %d, want 0", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
}

func TestHeadlessGui_LoadsTags(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if len(tg.gui.contexts.Tags.Items) != 3 {
		t.Errorf("Tags.Items = %d, want 3", len(tg.gui.contexts.Tags.Items))
	}
}

func TestHeadlessGui_LoadsQueries(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if len(tg.gui.contexts.Queries.Queries) != 2 {
		t.Errorf("Queries.Items = %d, want 2", len(tg.gui.contexts.Queries.Queries))
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

// testNotesDown moves the notes selection down by one (using the "notes").
func testNotesDown(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	items := notesCtx.Items
	idx := notesCtx.GetSelectedLineIdx()
	if idx+1 < len(items) {
		notesCtx.MoveSelectedLine(1)

		tg.gui.RenderNotes()
	}
}

// testNotesUp moves the notes selection up by one (using the "notes").
func testNotesUp(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	if notesCtx.GetSelectedLineIdx() > 0 {
		notesCtx.MoveSelectedLine(-1)

		tg.gui.RenderNotes()
	}
}

// testNotesTop jumps to the first note.
func testNotesTop(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	notesCtx.SetSelectedLineIdx(0)

	tg.gui.RenderNotes()
}

// testNotesBottom jumps to the last note.
func testNotesBottom(tg *testGui) {
	notesCtx := tg.gui.contexts.Notes
	if len(notesCtx.Items) > 0 {
		notesCtx.SetSelectedLineIdx(len(notesCtx.Items) - 1)

		tg.gui.RenderNotes()
	}
}

// testQueriesDown moves the queries selection down by one (using the "queries").
func testQueriesDown(tg *testGui) {
	queriesCtx := tg.gui.contexts.Queries
	t := queriesCtx.ActiveTrait()
	count := queriesCtx.ActiveItemCount()
	if t.GetSelectedLineIdx()+1 < count {
		t.MoveSelectedLine(1)

		tg.gui.RenderQueries()
	}
}

// --- Notes navigation tests ---

func TestNotesDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	testNotesDown(tg)

	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
}

func TestNotesDown_MultipleSteps(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	for i := 0; i < 3; i++ {
		testNotesDown(tg)
	}

	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 3 {
		t.Errorf("SelectedIndex = %d, want 3 after 3 down presses", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
}

func TestNotesUp_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move down first, then up
	testNotesDown(tg)
	testNotesDown(tg)
	testNotesUp(tg)

	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
}

func TestNotesTop_JumpsToFirst(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	testNotesDown(tg)
	testNotesDown(tg)
	testNotesDown(tg)
	testNotesTop(tg)

	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
}

func TestNotesBottom_JumpsToLast(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	testNotesBottom(tg)

	want := len(tg.gui.contexts.Notes.Items) - 1
	if tg.gui.contexts.Notes.GetSelectedLineIdx() != want {
		t.Errorf("SelectedIndex = %d, want %d", tg.gui.contexts.Notes.GetSelectedLineIdx(), want)
	}
}

func TestNotesDown_StopsAtBoundary(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	count := len(tg.gui.contexts.Notes.Items)
	for i := 0; i < count+5; i++ {
		testNotesDown(tg)
	}

	if tg.gui.contexts.Notes.GetSelectedLineIdx() != count-1 {
		t.Errorf("SelectedIndex = %d, want %d (should clamp at last)", tg.gui.contexts.Notes.GetSelectedLineIdx(), count-1)
	}
}

// --- Context switching tests ---

func TestNextPanel_CyclesThroughContexts(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Start at "notes"
	if tg.gui.contextMgr.Current() != "notes" {
		t.Fatalf("initial context = %v, want notes", tg.gui.contextMgr.Current())
	}

	// Tab → "queries"
	tg.gui.globalController.NextPanel()
	if tg.gui.contextMgr.Current() != "queries" {
		t.Errorf("after first Tab: context = %v, want queries", tg.gui.contextMgr.Current())
	}

	// Tab → "tags"
	tg.gui.globalController.NextPanel()
	if tg.gui.contextMgr.Current() != "tags" {
		t.Errorf("after second Tab: context = %v, want tags", tg.gui.contextMgr.Current())
	}

	// Tab → wraps to "notes"
	tg.gui.globalController.NextPanel()
	if tg.gui.contextMgr.Current() != "notes" {
		t.Errorf("after third Tab: context = %v, want notes (wrap)", tg.gui.contextMgr.Current())
	}
}

func TestPrevPanel_CyclesBackward(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// BackTab from "notes" → "tags" (wraps backward)
	tg.gui.globalController.PrevPanel()
	if tg.gui.contextMgr.Current() != "tags" {
		t.Errorf("after BackTab from Notes: context = %v, want tags", tg.gui.contextMgr.Current())
	}
}

func TestFocusNotes_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Switch to tags first
	tg.gui.globalController.FocusTags()
	if tg.gui.contextMgr.Current() != "tags" {
		t.Fatalf("context = %v, want tags", tg.gui.contextMgr.Current())
	}

	// Press 1 → "notes"
	tg.gui.globalController.FocusNotes()
	if tg.gui.contextMgr.Current() != "notes" {
		t.Errorf("context = %v, want notes", tg.gui.contextMgr.Current())
	}
}

func TestFocusQueries_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusQueries()
	if tg.gui.contextMgr.Current() != "queries" {
		t.Errorf("context = %v, want queries", tg.gui.contextMgr.Current())
	}
}

func TestFocusTags_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	if tg.gui.contextMgr.Current() != "tags" {
		t.Errorf("context = %v, want tags", tg.gui.contextMgr.Current())
	}
}

func TestFocusPreview_SwitchesContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusPreview()
	if tg.gui.contextMgr.Current() != "preview" {
		t.Errorf("context = %v, want preview", tg.gui.contextMgr.Current())
	}
}

func TestContextSwitch_TracksPrevious(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	if tg.gui.contextMgr.Previous() != "notes" {
		t.Errorf("previousContext() = %v, want notes", tg.gui.contextMgr.Previous())
	}

	tg.gui.globalController.FocusPreview()
	if tg.gui.contextMgr.Previous() != "tags" {
		t.Errorf("previousContext() = %v, want tags", tg.gui.contextMgr.Previous())
	}
}

// --- Search workflow tests ---

func TestOpenSearch_EntersSearchOverlay(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Search().OpenSearch()

	if !tg.gui.popupActive() {
		t.Error("popupActive() should be true after openSearch")
	}
	if tg.gui.contextMgr.Current() != "search" {
		t.Errorf("currentContext() = %v, want search", tg.gui.contextMgr.Current())
	}
}

func TestCancelSearch_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Start from tags context
	tg.gui.globalController.FocusTags()
	prev := tg.gui.contextMgr.Current()

	tg.gui.helpers.Search().OpenSearch()
	tg.gui.helpers.Search().CancelSearch()

	if tg.gui.popupActive() {
		t.Error("popupActive() should be false after cancelSearch")
	}
	if tg.gui.contextMgr.Current() != prev {
		t.Errorf("CurrentContext = %v, want %v (restored)", tg.gui.contextMgr.Current(), prev)
	}
}

func TestClearSearch_ResetsState(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Simulate an active search by setting SearchQuery
	tg.gui.state.SearchQuery = "#daily"

	tg.gui.helpers.Search().ClearSearch()

	if tg.gui.state.SearchQuery != "" {
		t.Errorf("SearchQuery = %q, want empty", tg.gui.state.SearchQuery)
	}
	if tg.gui.contexts.Notes.CurrentTab != context.NotesTabAll {
		t.Errorf("CurrentTab = %v, want NotesTabAll", tg.gui.contexts.Notes.CurrentTab)
	}
	if tg.gui.contextMgr.Current() != "notes" {
		t.Errorf("CurrentContext = %v, want notes", tg.gui.contextMgr.Current())
	}
}

// --- Tags navigation tests ---

func TestTagsDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Focus tags panel first
	tg.gui.globalController.FocusTags()

	// Use "tags" for navigation (migrated from old tagsDown)
	tagsCtx := tg.gui.contexts.Tags
	tagsCtx.MoveSelectedLine(1)

	if tg.gui.contexts.Tags.GetSelectedLineIdx() != 1 {
		t.Errorf("Tags.SelectedIndex = %d, want 1", tg.gui.contexts.Tags.GetSelectedLineIdx())
	}
}

func TestTagsUp_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tagsCtx := tg.gui.contexts.Tags
	tagsCtx.MoveSelectedLine(1)
	tagsCtx.MoveSelectedLine(1)
	tagsCtx.MoveSelectedLine(-1)

	if tg.gui.contexts.Tags.GetSelectedLineIdx() != 1 {
		t.Errorf("Tags.SelectedIndex = %d, want 1", tg.gui.contexts.Tags.GetSelectedLineIdx())
	}
}

// --- Queries navigation tests ---

func TestQueriesDown_MovesSelection(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusQueries()
	testQueriesDown(tg)

	if tg.gui.contexts.Queries.QueriesTrait().GetSelectedLineIdx() != 1 {
		t.Errorf("Queries.SelectedIndex = %d, want 1", tg.gui.contexts.Queries.QueriesTrait().GetSelectedLineIdx())
	}
}

// --- Preview state tests ---

func TestNotesDown_UpdatesPreview(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Move down to select Note Two
	testNotesDown(tg)

	// Preview should be in card list mode showing the selected note as a single card
	if tg.gui.contexts.Preview.Mode != context.PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want context.PreviewModeCardList", tg.gui.contexts.Preview.Mode)
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

	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 0 {
		t.Errorf("SelectedIndex = %d, want 0 for empty list", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
}

func TestEmptyTags_NoNavigationPanic(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tagsCtx := tg.gui.contexts.Tags
	tagsCtx.MoveSelectedLine(1)
	tagsCtx.MoveSelectedLine(-1)

	if tg.gui.contexts.Tags.GetSelectedLineIdx() != 0 {
		t.Errorf("SelectedIndex = %d, want 0 for empty list", tg.gui.contexts.Tags.GetSelectedLineIdx())
	}
}

// --- Filter by tag tests ---

func TestFilterByTag_SetsPreviewCardList(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())

	if tg.gui.contexts.Preview.Mode != context.PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want context.PreviewModeCardList", tg.gui.contexts.Preview.Mode)
	}
	if tg.gui.contexts.Preview.SelectedCardIndex != 0 {
		t.Errorf("SelectedCardIndex = %d, want 0", tg.gui.contexts.Preview.SelectedCardIndex)
	}
	if tg.gui.contextMgr.Current() != "preview" {
		t.Errorf("CurrentContext = %v, want preview", tg.gui.contextMgr.Current())
	}
}

func TestFilterByTag_EmptyTags_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())

	// Should remain in tags context, no switch to preview
	if tg.gui.contextMgr.Current() != "tags" {
		t.Errorf("CurrentContext = %v, want tags (noop for empty)", tg.gui.contextMgr.Current())
	}
}

// --- Run query tests ---

func TestRunQuery_SetsPreviewCardList(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusQueries()
	tg.gui.helpers.Queries().RunQuery()

	if tg.gui.contexts.Preview.Mode != context.PreviewModeCardList {
		t.Errorf("Preview.Mode = %v, want context.PreviewModeCardList", tg.gui.contexts.Preview.Mode)
	}
	if len(tg.gui.contexts.Preview.Cards) == 0 {
		t.Error("Preview.Cards should not be empty after running query")
	}
	if tg.gui.contextMgr.Current() != "preview" {
		t.Errorf("CurrentContext = %v, want preview", tg.gui.contextMgr.Current())
	}
}

func TestRunQuery_EmptyQueries_Noop(t *testing.T) {
	mock := testutil.NewMockExecutor()
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.globalController.FocusQueries()
	tg.gui.helpers.Queries().RunQuery()

	if tg.gui.contextMgr.Current() != "queries" {
		t.Errorf("CurrentContext = %v, want queries (noop for empty)", tg.gui.contextMgr.Current())
	}
}

// --- Preview navigation tests ---

func TestPreviewCardDown_CardListMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Enter card list mode via tag filter
	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())

	if len(tg.gui.contexts.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.contexts.Preview.Cards))
	}

	tg.gui.helpers.Preview().CardDown()

	if tg.gui.contexts.Preview.SelectedCardIndex != 1 {
		t.Errorf("SelectedCardIndex = %d, want 1", tg.gui.contexts.Preview.SelectedCardIndex)
	}
}

func TestPreviewCardUp_CardListMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())

	if len(tg.gui.contexts.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.contexts.Preview.Cards))
	}

	tg.gui.helpers.Preview().CardDown()
	tg.gui.helpers.Preview().CardUp()

	if tg.gui.contexts.Preview.SelectedCardIndex != 0 {
		t.Errorf("SelectedCardIndex = %d, want 0", tg.gui.contexts.Preview.SelectedCardIndex)
	}
}

func TestPreviewDown_CardMode_MovesCursor(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Set up card list mode — set CardLineRanges AFTER setContext since
	// setContext("preview") calls renderPreview which rebuilds them.
	tg.gui.contexts.Preview.Mode = context.PreviewModeCardList
	tg.gui.contexts.Preview.Cards = tg.gui.contexts.Notes.Items
	tg.gui.contextMgr.Push("preview")
	// Override with known ranges after any render
	tg.gui.contexts.Preview.CursorLine = 1
	tg.gui.contexts.Preview.CardLineRanges = [][2]int{{0, 5}, {6, 11}}

	tg.gui.helpers.Preview().MoveDown()

	if tg.gui.contexts.Preview.CursorLine != 2 {
		t.Errorf("CursorLine = %d, want 2 after previewDown", tg.gui.contexts.Preview.CursorLine)
	}
}

func TestPreviewUp_CardMode_MovesCursor(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.contexts.Preview.Mode = context.PreviewModeCardList
	tg.gui.contexts.Preview.Cards = tg.gui.contexts.Notes.Items
	tg.gui.contextMgr.Push("preview")
	tg.gui.contexts.Preview.CursorLine = 3
	tg.gui.contexts.Preview.CardLineRanges = [][2]int{{0, 5}, {6, 11}}

	tg.gui.helpers.Preview().MoveUp()

	if tg.gui.contexts.Preview.CursorLine != 2 {
		t.Errorf("CursorLine = %d, want 2 after previewUp", tg.gui.contexts.Preview.CursorLine)
	}
}

func TestPreviewUp_CardMode_ClampsAtTop(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.contexts.Preview.Mode = context.PreviewModeCardList
	tg.gui.contexts.Preview.Cards = tg.gui.contexts.Notes.Items
	tg.gui.contextMgr.Push("preview")
	tg.gui.contexts.Preview.CursorLine = 1
	tg.gui.contexts.Preview.CardLineRanges = [][2]int{{0, 5}}

	tg.gui.helpers.Preview().MoveUp()

	// Cursor should stay at 1 (line 0 is a separator, not content)
	if tg.gui.contexts.Preview.CursorLine != 1 {
		t.Errorf("CursorLine = %d, want 1 (clamped at first content line)", tg.gui.contexts.Preview.CursorLine)
	}
}

func TestPreviewBack_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.globalController.FocusPreview()
	tg.gui.helpers.Preview().Back()

	if tg.gui.contextMgr.Current() != "tags" {
		t.Errorf("CurrentContext = %v, want tags (restored)", tg.gui.contextMgr.Current())
	}
}

// --- Preview toggle tests ---

func TestToggleFrontmatter(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	if tg.gui.contexts.Preview.ShowFrontmatter {
		t.Fatal("ShowFrontmatter should default to false")
	}

	tg.gui.helpers.Preview().ToggleFrontmatter()
	if !tg.gui.contexts.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be true after toggle")
	}

	tg.gui.helpers.Preview().ToggleFrontmatter()
	if tg.gui.contexts.Preview.ShowFrontmatter {
		t.Error("ShowFrontmatter should be false after second toggle")
	}
}

func TestToggleMarkdown(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.contexts.Preview.RenderMarkdown
	tg.gui.helpers.Preview().ToggleMarkdown()
	if tg.gui.contexts.Preview.RenderMarkdown == initial {
		t.Error("RenderMarkdown should have toggled")
	}
}

func TestToggleTitle(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.contexts.Preview.ShowTitle
	tg.gui.helpers.Preview().ToggleTitle()
	if tg.gui.contexts.Preview.ShowTitle == initial {
		t.Error("ShowTitle should have toggled")
	}
}

func TestToggleGlobalTags(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	initial := tg.gui.contexts.Preview.ShowGlobalTags
	tg.gui.helpers.Preview().ToggleGlobalTags()
	if tg.gui.contexts.Preview.ShowGlobalTags == initial {
		t.Error("ShowGlobalTags should have toggled")
	}
}

// --- Notes tab cycling tests ---

func TestFocusNotes_CyclesTabs_WhenAlreadyFocused(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Already on "notes", pressing 1 again cycles tab
	if tg.gui.contexts.Notes.CurrentTab != context.NotesTabAll {
		t.Fatalf("initial tab = %v, want All", tg.gui.contexts.Notes.CurrentTab)
	}

	tg.gui.globalController.FocusNotes() // cycles All → Today
	if tg.gui.contexts.Notes.CurrentTab != context.NotesTabToday {
		t.Errorf("tab = %v, want Today", tg.gui.contexts.Notes.CurrentTab)
	}

	tg.gui.globalController.FocusNotes() // cycles Today → Recent
	if tg.gui.contexts.Notes.CurrentTab != context.NotesTabRecent {
		t.Errorf("tab = %v, want Recent", tg.gui.contexts.Notes.CurrentTab)
	}

	tg.gui.globalController.FocusNotes() // cycles Recent → All
	if tg.gui.contexts.Notes.CurrentTab != context.NotesTabAll {
		t.Errorf("tab = %v, want All (wrapped)", tg.gui.contexts.Notes.CurrentTab)
	}
}

// --- Queries tab cycling tests ---

func TestFocusQueries_CyclesTabs_WhenAlreadyFocused(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusQueries() // switch to queries
	if tg.gui.contexts.Queries.CurrentTab != context.QueriesTabQueries {
		t.Fatalf("initial tab = %v, want Queries", tg.gui.contexts.Queries.CurrentTab)
	}

	tg.gui.globalController.FocusQueries() // already focused → cycle to Parents
	if tg.gui.contexts.Queries.CurrentTab != context.QueriesTabParents {
		t.Errorf("tab = %v, want Parents", tg.gui.contexts.Queries.CurrentTab)
	}

	tg.gui.globalController.FocusQueries() // cycle back to Queries
	if tg.gui.contexts.Queries.CurrentTab != context.QueriesTabQueries {
		t.Errorf("tab = %v, want Queries (wrapped)", tg.gui.contexts.Queries.CurrentTab)
	}
}

// --- focusNoteFromPreview tests ---

func TestFocusNoteFromPreview_JumpsToNote(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Enter card list via tag filter, select second card
	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())

	if len(tg.gui.contexts.Preview.Cards) < 2 {
		t.Skipf("need at least 2 cards, got %d", len(tg.gui.contexts.Preview.Cards))
	}

	tg.gui.helpers.Preview().MoveDown()
	card := tg.gui.contexts.Preview.Cards[tg.gui.contexts.Preview.SelectedCardIndex]

	tg.gui.helpers.Preview().FocusNote()

	if tg.gui.contextMgr.Current() != "notes" {
		t.Errorf("CurrentContext = %v, want notes", tg.gui.contextMgr.Current())
	}

	// The selected note in the list should match the card we were on
	selectedNote := tg.gui.contexts.Notes.Items[tg.gui.contexts.Notes.GetSelectedLineIdx()]
	if selectedNote.UUID != card.UUID {
		t.Errorf("selected note UUID = %q, want %q", selectedNote.UUID, card.UUID)
	}
}

// --- Capture workflow tests ---

func TestOpenCapture_EntersCaptureOverlay(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()

	if !tg.gui.popupActive() {
		t.Error("popupActive() should be true after openCapture")
	}
	if tg.gui.contextMgr.Current() != "capture" {
		t.Errorf("currentContext() = %v, want capture", tg.gui.contextMgr.Current())
	}
	if tg.gui.contexts.Capture.Parent != nil {
		t.Error("CaptureParent should be nil initially")
	}
}

func TestNewNote_OpensCapture(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()

	if tg.gui.contextMgr.Current() != "capture" {
		t.Error("newNote should push capture context")
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
	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 2 {
		t.Fatalf("SelectedIndex = %d, want 2 before refresh", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}

	tg.gui.refresh(tg.g, nil)

	// After refresh, notes are reloaded and selection resets to 0
	if tg.gui.contexts.Notes.GetSelectedLineIdx() != 0 {
		t.Errorf("SelectedIndex = %d, want 0 after refresh", tg.gui.contexts.Notes.GetSelectedLineIdx())
	}
	if len(tg.gui.contexts.Notes.Items) != 5 {
		t.Errorf("Notes.Items = %d, want 5 after refresh", len(tg.gui.contexts.Notes.Items))
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
	tg.gui.contexts.Tags.Items = []models.Tag{{Name: "broken", Count: 1}}

	// filterByTag should not panic even though Search fails
	tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())
}

func TestRunQuery_WithError_NoPanic(t *testing.T) {
	mock := testutil.NewMockExecutor().
		WithQueries(models.Query{Name: "broken", Query: "#fail"}).
		WithError(fmt.Errorf("timeout"))
	tg := newTestGui(t, mock)
	defer tg.Close()

	tg.gui.contexts.Queries.Queries = []models.Query{{Name: "broken", Query: "#fail"}}

	// Should not panic
	tg.gui.helpers.Queries().RunQuery()
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
	tg.gui.pushContextByKey("searchFilter")
	tg.gui.globalController.NextPanel()
	if tg.gui.contextMgr.Current() != "notes" {
		t.Errorf("after Tab from SearchFilter: context = %v, want notes", tg.gui.contextMgr.Current())
	}
}

// --- Dialog system tests ---

func TestDeleteNote_ShowsConfirmDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Notes().DeleteNote(tg.gui.contexts.Notes.Selected())

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

	tg.gui.helpers.Notes().DeleteNote(tg.gui.contexts.Notes.Selected())

	if tg.gui.state.Dialog != nil {
		t.Error("Dialog should not be shown for empty notes")
	}
}

func TestDeleteTag_ShowsConfirmDialog(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().DeleteTag(tg.gui.contexts.Tags.Selected())

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

	tg.gui.globalController.FocusQueries()
	tg.gui.helpers.Queries().DeleteQuery()

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

	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().RenameTag(tg.gui.contexts.Tags.Selected())

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
	tg.gui.ShowConfirm("Test", "Are you sure?", func() error {
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
	tg.gui.ShowConfirm("Test", "Are you sure?", func() error {
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

	initialCount := len(tg.gui.contexts.Tags.Items)

	tg.gui.globalController.FocusTags()
	tg.gui.helpers.Tags().DeleteTag(tg.gui.contexts.Tags.Selected())

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
	if len(tg.gui.contexts.Tags.Items) != initialCount {
		t.Errorf("Tags.Items = %d, want %d (mock returns same data)", len(tg.gui.contexts.Tags.Items), initialCount)
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

	tg.gui.ShowConfirm("Test", "msg", func() error { return nil })
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
