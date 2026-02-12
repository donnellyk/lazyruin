package gui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jesseduffield/gocui"
)

// wrap converts a standard gocui handler into a closure for palette commands.
func (gui *Gui) wrap(fn func(*gocui.Gui, *gocui.View) error) func() error {
	return func() error {
		return fn(gui.g, nil)
	}
}

// paletteCommands builds the full list of commands available in the palette.
func (gui *Gui) paletteCommands() []PaletteCommand {
	return []PaletteCommand{
		// Focus
		{Name: "Focus Notes", Category: "Focus", Key: "1", OnRun: gui.wrap(gui.focusNotes)},
		{Name: "Focus Queries", Category: "Focus", Key: "2", OnRun: gui.wrap(gui.focusQueries)},
		{Name: "Focus Tags", Category: "Focus", Key: "3", OnRun: gui.wrap(gui.focusTags)},
		{Name: "Focus Preview", Category: "Focus", Key: "p", OnRun: gui.wrap(gui.focusPreview)},
		{Name: "Focus Search Filter", Category: "Focus", Key: "0", OnRun: gui.wrap(gui.focusSearchFilter)},

		// Tabs
		{Name: "Notes: All", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(0) }},
		{Name: "Notes: Today", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(1) }},
		{Name: "Notes: Recent", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(2) }},
		{Name: "Queries: Queries", Category: "Tabs", OnRun: func() error { return gui.switchQueriesTabByIndex(0) }},
		{Name: "Queries: Parents", Category: "Tabs", OnRun: func() error { return gui.switchQueriesTabByIndex(1) }},
		{Name: "Tags: All", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(0) }},
		{Name: "Tags: Global", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(1) }},
		{Name: "Tags: Inline", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(2) }},

		// Global
		{Name: "Search", Category: "Global", Key: "/", OnRun: gui.wrap(gui.openSearch)},
		{Name: "Pick", Category: "Global", Key: "\\", OnRun: gui.wrap(gui.openPick)},
		{Name: "New Note", Category: "Global", Key: "n", OnRun: gui.wrap(gui.newNote)},
		{Name: "Refresh", Category: "Global", Key: "ctrl+r", OnRun: gui.wrap(gui.refresh)},
		{Name: "Keybindings", Category: "Global", Key: "?", OnRun: gui.wrap(gui.showHelpHandler)},
		{Name: "Quit", Category: "Global", Key: "q", OnRun: gui.wrap(gui.quit)},

		// Notes
		{Name: "View in Preview", Category: "Notes", Key: "enter", OnRun: gui.wrap(gui.viewNoteInPreview), Context: NotesContext},
		{Name: "Open in Editor", Category: "Notes", Key: "E", OnRun: gui.wrap(gui.editNote), Context: NotesContext},
		{Name: "Delete Note", Category: "Notes", Key: "d", OnRun: gui.wrap(gui.deleteNote), Context: NotesContext},
		{Name: "Copy Note Path", Category: "Notes", Key: "y", OnRun: gui.wrap(gui.copyNotePath), Context: NotesContext},

		// Tags
		{Name: "Filter by Tag", Category: "Tags", Key: "enter", OnRun: gui.wrap(gui.filterByTag), Context: TagsContext},
		{Name: "Rename Tag", Category: "Tags", Key: "r", OnRun: gui.wrap(gui.renameTag), Context: TagsContext},
		{Name: "Delete Tag", Category: "Tags", Key: "d", OnRun: gui.wrap(gui.deleteTag), Context: TagsContext},

		// Queries
		{Name: "Run Query", Category: "Queries", Key: "enter", OnRun: gui.wrap(gui.runQuery), Context: QueriesContext},
		{Name: "Delete Query", Category: "Queries", Key: "d", OnRun: gui.wrap(gui.deleteQuery), Context: QueriesContext},

		// Preview
		{Name: "Delete Card", Category: "Preview", Key: "d", OnRun: gui.wrap(gui.deleteCardFromPreview), Context: PreviewContext},
		{Name: "Move Card", Category: "Preview", Key: "m", OnRun: gui.wrap(gui.moveCardHandler), Context: PreviewContext},
		{Name: "Toggle Frontmatter", Category: "Preview", Key: "f", OnRun: gui.wrap(gui.toggleFrontmatter), Context: PreviewContext},
		{Name: "Toggle Title", Category: "Preview", Key: "t", OnRun: gui.wrap(gui.toggleTitle), Context: PreviewContext},
		{Name: "Toggle Global Tags", Category: "Preview", Key: "T", OnRun: gui.wrap(gui.toggleGlobalTags), Context: PreviewContext},
		{Name: "Toggle Markdown", Category: "Preview", Key: "M", OnRun: gui.wrap(gui.toggleMarkdown), Context: PreviewContext},
		{Name: "Toggle Todo", Category: "Preview", Key: "x", OnRun: gui.wrap(gui.toggleTodo), Context: PreviewContext},
		{Name: "Focus Note from Preview", Category: "Preview", Key: "enter", OnRun: gui.wrap(gui.focusNoteFromPreview), Context: PreviewContext},

		// Search
		{Name: "Clear Search", Category: "Search", Key: "x", OnRun: gui.wrap(gui.clearSearch), Context: SearchFilterContext},
	}
}

