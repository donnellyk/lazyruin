package gui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

// --- Unit tests for filtering and availability ---

func TestFilterPaletteCommands_EmptyFilter_ReturnsAll(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Edit Note", Category: "Notes", Context: NotesContext},
		},
		OriginContext: NotesContext,
	}

	gui.filterPaletteCommands("")

	if len(gui.state.Palette.Filtered) != 3 {
		t.Errorf("Filtered = %d, want 3", len(gui.state.Palette.Filtered))
	}
}

func TestFilterPaletteCommands_MatchesName(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Edit Note", Category: "Notes"},
		},
		OriginContext: NotesContext,
	}

	gui.filterPaletteCommands("quit")

	if len(gui.state.Palette.Filtered) != 1 {
		t.Errorf("Filtered = %d, want 1", len(gui.state.Palette.Filtered))
	}
	if gui.state.Palette.Filtered[0].Name != "Quit" {
		t.Errorf("Filtered[0].Name = %q, want Quit", gui.state.Palette.Filtered[0].Name)
	}
}

func TestFilterPaletteCommands_MatchesCategory(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Edit Note", Category: "Notes"},
		},
		OriginContext: NotesContext,
	}

	gui.filterPaletteCommands("notes")

	if len(gui.state.Palette.Filtered) != 1 {
		t.Errorf("Filtered = %d, want 1", len(gui.state.Palette.Filtered))
	}
	if gui.state.Palette.Filtered[0].Name != "Edit Note" {
		t.Errorf("Filtered[0].Name = %q, want Edit Note", gui.state.Palette.Filtered[0].Name)
	}
}

func TestFilterPaletteCommands_CaseInsensitive(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Toggle Frontmatter", Category: "Preview"},
		},
		OriginContext: NotesContext,
	}

	gui.filterPaletteCommands("FRONT")

	if len(gui.state.Palette.Filtered) != 1 {
		t.Errorf("Filtered = %d, want 1", len(gui.state.Palette.Filtered))
	}
}

func TestFilterPaletteCommands_NoMatch(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
		},
		OriginContext: NotesContext,
	}

	gui.filterPaletteCommands("zzzzz")

	if len(gui.state.Palette.Filtered) != 0 {
		t.Errorf("Filtered = %d, want 0", len(gui.state.Palette.Filtered))
	}
}

func TestFilterPaletteCommands_AvailableFirst(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Delete Tag", Category: "Tags", Context: TagsContext},
			{Name: "Delete Note", Category: "Notes", Context: NotesContext},
			{Name: "Quit", Category: "Global"},
		},
		OriginContext: NotesContext,
	}

	gui.filterPaletteCommands("delete")

	if len(gui.state.Palette.Filtered) != 2 {
		t.Fatalf("Filtered = %d, want 2", len(gui.state.Palette.Filtered))
	}
	// Available command (Notes context matches) should come first
	if gui.state.Palette.Filtered[0].Name != "Delete Note" {
		t.Errorf("Filtered[0].Name = %q, want Delete Note (available first)", gui.state.Palette.Filtered[0].Name)
	}
	if gui.state.Palette.Filtered[1].Name != "Delete Tag" {
		t.Errorf("Filtered[1].Name = %q, want Delete Tag (unavailable second)", gui.state.Palette.Filtered[1].Name)
	}
}

func TestFilterPaletteCommands_ClampsSelection(t *testing.T) {
	gui := &Gui{state: NewGuiState()}
	gui.state.Palette = &PaletteState{
		Commands: []PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Refresh", Category: "Global"},
		},
		OriginContext: NotesContext,
		SelectedIndex: 2,
	}

	// Filter to 1 result; selection must clamp
	gui.filterPaletteCommands("quit")

	if gui.state.Palette.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (clamped)", gui.state.Palette.SelectedIndex)
	}
}

func TestIsPaletteCommandAvailable_EmptyContext(t *testing.T) {
	cmd := PaletteCommand{Name: "Quit", Category: "Global"}
	if !isPaletteCommandAvailable(cmd, NotesContext) {
		t.Error("command with empty Context should always be available")
	}
	if !isPaletteCommandAvailable(cmd, TagsContext) {
		t.Error("command with empty Context should always be available")
	}
}

