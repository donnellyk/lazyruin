package helpers

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
	"github.com/donnellyk/lazyruin/pkg/testutil"
)

// navTestGui extends mockGuiCommon with a recording PushContextByKey, and a
// mutable CurrentContextKey backing the preview pane. It also tracks calls
// to ReplaceContextByKey so the test can assert Back/Forward restore state.
type navTestGui struct {
	mockGuiCommon
	current       types.ContextKey
	pushed        []types.ContextKey
	replaced      []types.ContextKey
	renderPreview int
}

func (g *navTestGui) CurrentContextKey() types.ContextKey { return g.current }
func (g *navTestGui) PushContextByKey(k types.ContextKey) {
	g.pushed = append(g.pushed, k)
	g.current = k
}
func (g *navTestGui) ReplaceContextByKey(k types.ContextKey) {
	g.replaced = append(g.replaced, k)
	g.current = k
}
func (g *navTestGui) RenderPreview() { g.renderPreview++ }

func newNavigatorFixture() (*Navigator, *navTestGui, *Helpers) {
	contexts := &context.ContextTree{
		CardList:    context.NewCardListContext(),
		PickResults: context.NewPickResultsContext(),
		Compose:     context.NewComposeContext(),
		DatePreview: context.NewDatePreviewContext(),
		Notes:       context.NewNotesContext(func() {}, func() {}),
	}
	contexts.ActivePreviewKey = "cardList"
	gui := &navTestGui{
		mockGuiCommon: mockGuiCommon{contexts: contexts},
	}
	ruinCmd := commands.NewRuinCommandWithExecutor(testutil.NewMockExecutor(), "/mock")
	common := NewHelperCommon(ruinCmd, nil, gui)
	helpers := NewHelpers(common)
	return helpers.Navigator(), gui, helpers
}

// TestNavigator_NoteDeleted_RemovesHistoryAndStepsBack simulates deleting
// the note that is currently being previewed. The note's history entries
// (single-note views of it, identified by DedupID) should be removed and
// the preview should fall back to the previous history entry.
func TestNavigator_NoteDeleted_RemovesHistoryAndStepsBack(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	// Record two single-note views: first Note A, then Note B.
	_ = nav.NavigateTo("cardList", "Note A", func() error {
		cl := gui.contexts.CardList
		cl.Cards = []models.Note{{UUID: "a", Title: "Note A"}}
		cl.Source = context.CardListSource{Query: "a"}
		return nil
	})
	_ = nav.NavigateTo("cardList", "Note B", func() error {
		cl := gui.contexts.CardList
		cl.Cards = []models.Note{{UUID: "b", Title: "Note B"}}
		cl.Source = context.CardListSource{Query: "b"}
		return nil
	})

	if got := nav.Manager().Len(); got != 2 {
		t.Fatalf("precondition: history len = %d, want 2", got)
	}

	// Delete the currently-viewed note (B).
	nav.NoteDeleted("b")

	entries := nav.Manager().Entries()
	if len(entries) != 1 {
		t.Fatalf("history len after deleting current note = %d, want 1: %+v", len(entries), entries)
	}
	if entries[0].ID != "note:a" {
		t.Errorf("remaining entry ID = %q, want %q", entries[0].ID, "note:a")
	}
	if nav.Manager().Index() != 0 {
		t.Errorf("Index = %d, want 0 (restored to prior note)", nav.Manager().Index())
	}
}

// TestNavigator_NoteDeleted_RemovesNonCurrentEntries scrubs history of a
// deleted note even when it's not the current view, so Back can't take
// the user to a stale single-note view of a now-deleted note.
func TestNavigator_NoteDeleted_RemovesNonCurrentEntries(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	_ = nav.NavigateTo("cardList", "Note A", func() error {
		cl := gui.contexts.CardList
		cl.Cards = []models.Note{{UUID: "a"}}
		cl.Source = context.CardListSource{Query: "a"}
		return nil
	})
	_ = nav.NavigateTo("cardList", "Note B", func() error {
		cl := gui.contexts.CardList
		cl.Cards = []models.Note{{UUID: "b"}}
		cl.Source = context.CardListSource{Query: "b"}
		return nil
	})
	_ = nav.NavigateTo("cardList", "Note C", func() error {
		cl := gui.contexts.CardList
		cl.Cards = []models.Note{{UUID: "c"}}
		cl.Source = context.CardListSource{Query: "c"}
		return nil
	})

	// Delete Note A (not the current view — current is C).
	nav.NoteDeleted("a")

	entries := nav.Manager().Entries()
	if len(entries) != 2 {
		t.Fatalf("history len = %d, want 2", len(entries))
	}
	if entries[0].ID != "note:b" || entries[1].ID != "note:c" {
		t.Errorf("remaining IDs = [%q, %q], want [note:b, note:c]",
			entries[0].ID, entries[1].ID)
	}
	if nav.Manager().Index() != 1 {
		t.Errorf("Index = %d, want 1 (still at C)", nav.Manager().Index())
	}
}

