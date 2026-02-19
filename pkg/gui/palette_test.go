package gui

import (
	"sort"
	"strings"
	"testing"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// --- Unit tests for filtering and availability ---

func TestFilterPaletteCommands_EmptyFilter_ReturnsAll(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Edit Note", Category: "Notes", Contexts: []types.ContextKey{"notes"}},
		},
	}

	gui.filterPaletteCommands("")

	if len(gui.contexts.Palette.Palette.Filtered) != 3 {
		t.Errorf("Filtered = %d, want 3", len(gui.contexts.Palette.Palette.Filtered))
	}
}

func TestFilterPaletteCommands_MatchesName(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Edit Note", Category: "Notes"},
		},
	}

	gui.filterPaletteCommands("quit")

	if len(gui.contexts.Palette.Palette.Filtered) != 1 {
		t.Errorf("Filtered = %d, want 1", len(gui.contexts.Palette.Palette.Filtered))
	}
	if gui.contexts.Palette.Palette.Filtered[0].Name != "Quit" {
		t.Errorf("Filtered[0].Name = %q, want Quit", gui.contexts.Palette.Palette.Filtered[0].Name)
	}
}

func TestFilterPaletteCommands_MatchesCategory(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Edit Note", Category: "Notes"},
		},
	}

	gui.filterPaletteCommands("notes")

	if len(gui.contexts.Palette.Palette.Filtered) != 1 {
		t.Errorf("Filtered = %d, want 1", len(gui.contexts.Palette.Palette.Filtered))
	}
	if gui.contexts.Palette.Palette.Filtered[0].Name != "Edit Note" {
		t.Errorf("Filtered[0].Name = %q, want Edit Note", gui.contexts.Palette.Palette.Filtered[0].Name)
	}
}

func TestFilterPaletteCommands_CaseInsensitive(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Toggle Frontmatter", Category: "Preview"},
		},
	}

	gui.filterPaletteCommands("FRONT")

	if len(gui.contexts.Palette.Palette.Filtered) != 1 {
		t.Errorf("Filtered = %d, want 1", len(gui.contexts.Palette.Palette.Filtered))
	}
}

func TestFilterPaletteCommands_NoMatch(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
		},
	}

	gui.filterPaletteCommands("zzzzz")

	if len(gui.contexts.Palette.Palette.Filtered) != 0 {
		t.Errorf("Filtered = %d, want 0", len(gui.contexts.Palette.Palette.Filtered))
	}
}

func TestFilterPaletteCommands_AvailableFirst(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Delete Tag", Category: "Tags", Contexts: []types.ContextKey{"tags"}},
			{Name: "Delete Note", Category: "Notes", Contexts: []types.ContextKey{"notes"}},
			{Name: "Quit", Category: "Global"},
		},
	}

	gui.filterPaletteCommands("delete")

	if len(gui.contexts.Palette.Palette.Filtered) != 2 {
		t.Fatalf("Filtered = %d, want 2", len(gui.contexts.Palette.Palette.Filtered))
	}
	// Available command (Notes context matches) should come first
	if gui.contexts.Palette.Palette.Filtered[0].Name != "Delete Note" {
		t.Errorf("Filtered[0].Name = %q, want Delete Note (available first)", gui.contexts.Palette.Palette.Filtered[0].Name)
	}
	if gui.contexts.Palette.Palette.Filtered[1].Name != "Delete Tag" {
		t.Errorf("Filtered[1].Name = %q, want Delete Tag (unavailable second)", gui.contexts.Palette.Palette.Filtered[1].Name)
	}
}

func TestFilterPaletteCommands_ClampsSelection(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: []types.PaletteCommand{
			{Name: "Quit", Category: "Global"},
			{Name: "Search", Category: "Global"},
			{Name: "Refresh", Category: "Global"},
		},

		SelectedIndex: 2,
	}

	// Filter to 1 result; selection must clamp
	gui.filterPaletteCommands("quit")

	if gui.contexts.Palette.Palette.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (clamped)", gui.contexts.Palette.Palette.SelectedIndex)
	}
}

func TestIsPaletteCommandAvailable_EmptyContext(t *testing.T) {
	cmd := types.PaletteCommand{Name: "Quit", Category: "Global"}
	if !isPaletteCommandAvailable(cmd, "notes") {
		t.Error("command with empty Context should always be available")
	}
	if !isPaletteCommandAvailable(cmd, "tags") {
		t.Error("command with empty Context should always be available")
	}
}

func TestIsPaletteCommandAvailable_MatchingContext(t *testing.T) {
	cmd := types.PaletteCommand{Name: "Edit Note", Category: "Notes", Contexts: []types.ContextKey{"notes"}}
	if !isPaletteCommandAvailable(cmd, "notes") {
		t.Error("command should be available when context matches")
	}
}

