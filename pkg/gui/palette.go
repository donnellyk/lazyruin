package gui

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// paletteCommands builds the full palette command list from controller bindings
// and palette-only commands (tabs, snippets, etc. without a controller home).
func (gui *Gui) paletteCommands() []types.PaletteCommand {
	var cmds []types.PaletteCommand

	// Palette-only commands (no controller home)
	cmds = append(cmds, gui.paletteOnlyCommands()...)

	// Controller bindings
	opts := types.KeybindingsOpts{}
	for _, ctx := range gui.contexts.All() {
		ctxKey := ctx.GetKey()
		isGlobal := ctx.GetKind() == types.GLOBAL_CONTEXT
		for _, b := range ctx.GetKeybindings(opts) {
			if b.Description == "" {
				continue // nav-only, skip
			}
			keyHint := ""
			if b.Key != nil {
				keyHint = keyDisplayString(b.Key)
			}
			// Global context bindings are available regardless of active context.
			var contexts []types.ContextKey
			if !isGlobal {
				contexts = []types.ContextKey{ctxKey}
			}
			cmds = append(cmds, types.PaletteCommand{
				Name:     b.Description,
				Category: b.Category,
				Key:      keyHint,
				OnRun:    b.Handler,
				Contexts: contexts,
			})
		}
	}

	return cmds
}

// isPaletteCommandAvailable checks if a command is available given the origin context.
func isPaletteCommandAvailable(cmd types.PaletteCommand, origin types.ContextKey) bool {
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
	if gui.popupActive() {
		return nil
	}

	cmds := gui.paletteCommands()

	gui.contexts.Palette.Palette = &types.PaletteState{
		Commands: cmds,
	}

	gui.filterPaletteCommands("")
	gui.pushContextByKey("palette")
	return nil
}

// closePalette tears down the palette popup and restores previous context.
func (gui *Gui) closePalette() {
	if gui.contexts.Palette.Palette == nil {
		return
	}
	gui.contexts.Palette.SeedDone = false
	gui.contexts.Palette.Palette = nil
	gui.g.Cursor = false
	gui.popContext()
}