func TestNavigator_NavigateToRecordsOneEntry(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()
	loadCalls := 0

	err := nav.NavigateTo("cardList", "First", func() error {
		loadCalls++
		gui.contexts.CardList.Cards = []models.Note{{UUID: "a"}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loadCalls != 1 {
		t.Errorf("load calls = %d, want 1", loadCalls)
	}
	if nav.Manager().Len() != 1 {
		t.Errorf("history len = %d, want 1", nav.Manager().Len())
	}
	if len(gui.pushed) != 1 || gui.pushed[0] != "cardList" {
		t.Errorf("PushContextByKey calls = %v, want [cardList]", gui.pushed)
	}
}

func TestNavigator_ShowHoverDoesNotRecord(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	for range 10 {
		err := nav.ShowHover("cardList", "hover", func() error {
			gui.contexts.CardList.Cards = []models.Note{{UUID: "x"}}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if nav.Manager().Len() != 0 {
		t.Errorf("10 hovers produced %d history entries, want 0", nav.Manager().Len())
	}
	if len(gui.pushed) != 0 {
		t.Errorf("hover pushed context %d times, want 0", len(gui.pushed))
	}
	// Hover title decorated with italics
	if title := gui.contexts.CardList.Title(); title == "" || title == "hover" {
		t.Errorf("hover title = %q, want italicized decoration", title)
	}
}

func TestNavigator_CommitHoverRecordsEntryAndStripsDecoration(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	_ = nav.ShowHover("cardList", "hover", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "x"}}
		return nil
	})
	if nav.Manager().Len() != 0 {
		t.Fatalf("precondition: history len = %d, want 0", nav.Manager().Len())
	}
	if nav.IsCurrentCommitted() {
		t.Fatal("precondition: expected hover state to be uncommitted")
	}

	nav.CommitHover()

	if !nav.IsCurrentCommitted() {
		t.Error("after CommitHover: expected currentIsCommitted = true")
	}
	if nav.Manager().Len() != 1 {
		t.Errorf("after CommitHover: history len = %d, want 1", nav.Manager().Len())
	}
	if title := gui.contexts.CardList.Title(); title != "hover" {
		t.Errorf("after CommitHover: title = %q, want %q (decoration stripped)", title, "hover")
	}
}

func TestNavigator_CommitHoverNoOpWhenAlreadyCommitted(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	_ = nav.NavigateTo("cardList", "A", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "a"}}
		gui.contexts.CardList.SetTitle("A")
		return nil
	})
	if nav.Manager().Len() != 1 {
		t.Fatalf("precondition: history len = %d, want 1", nav.Manager().Len())
	}

	nav.CommitHover()

	if nav.Manager().Len() != 1 {
		t.Errorf("after CommitHover on committed view: history len = %d, want 1 (no-op)", nav.Manager().Len())
	}
	if gui.contexts.CardList.Title() != "A" {
		t.Errorf("after CommitHover on committed view: title = %q, want %q", gui.contexts.CardList.Title(), "A")
	}
}

func TestNavigator_HoverDoesNotCorruptPriorCommittedEntry(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	// Commit view A with a specific note.
	_ = nav.NavigateTo("cardList", "A", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "a-note"}}
		gui.contexts.CardList.SetTitle("A")
		return nil
	})

	// Hover H (overwrites cardList state).
	_ = nav.ShowHover("cardList", "H", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "h-note"}}
		gui.contexts.CardList.SetTitle("H")
		return nil
	})

	// Commit B. Capture-on-departure must NOT snapshot the hover state
	// into entry A.
	_ = nav.NavigateTo("cardList", "B", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "b-note"}}
		gui.contexts.CardList.SetTitle("B")
		return nil
	})

	if nav.Manager().Len() != 2 {
		t.Fatalf("history len = %d, want 2 (A, B)", nav.Manager().Len())
	}

	// Back restores A.
	if err := nav.Back(); err != nil {
		t.Fatal(err)
	}
	if gui.contexts.CardList.Title() != "A" {
		t.Errorf("Title after Back = %q, want A (hover state should not have leaked into A)", gui.contexts.CardList.Title())
	}
	if len(gui.contexts.CardList.Cards) == 0 || gui.contexts.CardList.Cards[0].UUID != "a-note" {
		t.Errorf("Cards after Back = %+v, want [{a-note}]", gui.contexts.CardList.Cards)
	}
}

func TestNavigator_CaptureOnDepartureUpdatesCurrentEntry(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	_ = nav.NavigateTo("cardList", "A", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "a"}}
		return nil
	})

	// User interacts with A without navigating (e.g. scrolls, toggles markdown).
	// These mutations go to the live context; capture-on-departure should snapshot
	// them the next time Navigator fires.
	gui.contexts.CardList.NavState().ScrollOffset = 42
	gui.contexts.CardList.DisplayState().RenderMarkdown = false

	_ = nav.NavigateTo("cardList", "B", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "b"}}
		return nil
	})

	// Back restores A with its scroll/markdown updates preserved.
	_ = nav.Back()
	if gui.contexts.CardList.NavState().ScrollOffset != 42 {
		t.Errorf("ScrollOffset after Back = %d, want 42 (capture-on-departure)", gui.contexts.CardList.NavState().ScrollOffset)
	}
	if gui.contexts.CardList.DisplayState().RenderMarkdown != false {
		t.Error("RenderMarkdown after Back = true, want false (capture-on-departure)")
	}
}