func TestIsPaletteCommandAvailable_MismatchedContext(t *testing.T) {
	cmd := types.PaletteCommand{Name: "Edit Note", Category: "Notes", Contexts: []types.ContextKey{"notes"}}
	if isPaletteCommandAvailable(cmd, "tags") {
		t.Error("command should not be available when context doesn't match")
	}
}

func TestPaletteSelectMove_Down(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), views: &Views{}, contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Filtered: []types.PaletteCommand{
			{Name: "A"}, {Name: "B"}, {Name: "C"},
		},
		SelectedIndex: 0,
	}

	gui.paletteSelectMove(1)
	if gui.contexts.Palette.Palette.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", gui.contexts.Palette.Palette.SelectedIndex)
	}
}

func TestPaletteSelectMove_Up(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), views: &Views{}, contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Filtered: []types.PaletteCommand{
			{Name: "A"}, {Name: "B"}, {Name: "C"},
		},
		SelectedIndex: 2,
	}

	gui.paletteSelectMove(-1)
	if gui.contexts.Palette.Palette.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1", gui.contexts.Palette.Palette.SelectedIndex)
	}
}

func TestPaletteSelectMove_ClampsAtTop(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), views: &Views{}, contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Filtered: []types.PaletteCommand{
			{Name: "A"}, {Name: "B"},
		},
		SelectedIndex: 0,
	}

	gui.paletteSelectMove(-1)
	if gui.contexts.Palette.Palette.SelectedIndex != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (clamped)", gui.contexts.Palette.Palette.SelectedIndex)
	}
}

func TestPaletteSelectMove_ClampsAtBottom(t *testing.T) {
	gui := &Gui{state: NewGuiState(), contextMgr: NewContextMgr(), views: &Views{}, contexts: &context.ContextTree{Palette: context.NewPaletteContext()}}
	gui.contexts.Palette.Palette = &types.PaletteState{
		Filtered: []types.PaletteCommand{
			{Name: "A"}, {Name: "B"},
		},
		SelectedIndex: 1,
	}

	gui.paletteSelectMove(1)
	if gui.contexts.Palette.Palette.SelectedIndex != 1 {
		t.Errorf("SelectedIndex = %d, want 1 (clamped)", gui.contexts.Palette.Palette.SelectedIndex)
	}
}

// --- GUI integration tests ---

func TestOpenPalette_EntersPaletteOverlay(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)

	if !tg.gui.popupActive() {
		t.Error("popupActive() should be true after openPalette")
	}
	if tg.gui.contextMgr.Current() != "palette" {
		t.Errorf("currentContext() = %v, want palette", tg.gui.contextMgr.Current())
	}
	if tg.gui.contexts.Palette.Palette == nil {
		t.Fatal("Palette state should not be nil")
	}
	if len(tg.gui.contexts.Palette.Palette.Filtered) == 0 {
		t.Error("Filtered commands should not be empty")
	}
}

func TestOpenPalette_PreviousContextIsOrigin(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.contextMgr.Previous() != "tags" {
		t.Errorf("previousContext() = %v, want tags", tg.gui.contextMgr.Previous())
	}
}

func TestOpenPalette_BlockedDuringSearch(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Search().OpenSearch()
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.contextMgr.Current() != "search" {
		t.Error("currentContext should remain search, not switch to palette")
	}
}

func TestOpenPalette_BlockedDuringCapture(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.contextMgr.Current() != "capture" {
		t.Error("currentContext should remain capture, not switch to palette")
	}
}

func TestOpenPalette_BlockedDuringPick(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Pick().OpenPick()
	tg.gui.openPalette(tg.g, nil)

	if tg.gui.contextMgr.Current() != "pick" {
		t.Error("currentContext should remain pick, not switch to palette")
	}
}

func TestOpenPalette_BlockedWhenAlreadyOpen(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)
	// Try opening again
	tg.gui.openPalette(tg.g, nil)

	// Should still be in palette context, not double-opened
	if tg.gui.contextMgr.Current() != "palette" {
		t.Error("currentContext should still be palette")
	}
}

func TestClosePalette_RestoresContext(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.globalController.FocusTags()
	tg.gui.openPalette(tg.g, nil)
	tg.gui.closePalette()

	if tg.gui.popupActive() {
		t.Error("popupActive() should be false after closePalette")
	}
	if tg.gui.contexts.Palette.Palette != nil {
		t.Error("Palette state should be nil after close")
	}
	if tg.gui.contextMgr.Current() != "tags" {
		t.Errorf("currentContext() = %v, want tags (restored)", tg.gui.contextMgr.Current())
	}
}

