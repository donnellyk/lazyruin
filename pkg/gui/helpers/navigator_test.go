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