// isPaletteCommandAvailable checks if a command is available given the origin context.
func isPaletteCommandAvailable(cmd PaletteCommand, origin ContextKey) bool {
	if cmd.Context == "" {
		return true
	}
	return cmd.Context == origin
}

// openPalette opens the command palette popup.
func (gui *Gui) openPalette(g *gocui.Gui, v *gocui.View) error {
	// Don't open palette during other popups
	if gui.state.SearchMode || gui.state.CaptureMode || gui.state.PickMode || gui.state.PaletteMode {
		return nil
	}

	cmds := gui.paletteCommands()
	origin := gui.state.CurrentContext

	gui.state.PaletteMode = true
	gui.state.Palette = &PaletteState{
		Commands:      cmds,
		OriginContext: origin,
	}

	gui.filterPaletteCommands("")
	gui.setContext(PaletteContext)
	return nil
}

// closePalette tears down the palette popup and restores previous context.
func (gui *Gui) closePalette() {
	if gui.state.Palette == nil {
		return
	}
	origin := gui.state.Palette.OriginContext
	gui.state.PaletteMode = false
	gui.state.Palette = nil
	gui.g.Cursor = false
	gui.setContext(origin)
}

// executePaletteCommand closes the palette and runs the selected command.
func (gui *Gui) executePaletteCommand() error {
	if gui.state.Palette == nil {
		return nil
	}
	idx := gui.state.Palette.SelectedIndex
	if idx < 0 || idx >= len(gui.state.Palette.Filtered) {
		return nil
	}

	cmd := gui.state.Palette.Filtered[idx]

	// Don't execute unavailable commands
	if !isPaletteCommandAvailable(cmd, gui.state.Palette.OriginContext) {
		return nil
	}

	// Close palette first, then execute (so commands that open popups work)
	gui.closePalette()
	return cmd.OnRun()
}