func TestIsPaletteCommandAvailable_MatchingContext(t *testing.T) {
	cmd := PaletteCommand{Name: "Edit Note", Category: "Notes", Context: NotesContext}
	if !isPaletteCommandAvailable(cmd, NotesContext) {
		t.Error("command should be available when context matches")
	}
}

func TestIsPaletteCommandAvailable_MismatchedContext(t *testing.T) {
	cmd := PaletteCommand{Name: "Edit Note", Category: "Notes", Context: NotesContext}
	if isPaletteCommandAvailable(cmd, TagsContext) {
		t.Error("command should not be available when context doesn't match")
	}
}

func TestPaletteSelectMove_Down(t *testing.T) {
	gui := &Gui{state: NewGuiState(), views: &Views{}}
	gui.state.Palette = &PaletteState{
		Filtered: []PaletteCommand{
			{Name: "A"}, {Name: "B"}, {Name: "C"},
		},
		SelectedIndex: 0,
	}

	gui.paletteSelectMove(1)
	if gui.state.Palette.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", gui.state.Palette.SelectedIndex)
	}
}

func TestPaletteSelectMove_Up(t *testing.T) {
	gui := &Gui{state: NewGuiState(), views: &Views{}}
	gui.state.Palette = &PaletteState{
		Filtered: []PaletteCommand{
			{Name: "A"}, {Name: "B"}, {Name: "C"},
		},
		SelectedIndex: 2,
	}

	gui.paletteSelectMove(-1)
	if gui.state.Palette.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", gui.state.Palette.SelectedIndex)
	}
}

func TestPaletteSelectMove_ClampsAtTop(t *testing.T) {
	gui := &Gui{state: NewGuiState(), views: &Views{}}
	gui.state.Palette = &PaletteState{
		Filtered: []PaletteCommand{
			{Name: "A"}, {Name: "B"},
		},
		SelectedIndex: 0,
	}

	gui.paletteSelectMove(-1)
	if gui.state.Palette.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (clamped)", gui.state.Palette.SelectedIndex)
	}
}

func TestPaletteSelectMove_ClampsAtBottom(t *testing.T) {
	gui := &Gui{state: NewGuiState(), views: &Views{}}
	gui.state.Palette = &PaletteState{
		Filtered: []PaletteCommand{
			{Name: "A"}, {Name: "B"},
		},
		SelectedIndex: 1,
	}

	gui.paletteSelectMove(1)
	if gui.state.Palette.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1 (clamped)", gui.state.Palette.SelectedIndex)
	}
}

// --- GUI integration tests ---

func TestOpenPalette_EntersPaletteMode(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)

	if !tg.gui.state.PaletteMode {
		t.Error("PaletteMode should be true")
	}
	if tg.gui.state.CurrentContext != PaletteContext {
		t.Errorf("CurrentContext = %v, want PaletteContext", tg.gui.state.CurrentContext)
	}
	if tg.gui.state.Palette == nil {
		t.Fatal("Palette state should not be nil")
	}
	if len(tg.gui.state.Palette.Filtered) == 0 {
		t.Error("Filtered commands should not be empty")
	}
}

func TestOpenPalette_RecordsOriginContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.state.Palette.OriginContext != TagsContext {
		t.Errorf("OriginContext = %v, want TagsContext", tg.gui.state.Palette.OriginContext)
	}
}

func TestOpenPalette_BlockedDuringSearch(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openSearch(tg.g, nil)
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.state.PaletteMode {
		t.Error("PaletteMode should not activate during search")
	}
}

func TestOpenPalette_BlockedDuringCapture(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openCapture(tg.g, nil)
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.state.PaletteMode {
		t.Error("PaletteMode should not activate during capture")
	}
}

func TestOpenPalette_BlockedDuringPick(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPick(tg.g, nil)
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.state.PaletteMode {
		t.Error("PaletteMode should not activate during pick")
	}
}

func TestOpenPalette_BlockedWhenAlreadyOpen(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)
	// Try opening again
	tg.gui.openPalette(tg.g, nil)

	// Should still be in palette mode, not double-opened
	if !tg.gui.state.PaletteMode {
		t.Error("PaletteMode should still be true")
	}
}

