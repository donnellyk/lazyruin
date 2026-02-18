package gui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
)

// Command is the single source of truth for user-facing actions.
// It drives both keybinding registration and palette command generation.
type Command struct {
	Name      string                              // palette/hint display name
	Category  string                              // palette grouping ("Notes", "Global", etc.)
	Keys      []any                               // gocui keys to bind; nil = palette-only
	Views     []string                            // gocui view names to bind; nil = global
	Handler   func(*gocui.Gui, *gocui.View) error // keybinding handler
	OnRun     func() error                        // palette-only runner (when Handler is nil)
	Contexts  []ContextKey                        // palette context filter; nil = always available
	KeyHint   string                              // display string ("<c-r>"); auto-derived if empty
	NoPalette bool                                // true = suppress from command palette
}

// commands returns the unified command table.
// Global, Focus, Notes, Tags, Queries, and Preview entries have been migrated
// to their respective controllers; palette entries are generated from controller bindings.
func (gui *Gui) commands() []Command {
	return []Command{
		// Tabs (palette-only, no keybindings)
		{Name: "Notes: All", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(0) }},
		{Name: "Notes: Today", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(1) }},
		{Name: "Notes: Recent", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(2) }},
		{Name: "Queries: Queries", Category: "Tabs", OnRun: func() error { return gui.switchQueriesTabByIndex(0) }},
		{Name: "Queries: Parents", Category: "Tabs", OnRun: func() error { return gui.switchQueriesTabByIndex(1) }},
		{Name: "Tags: All", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(0) }},
		{Name: "Tags: Global", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(1) }},
		{Name: "Tags: Inline", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(2) }},

		// Search Filter
		{Name: "Clear Search", Category: "Search", Keys: []any{'x'}, Views: []string{SearchFilterView}, Handler: gui.clearSearch, Contexts: []ContextKey{SearchFilterContext}},

		// Snippets
		{Name: "List Snippets", Category: "Snippets", OnRun: gui.listSnippets},
		{Name: "Create Snippet", Category: "Snippets", OnRun: gui.createSnippet},
		{Name: "Delete Snippet", Category: "Snippets", OnRun: gui.deleteSnippet},
	}
}

// wrap converts a standard gocui handler into a closure for palette commands.
func (gui *Gui) wrap(fn func(*gocui.Gui, *gocui.View) error) func() error {
	return func() error {
		return fn(gui.g, nil)
	}
}

// dialogActive returns true when a dialog or overlay is open.
func (gui *Gui) dialogActive() bool {
	return gui.overlayActive()
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

// isMainPanelView returns true for the four main panel views
// whose interactions should be suppressed when a dialog is open.
func isMainPanelView(view string) bool {
	switch view {
	case NotesView, QueriesView, TagsView, PreviewView:
		return true
	}
	return false
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
