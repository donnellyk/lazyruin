package gui

import (
	"fmt"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// paletteOnlyCommands returns commands that have no controller home.
// These are accessible only via the command palette (no keybinding).
func (gui *Gui) paletteOnlyCommands() []types.PaletteCommand {
	return []types.PaletteCommand{
		// Tab switching (palette-only, no keybinding)
		{Name: "Notes: All", Category: "Tabs", OnRun: func() error { return gui.helpers.Notes().SwitchNotesTabByIndex(0) }},
		{Name: "Notes: Today", Category: "Tabs", OnRun: func() error { return gui.helpers.Notes().SwitchNotesTabByIndex(1) }},
		{Name: "Notes: Recent", Category: "Tabs", OnRun: func() error { return gui.helpers.Notes().SwitchNotesTabByIndex(2) }},
		{Name: "Queries: Queries", Category: "Tabs", OnRun: func() error { return gui.helpers.Queries().SwitchQueriesTabByIndex(0) }},
		{Name: "Queries: Parents", Category: "Tabs", OnRun: func() error { return gui.helpers.Queries().SwitchQueriesTabByIndex(1) }},
		{Name: "Tags: All", Category: "Tabs", OnRun: func() error { return gui.helpers.Tags().SwitchTagsTabByIndex(0) }},
		{Name: "Tags: Global", Category: "Tabs", OnRun: func() error { return gui.helpers.Tags().SwitchTagsTabByIndex(1) }},
		{Name: "Tags: Inline", Category: "Tabs", OnRun: func() error { return gui.helpers.Tags().SwitchTagsTabByIndex(2) }},

		// Search filter (palette-only; keybinding registered in setupKeybindings)
		{Name: "Clear Search", Category: "Search", Key: "x", Contexts: []types.ContextKey{"searchFilter"}, OnRun: func() error {
			gui.helpers.Search().ClearSearch()
			return nil
		}},

		// Snippets (palette-only, no keybinding)
		{Name: "List Snippets", Category: "Snippets", OnRun: func() error { return gui.helpers.Snippet().ListSnippets() }},
		{Name: "Create Snippet", Category: "Snippets", OnRun: func() error { return gui.helpers.Snippet().CreateSnippet() }},
		{Name: "Delete Snippet", Category: "Snippets", OnRun: func() error { return gui.helpers.Snippet().DeleteSnippet() }},
	}
}

// suppressDuringDialog wraps a handler to no-op when a dialog is active.
func (gui *Gui) suppressDuringDialog(fn func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if gui.overlayActive() {
			return nil
		}
		return fn(g, v)
	}
}

// suppressTabClickDuringDialog wraps a tab-click handler to no-op when a dialog is active.
func (gui *Gui) suppressTabClickDuringDialog(fn func(int) error) func(int) error {
	return func(tabIndex int) error {
		if gui.overlayActive() {
			return nil
		}
		return fn(tabIndex)
	}
}

// keyNames maps special gocui keys to display strings.
var keyNames = map[gocui.Key]string{
	gocui.KeyEnter:      "enter",
	gocui.KeyEsc:        "esc",
	gocui.KeyTab:        "tab",
	gocui.KeyBacktab:    "backtab",
	gocui.KeySpace:      "space",
	gocui.KeyBackspace:  "backspace",
	gocui.KeyDelete:     "delete",
	gocui.KeyArrowUp:    "up",
	gocui.KeyArrowDown:  "down",
	gocui.KeyArrowLeft:  "left",
	gocui.KeyArrowRight: "right",
}

// ctrlKeyNames maps ctrl+letter gocui keys to display strings.
var ctrlKeyNames = map[gocui.Key]string{
	gocui.KeyCtrlA: "<c-a>", gocui.KeyCtrlB: "<c-b>", gocui.KeyCtrlC: "<c-c>",
	gocui.KeyCtrlD: "<c-d>", gocui.KeyCtrlE: "<c-e>", gocui.KeyCtrlF: "<c-f>",
	gocui.KeyCtrlG: "<c-g>", gocui.KeyCtrlH: "<c-h>", gocui.KeyCtrlJ: "<c-j>",
	gocui.KeyCtrlK: "<c-k>", gocui.KeyCtrlL: "<c-l>", gocui.KeyCtrlN: "<c-n>",
	gocui.KeyCtrlO: "<c-o>", gocui.KeyCtrlP: "<c-p>", gocui.KeyCtrlQ: "<c-q>",
	gocui.KeyCtrlR: "<c-r>", gocui.KeyCtrlS: "<c-s>", gocui.KeyCtrlT: "<c-t>",
	gocui.KeyCtrlU: "<c-u>", gocui.KeyCtrlV: "<c-v>", gocui.KeyCtrlW: "<c-w>",
	gocui.KeyCtrlX: "<c-x>", gocui.KeyCtrlY: "<c-y>", gocui.KeyCtrlZ: "<c-z>",
}

// keyDisplayString returns a human-readable display string for a gocui key.
func keyDisplayString(key any) string {
	switch k := key.(type) {
	case rune:
		return string(k)
	case gocui.Key:
		if name, ok := keyNames[k]; ok {
			return name
		}
		if name, ok := ctrlKeyNames[k]; ok {
			return name
		}
		return fmt.Sprintf("key(%d)", k)
	default:
		return fmt.Sprintf("%v", key)
	}
}