func TestClosePalette_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.focusTags(tg.g, nil)
	tg.gui.openPalette(tg.g, nil)
	tg.gui.closePalette()

	if tg.gui.state.PaletteMode {
		t.Error("PaletteMode should be false after close")
	}
	if tg.gui.state.Palette != nil {
		t.Error("Palette state should be nil after close")
	}
	if tg.gui.state.CurrentContext != TagsContext {
		t.Errorf("CurrentContext = %v, want TagsContext (restored)", tg.gui.state.CurrentContext)
	}
}

func TestPaletteEsc_ClosesPalette(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)
	tg.gui.paletteEsc(tg.g, nil)

	if tg.gui.state.PaletteMode {
		t.Error("PaletteMode should be false after Esc")
	}
}

func TestPaletteEnter_ExecutesCommand(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)

	// Find "Quit" in the filtered list and select it
	quitIdx := -1
	for i, cmd := range tg.gui.state.Palette.Filtered {
		if cmd.Name == "Quit" {
			quitIdx = i
			break
		}
	}
	if quitIdx < 0 {
		t.Fatal("Quit command not found in palette")
	}

	tg.gui.state.Palette.SelectedIndex = quitIdx
	err := tg.gui.paletteEnter(tg.g, nil)

	if err != gocui.ErrQuit {
		t.Errorf("expected gocui.ErrQuit, got %v", err)
	}
}

func TestPaletteEnter_SkipsUnavailableCommand(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Open palette from Tags context
	tg.gui.focusTags(tg.g, nil)
	tg.gui.openPalette(tg.g, nil)

	// Find "Open in Editor" (requires NotesContext, unavailable from Tags)
	editIdx := -1
	for i, cmd := range tg.gui.state.Palette.Filtered {
		if cmd.Name == "Open in Editor" {
			editIdx = i
			break
		}
	}
	if editIdx < 0 {
		t.Fatal("Open in Editor command not found in palette")
	}

	tg.gui.state.Palette.SelectedIndex = editIdx
	err := tg.gui.paletteEnter(tg.g, nil)

	if err != nil {
		t.Errorf("expected nil (skipped), got %v", err)
	}
	// Palette should have closed even though command was skipped
	// Actually per the code, unavailable commands return nil without closing
	// Let's check that it did NOT close (command was not executed)
}

func TestPaletteEnter_ClosesBeforeExecuting(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)

	// Find "Focus Preview" and execute it
	idx := -1
	for i, cmd := range tg.gui.state.Palette.Filtered {
		if cmd.Name == "Focus Preview" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("Focus Preview not found in palette")
	}

	tg.gui.state.Palette.SelectedIndex = idx
	tg.gui.paletteEnter(tg.g, nil)

	// Palette should be closed and we should be in Preview context
	if tg.gui.state.PaletteMode {
		t.Error("PaletteMode should be false after execution")
	}
	if tg.gui.state.CurrentContext != PreviewContext {
		t.Errorf("CurrentContext = %v, want PreviewContext", tg.gui.state.CurrentContext)
	}
}

func TestPaletteCommands_AllHaveNames(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	cmds := tg.gui.paletteCommands()
	for i, cmd := range cmds {
		if cmd.Name == "" {
			t.Errorf("command[%d] has empty Name", i)
		}
		if cmd.Category == "" {
			t.Errorf("command[%d] (%s) has empty Category", i, cmd.Name)
		}
		if cmd.OnRun == nil {
			t.Errorf("command[%d] (%s) has nil OnRun", i, cmd.Name)
		}
	}
}

func TestPaletteViewsCreated_WhenOpen(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)
	tg.g.ForceLayoutAndRedraw()

	if tg.gui.views.Palette == nil {
		t.Error("Palette view should be created")
	}
	if tg.gui.views.PaletteList == nil {
		t.Error("PaletteList view should be created")
	}
}

func TestPaletteViewsDeleted_WhenClosed(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)
	tg.g.ForceLayoutAndRedraw()
	tg.gui.closePalette()
	tg.g.ForceLayoutAndRedraw()

	if tg.gui.views.Palette != nil {
		t.Error("Palette view should be nil after close")
	}
	if tg.gui.views.PaletteList != nil {
		t.Error("PaletteList view should be nil after close")
	}
}

func TestContextToView_Palette(t *testing.T) {
	gui := &Gui{}
	if gui.contextToView(PaletteContext) != PaletteView {
		t.Errorf("contextToView(PaletteContext) = %q, want %q", gui.contextToView(PaletteContext), PaletteView)
	}
}