// filterPaletteCommands filters commands by a case-insensitive substring match.
func (gui *Gui) filterPaletteCommands(filter string) {
	if gui.state.Palette == nil {
		return
	}

	gui.state.Palette.FilterText = filter
	lower := strings.ToLower(filter)

	var available, unavailable []PaletteCommand
	origin := gui.state.Palette.OriginContext

	for _, cmd := range gui.state.Palette.Commands {
		match := lower == "" ||
			strings.Contains(strings.ToLower(cmd.Name), lower) ||
			strings.Contains(strings.ToLower(cmd.Category), lower)
		if !match {
			continue
		}
		if isPaletteCommandAvailable(cmd, origin) {
			available = append(available, cmd)
		} else {
			unavailable = append(unavailable, cmd)
		}
	}

	sort.Slice(available, func(i, j int) bool {
		return available[i].Name < available[j].Name
	})
	sort.Slice(unavailable, func(i, j int) bool {
		return unavailable[i].Name < unavailable[j].Name
	})

	gui.state.Palette.Filtered = append(available, unavailable...)

	// Clamp selection
	if gui.state.Palette.SelectedIndex >= len(gui.state.Palette.Filtered) {
		gui.state.Palette.SelectedIndex = max(0, len(gui.state.Palette.Filtered)-1)
	}
}

// paletteSelectMove moves the selection by delta and re-renders with scroll-to-selection.
func (gui *Gui) paletteSelectMove(delta int) {
	if gui.state.Palette == nil {
		return
	}
	next := gui.state.Palette.SelectedIndex + delta
	if next < 0 {
		next = 0
	}
	if max := len(gui.state.Palette.Filtered) - 1; next > max {
		next = max
	}
	if next < 0 {
		next = 0
	}
	gui.state.Palette.SelectedIndex = next
	gui.renderPaletteList()
	gui.scrollPaletteToSelection()
}

// scrollPaletteToSelection adjusts the viewport to keep the selected item visible.
// Called after selection or filter changes, but NOT after mouse wheel scrolling.
func (gui *Gui) scrollPaletteToSelection() {
	v := gui.views.PaletteList
	if gui.state.Palette == nil || v == nil {
		return
	}
	_, viewHeight := v.InnerSize()
	scrollListView(v, gui.state.Palette.SelectedIndex, 1, viewHeight)
}

// paletteEnter handles Enter in the palette view.
func (gui *Gui) paletteEnter(g *gocui.Gui, v *gocui.View) error {
	return gui.executePaletteCommand()
}

// paletteEsc handles Esc in the palette view.
func (gui *Gui) paletteEsc(g *gocui.Gui, v *gocui.View) error {
	gui.closePalette()
	return nil
}

// paletteListClick handles mouse clicks on the palette list.
func (gui *Gui) paletteListClick(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Palette == nil {
		return nil
	}
	idx := listClickIndex(v, 1)
	if idx >= 0 && idx < len(gui.state.Palette.Filtered) {
		gui.state.Palette.SelectedIndex = idx
		gui.renderPaletteList()
		gui.scrollPaletteToSelection()
	}
	return nil
}

// renderPaletteList writes the filtered command list to the view.
// It does NOT adjust scroll origin -- callers that change selection or filter
// should follow up with scrollPaletteToSelection(). Mouse-wheel handlers
// only call scrollViewport() and never re-render, so their origin persists.
func (gui *Gui) renderPaletteList() {
	if gui.state.Palette == nil || gui.views.PaletteList == nil {
		return
	}

	v := gui.views.PaletteList
	v.Clear()

	filtered := gui.state.Palette.Filtered
	if len(filtered) == 0 {
		fmt.Fprintln(v, " No matching commands.")
		return
	}

	originCtx := gui.state.Palette.OriginContext
	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	pad := func(s string) string {
		return s + strings.Repeat(" ", max(0, width-len([]rune(s))))
	}

	for i, cmd := range filtered {
		avail := isPaletteCommandAvailable(cmd, originCtx)

		label := fmt.Sprintf("%s: %s", cmd.Category, cmd.Name)
		keyHint := cmd.Key
		spacing := width - len(label) - len(keyHint)
		if spacing < 1 {
			spacing = 1
		}
		line := label + strings.Repeat(" ", spacing) + keyHint

		if i == gui.state.Palette.SelectedIndex {
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line), AnsiReset)
		} else if !avail {
			fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line, AnsiReset)
		} else {
			fmt.Fprintln(v, line)
		}
	}
}