func TestNavigator_BackForwardBounds(t *testing.T) {
	nav, _, _ := newNavigatorFixture()

	if err := nav.Back(); err != nil {
		t.Fatal(err)
	}
	if err := nav.Forward(); err != nil {
		t.Fatal(err)
	}
	// Both should be no-ops; no assertion needed beyond "does not panic or error"
}

func TestNavigator_JumpTo(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	load := func(uuid string) func() error {
		return func() error {
			gui.contexts.CardList.Cards = []models.Note{{UUID: uuid}}
			return nil
		}
	}
	_ = nav.NavigateTo("cardList", "A", load("a"))
	_ = nav.NavigateTo("cardList", "B", load("b"))
	_ = nav.NavigateTo("cardList", "C", load("c"))
	// Index now at 2 (C).

	if err := nav.JumpTo(0); err != nil {
		t.Fatalf("JumpTo(0): %v", err)
	}
	if nav.Manager().Index() != 0 {
		t.Errorf("Index = %d, want 0", nav.Manager().Index())
	}
	if got := gui.contexts.CardList.Cards[0].UUID; got != "a" {
		t.Errorf("restored UUID = %q, want a", got)
	}

	// Out-of-range jump is a no-op.
	if err := nav.JumpTo(99); err != nil {
		t.Fatalf("JumpTo(99): %v", err)
	}
	if nav.Manager().Index() != 0 {
		t.Errorf("Index after OOB jump = %d, want 0 (unchanged)", nav.Manager().Index())
	}
}

func TestNavigator_NewNavigationTruncatesForward(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	load := func(uuid string) func() error {
		return func() error {
			gui.contexts.CardList.Cards = []models.Note{{UUID: uuid}}
			return nil
		}
	}

	_ = nav.NavigateTo("cardList", "A", load("a"))
	_ = nav.NavigateTo("cardList", "B", load("b"))
	_ = nav.NavigateTo("cardList", "C", load("c"))
	_ = nav.Back() // at B
	_ = nav.Back() // at A
	_ = nav.NavigateTo("cardList", "D", load("d"))

	// History should now be [A, D] — B and C truncated.
	if nav.Manager().Len() != 2 {
		t.Errorf("history len = %d, want 2 after truncation", nav.Manager().Len())
	}
	if err := nav.Forward(); err != nil {
		t.Fatal(err)
	}
	// No forward available.
}

func TestNavigator_NotesCursorSnapOnRestore(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()

	notesCtx := gui.contexts.Notes
	notesCtx.Items = []models.Note{
		{UUID: "n1"},
		{UUID: "n2"},
		{UUID: "n3"},
	}
	notesCtx.SetSelectedLineIdx(0)

	// Navigate to card-list showing note n2.
	_ = nav.NavigateTo("cardList", "Two", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "n2"}}
		return nil
	})
	// Move side-pane cursor elsewhere.
	notesCtx.SetSelectedLineIdx(0)
	// Also commit a second view so we have something to go back to.
	_ = nav.NavigateTo("cardList", "Three", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "n3"}}
		return nil
	})
	// Back should restore selected card UUID=n2 and snap Notes cursor to idx 1.
	_ = nav.Back()
	if notesCtx.GetSelectedLineIdx() != 1 {
		t.Errorf("Notes cursor after Back = %d, want 1 (snap to n2)", notesCtx.GetSelectedLineIdx())
	}
}

func TestNavigator_ReplaceCurrentDoesNotStackContext(t *testing.T) {
	nav, gui, _ := newNavigatorFixture()
	gui.current = "cardList"

	err := nav.ReplaceCurrent("cardList", "Search", func() error {
		gui.contexts.CardList.Cards = []models.Note{{UUID: "s1"}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// PushContextByKey should NOT have been called (we're already on cardList).
	if len(gui.pushed) != 0 {
		t.Errorf("ReplaceCurrent pushed context %d times, want 0", len(gui.pushed))
	}
	if nav.Manager().Len() != 1 {
		t.Errorf("history len = %d, want 1", nav.Manager().Len())
	}
}

func TestHoverTitle_Empty(t *testing.T) {
	if HoverTitle("") != "" {
		t.Error("HoverTitle(\"\") should return empty string")
	}
}

func TestHoverTitle_NonEmpty(t *testing.T) {
	out := HoverTitle("foo")
	if out == "foo" {
		t.Error("HoverTitle should decorate the title")
	}
	if out == "" {
		t.Error("HoverTitle returned empty string for non-empty input")
	}
}