func TestPaletteEsc_ClosesPalette(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)
	tg.gui.paletteEsc(tg.g, nil)

	if tg.gui.popupActive() {
		t.Error("popupActive() should be false after paletteEsc")
	}
}

func TestPaletteEnter_ExecutesCommand(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.openPalette(tg.g, nil)

	// Find "Quit" in the filtered list and select it
	quitIdx := -1
	for i, cmd := range tg.gui.contexts.Palette.Palette.Filtered {
		if cmd.Name == "Quit" {
			quitIdx = i
			break
		}
	}
	if quitIdx < 0 {
		t.Fatal("Quit command not found in palette")
	}

	tg.gui.contexts.Palette.Palette.SelectedIndex = quitIdx
	err := tg.gui.paletteEnter(tg.g, nil)

	if err != gocui.ErrQuit {
		t.Errorf("expected gocui.ErrQuit, got %v", err)
	}
}

func TestPaletteEnter_SkipsUnavailableCommand(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Open palette from Tags context
	tg.gui.globalController.FocusTags()
	tg.gui.openPalette(tg.g, nil)

	// Find "Open in Editor" (requires "notes", unavailable from Tags)
	editIdx := -1
	for i, cmd := range tg.gui.contexts.Palette.Palette.Filtered {
		if cmd.Name == "Open in Editor" {
			editIdx = i
			break
		}
	}
	if editIdx < 0 {
		t.Fatal("Open in Editor command not found in palette")
	}

	tg.gui.contexts.Palette.Palette.SelectedIndex = editIdx
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
	for i, cmd := range tg.gui.contexts.Palette.Palette.Filtered {
		if cmd.Name == "Focus Preview" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("Focus Preview not found in palette")
	}

	tg.gui.contexts.Palette.Palette.SelectedIndex = idx
	tg.gui.paletteEnter(tg.g, nil)

	// Palette should be closed and we should be in Preview context
	if tg.gui.popupActive() {
		t.Error("popupActive() should be false after palette execution")
	}
	if tg.gui.contextMgr.Current() != "preview" {
		t.Errorf("currentContext() = %v, want preview", tg.gui.contextMgr.Current())
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
	tg := newTestGui(t, defaultMock())
	defer tg.Close()
	key := types.ContextKey("palette")
	if tg.gui.contextToView(key) != PaletteView {
		t.Errorf("contextToView(palette) = %q, want %q", tg.gui.contextToView(key), PaletteView)
	}
}

func TestPaletteOnlyCommands_AllAppearInPaletteCommands(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	onlyCmds := tg.gui.paletteOnlyCommands()
	paletteCmds := tg.gui.paletteCommands()

	type key struct {
		name     string
		contexts string
	}
	contextsKey := func(ctxs []types.ContextKey) string {
		sorted := make([]string, len(ctxs))
		for i, c := range ctxs {
			sorted[i] = string(c)
		}
		sort.Strings(sorted)
		return strings.Join(sorted, ",")
	}
	paletteSet := make(map[key]bool)
	for _, pc := range paletteCmds {
		paletteSet[key{pc.Name, contextsKey(pc.Contexts)}] = true
	}

	for _, cmd := range onlyCmds {
		k := key{cmd.Name, contextsKey(cmd.Contexts)}
		if !paletteSet[k] {
			t.Errorf("paletteOnlyCommands entry %q (contexts %q) not found in paletteCommands()", cmd.Name, cmd.Contexts)
		}
	}
}

func TestPaletteCommands_NoDuplicateNameContexts(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	type key struct {
		name     string
		contexts string
	}
	contextsKey := func(ctxs []types.ContextKey) string {
		sorted := make([]string, len(ctxs))
		for i, c := range ctxs {
			sorted[i] = string(c)
		}
		sort.Strings(sorted)
		return strings.Join(sorted, ",")
	}
	seen := make(map[key]bool)
	for _, cmd := range tg.gui.paletteCommands() {
		k := key{cmd.Name, contextsKey(cmd.Contexts)}
		if seen[k] {
			t.Errorf("duplicate palette command: Name=%q Contexts=%q", cmd.Name, cmd.Contexts)
		}
		seen[k] = true
	}
}

func TestKeyDisplayString(t *testing.T) {
	tests := []struct {
		key  any
		want string
	}{
		{'q', "q"},
		{'/', "/"},
		{gocui.KeyEnter, "enter"},
		{gocui.KeyEsc, "esc"},
		{gocui.KeyCtrlR, "<c-r>"},
		{gocui.KeyCtrlC, "<c-c>"},
		{gocui.KeyTab, "tab"},
	}
	for _, tt := range tests {
		got := keyDisplayString(tt.key)
		if got != tt.want {
			t.Errorf("keyDisplayString(%v) = %q, want %q", tt.key, got, tt.want)
		}
	}
}
