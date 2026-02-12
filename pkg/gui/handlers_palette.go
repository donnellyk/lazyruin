package gui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jesseduffield/gocui"
)

// paletteCommands derives the palette command list from the unified command table.
func (gui *Gui) paletteCommands() []PaletteCommand {
	var cmds []PaletteCommand
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
			runner = gui.wrap(c.Handler)
		}

		cmds = append(cmds, PaletteCommand{
			Name:     c.Name,
			Category: c.Category,
			Key:      hint,
			OnRun:    runner,
			Context:  c.Context,
		})
	}
	return cmds
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
