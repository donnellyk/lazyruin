package gui

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/types"
)

// paletteCommands derives the palette command list from both the legacy command
// table and new controller bindings (transitional aggregator during migration).
func (gui *Gui) paletteCommands() []PaletteCommand {
	var cmds []PaletteCommand

	// Legacy commands (non-migrated panels)
	for _, c := range gui.commands() {
		if c.NoPalette || c.Name == "" {
			continue
		}

		hint := c.KeyHint
		if hint == "" && len(c.Keys) > 0 {
			hint = keyDisplayString(c.Keys[0])
		}

		var runner func() error
		if c.OnRun != nil {
			runner = c.OnRun
		} else if c.Handler != nil {
			if len(c.Views) > 0 {
				h := c.Handler
				viewName := c.Views[0]
				runner = func() error {
					v, _ := gui.g.View(viewName)
					return h(gui.g, v)
				}
			} else {
				runner = gui.wrap(c.Handler)
			}
		}

		cmds = append(cmds, PaletteCommand{
			Name:     c.Name,
			Category: c.Category,
			Key:      hint,
			OnRun:    runner,
			Contexts: c.Contexts,
		})
	}

	// New controller bindings (migrated panels: Tags)
	opts := types.KeybindingsOpts{}
	for _, ctx := range gui.contexts.All() {
		ctxKey := ctx.GetKey()
		for _, b := range ctx.GetKeybindings(opts) {
			if b.Description == "" {
				continue // nav-only, skip
			}
			cmds = append(cmds, PaletteCommand{
				Name:     b.Description,
				Category: b.Category,
				Key:      keyDisplayString(b.Key),
				OnRun:    b.Handler,
				Contexts: []ContextKey{ctxKey},
			})
		}
	}

	return cmds
}

// isPaletteCommandAvailable checks if a command is available given the origin context.
func isPaletteCommandAvailable(cmd PaletteCommand, origin ContextKey) bool {
	if len(cmd.Contexts) == 0 {
		return true
	}
	for _, ctx := range cmd.Contexts {
		if ctx == origin {
			return true
		}
	}
	return false
}

// openPalette opens the command palette popup.
func (gui *Gui) openPalette(g *gocui.Gui, v *gocui.View) error {
	if !gui.openOverlay(OverlayPalette) {
		return nil
	}

	cmds := gui.paletteCommands()

	gui.state.Palette = &PaletteState{
		Commands: cmds,
	}

	gui.filterPaletteCommands("")
	gui.pushContext(PaletteContext)
	return nil
}

// closePalette tears down the palette popup and restores previous context.
func (gui *Gui) closePalette() {
	if gui.state.Palette == nil {
		return
	}
	gui.closeOverlay()
	gui.state.PaletteSeedDone = false
	gui.state.Palette = nil
	gui.g.Cursor = false
	gui.popContext()
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
	if !isPaletteCommandAvailable(cmd, gui.state.previousContext()) {
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
	origin := gui.state.previousContext()

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

	originCtx := gui.state.previousContext()
	width, _ := v.InnerSize()
	if width < 10 {
		width = 30
	}

	const keyCol = 8

	pad := func(s string) string {
		return s + strings.Repeat(" ", max(0, width-len([]rune(s))))
	}

	for i, cmd := range filtered {
		avail := isPaletteCommandAvailable(cmd, originCtx)

		key := cmd.Key
		label := fmt.Sprintf("%s: %s", cmd.Category, cmd.Name)
		keyPad := keyCol - len(key)
		if keyPad < 1 {
			keyPad = 1
		}

		if i == gui.state.Palette.SelectedIndex {
			line := " " + key + strings.Repeat(" ", keyPad) + label
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line), AnsiReset)
		} else if !avail {
			fmt.Fprintf(v, "%s %s%-*s%s%s\n", AnsiDim, AnsiGreen, keyCol, key, AnsiReset+AnsiDim, label)
		} else {
			fmt.Fprintf(v, " %s%-*s%s%s\n", AnsiGreen, keyCol, key, AnsiReset, label)
		}
	}
}

// quickOpenItems builds PaletteCommand entries from all navigable items in rank order.
func (gui *Gui) quickOpenItems() []PaletteCommand {
	var items []PaletteCommand

	// Saved queries
	for i, q := range gui.state.Queries.Items {
		idx := i
		query := q
		items = append(items, PaletteCommand{
			Name:     query.Name,
			Category: "Query",
			OnRun: func() error {
				gui.state.Queries.CurrentTab = QueriesTabQueries
				gui.state.Queries.SelectedIndex = idx
				gui.setContext(QueriesContext)
				gui.renderQueries()
				return gui.runQuery(nil, nil)
			},
		})
	}

	// Bookmark parents
	for i := range gui.state.Parents.Items {
		idx := i
		items = append(items, PaletteCommand{
			Name:     gui.state.Parents.Items[idx].Name,
			Category: "Parent",
			OnRun: func() error {
				gui.state.Queries.CurrentTab = QueriesTabParents
				gui.state.Parents.SelectedIndex = idx
				gui.setContext(QueriesContext)
				gui.renderQueries()
				return gui.viewParent(nil, nil)
			},
		})
	}

	// Tags
	for _, t := range gui.state.Tags.Items {
		tag := t
		name := "#" + tag.Name
		if slices.Contains(tag.Scope, "inline") {
			items = append(items, PaletteCommand{
				Name:     name,
				Category: "Tag",
				OnRun: func() error {
					gui.setContext(TagsContext)
					return gui.filterByTagPick(&tag)
				},
			})
		} else {
			items = append(items, PaletteCommand{
				Name:     name,
				Category: "Tag",
				OnRun: func() error {
					gui.setContext(TagsContext)
					return gui.filterByTagSearch(&tag)
				},
			})
		}
	}

	// Notes (deduplicated by title)
	seen := make(map[string]bool)
	for i, n := range gui.state.Notes.Items {
		idx := i
		if seen[n.Title] {
			continue
		}
		seen[n.Title] = true
		items = append(items, PaletteCommand{
			Name:     n.Title,
			Category: "Note",
			OnRun: func() error {
				gui.state.Notes.SelectedIndex = idx
				gui.setContext(NotesContext)
				gui.renderNotes()
				gui.preview.updatePreviewForNotes()
				return nil
			},
		})
	}

	return items
}

// filterQuickOpenItems filters Quick Open items by case-insensitive substring match.
func (gui *Gui) filterQuickOpenItems(filter string) {
	if gui.state.Palette == nil {
		return
	}

	gui.state.Palette.FilterText = filter
	lower := strings.ToLower(filter)

	var filtered []PaletteCommand
	for _, item := range gui.quickOpenItems() {
		if lower == "" ||
			strings.Contains(strings.ToLower(item.Name), lower) ||
			strings.Contains(strings.ToLower(item.Category), lower) {
			filtered = append(filtered, item)
		}
	}

	gui.state.Palette.Filtered = filtered

	// Clamp selection
	if gui.state.Palette.SelectedIndex >= len(gui.state.Palette.Filtered) {
		gui.state.Palette.SelectedIndex = max(0, len(gui.state.Palette.Filtered)-1)
	}
}