// executePaletteCommand closes the palette and runs the selected command.
func (gui *Gui) executePaletteCommand() error {
	if gui.contexts.Palette.Palette == nil {
		return nil
	}
	idx := gui.contexts.Palette.Palette.SelectedIndex
	if idx < 0 || idx >= len(gui.contexts.Palette.Palette.Filtered) {
		return nil
	}

	cmd := gui.contexts.Palette.Palette.Filtered[idx]

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
	if gui.contexts.Palette.Palette == nil {
		return
	}

	gui.contexts.Palette.Palette.FilterText = filter
	lower := strings.ToLower(filter)

	var available, unavailable []types.PaletteCommand
	origin := gui.state.previousContext()

	for _, cmd := range gui.contexts.Palette.Palette.Commands {
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

	gui.contexts.Palette.Palette.Filtered = append(available, unavailable...)

	// Clamp selection
	if gui.contexts.Palette.Palette.SelectedIndex >= len(gui.contexts.Palette.Palette.Filtered) {
		gui.contexts.Palette.Palette.SelectedIndex = max(0, len(gui.contexts.Palette.Palette.Filtered)-1)
	}
}

// paletteSelectMove moves the selection by delta and re-renders with scroll-to-selection.
func (gui *Gui) paletteSelectMove(delta int) {
	if gui.contexts.Palette.Palette == nil {
		return
	}
	next := gui.contexts.Palette.Palette.SelectedIndex + delta
	if next < 0 {
		next = 0
	}
	if max := len(gui.contexts.Palette.Palette.Filtered) - 1; next > max {
		next = max
	}
	if next < 0 {
		next = 0
	}
	gui.contexts.Palette.Palette.SelectedIndex = next
	gui.renderPaletteList()
	gui.scrollPaletteToSelection()
}

// scrollPaletteToSelection adjusts the viewport to keep the selected item visible.
// Called after selection or filter changes, but NOT after mouse wheel scrolling.
func (gui *Gui) scrollPaletteToSelection() {
	v := gui.views.PaletteList
	if gui.contexts.Palette.Palette == nil || v == nil {
		return
	}
	_, viewHeight := v.InnerSize()
	scrollListView(v, gui.contexts.Palette.Palette.SelectedIndex, 1, viewHeight)
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
	if gui.contexts.Palette.Palette == nil {
		return nil
	}
	idx := listClickIndex(v, 1)
	if idx >= 0 && idx < len(gui.contexts.Palette.Palette.Filtered) {
		gui.contexts.Palette.Palette.SelectedIndex = idx
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
	if gui.contexts.Palette.Palette == nil || gui.views.PaletteList == nil {
		return
	}

	v := gui.views.PaletteList
	v.Clear()

	filtered := gui.contexts.Palette.Palette.Filtered
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

		if i == gui.contexts.Palette.Palette.SelectedIndex {
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
func (gui *Gui) quickOpenItems() []types.PaletteCommand {
	var items []types.PaletteCommand

	// Saved queries
	for i, q := range gui.contexts.Queries.Queries {
		idx := i
		query := q
		items = append(items, types.PaletteCommand{
			Name:     query.Name,
			Category: "Query",
			OnRun: func() error {
				gui.contexts.Queries.CurrentTab = context.QueriesTabQueries
				gui.contexts.Queries.QueriesTrait().SetSelectedLineIdx(idx)
				gui.pushContextByKey("queries")
				gui.RenderQueries()
				return gui.helpers.Queries().RunQuery()
			},
		})
	}

	// Bookmark parents
	for i := range gui.contexts.Queries.Parents {
		idx := i
		items = append(items, types.PaletteCommand{
			Name:     gui.contexts.Queries.Parents[idx].Name,
			Category: "Parent",
			OnRun: func() error {
				gui.contexts.Queries.CurrentTab = context.QueriesTabParents
				gui.contexts.Queries.ParentsTrait().SetSelectedLineIdx(idx)
				gui.pushContextByKey("queries")
				gui.RenderQueries()
				return gui.helpers.Queries().ViewParent()
			},
		})
	}

	// Tags
	for _, t := range gui.contexts.Tags.Items {
		tag := t
		name := "#" + tag.Name
		if slices.Contains(tag.Scope, "inline") {
			items = append(items, types.PaletteCommand{
				Name:     name,
				Category: "Tag",
				OnRun: func() error {
					gui.pushContextByKey("tags")
					return gui.helpers.Tags().FilterByTagPick(&tag)
				},
			})
		} else {
			items = append(items, types.PaletteCommand{
				Name:     name,
				Category: "Tag",
				OnRun: func() error {
					gui.pushContextByKey("tags")
					return gui.helpers.Tags().FilterByTagSearch(&tag)
				},
			})
		}
	}

	// Notes (deduplicated by title)
	seen := make(map[string]bool)
	for i, n := range gui.contexts.Notes.Items {
		idx := i
		if seen[n.Title] {
			continue
		}
		seen[n.Title] = true
		items = append(items, types.PaletteCommand{
			Name:     n.Title,
			Category: "Note",
			OnRun: func() error {
				gui.contexts.Notes.SetSelectedLineIdx(idx)
				gui.pushContextByKey("notes")
				gui.RenderNotes()
				gui.helpers.Preview().UpdatePreviewForNotes()
				return nil
			},
		})
	}

	return items
}

// filterQuickOpenItems filters Quick Open items by case-insensitive substring match.
func (gui *Gui) filterQuickOpenItems(filter string) {
	if gui.contexts.Palette.Palette == nil {
		return
	}

	gui.contexts.Palette.Palette.FilterText = filter
	lower := strings.ToLower(filter)

	var filtered []types.PaletteCommand
	for _, item := range gui.quickOpenItems() {
		if lower == "" ||
			strings.Contains(strings.ToLower(item.Name), lower) ||
			strings.Contains(strings.ToLower(item.Category), lower) {
			filtered = append(filtered, item)
		}
	}

	gui.contexts.Palette.Palette.Filtered = filtered

	// Clamp selection
	if gui.contexts.Palette.Palette.SelectedIndex >= len(gui.contexts.Palette.Palette.Filtered) {
		gui.contexts.Palette.Palette.SelectedIndex = max(0, len(gui.contexts.Palette.Palette.Filtered)-1)
	}
}
