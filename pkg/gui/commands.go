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
	View      string                              // gocui view name; "" = global
	Handler   func(*gocui.Gui, *gocui.View) error // keybinding handler
	OnRun     func() error                        // palette-only runner (when Handler is nil)
	Context   ContextKey                          // palette context filter; "" = always available
	KeyHint   string                              // display string ("<c-r>"); auto-derived if empty
	NoPalette bool                                // true = suppress from command palette
}

// commands returns the unified command table.
func (gui *Gui) commands() []Command {
	return []Command{
		// Global
		{Name: "Quit", Category: "Global", Keys: []any{'q', gocui.KeyCtrlC}, Handler: gui.quit},
		{Name: "Search", Category: "Global", Keys: []any{'/'}, Handler: gui.openSearch},
		{Name: "Pick", Category: "Global", Keys: []any{'\\'}, Handler: gui.openPick},
		{Name: "New Note", Category: "Global", Keys: []any{'n'}, Handler: gui.newNote},
		{Name: "Refresh", Category: "Global", Keys: []any{gocui.KeyCtrlR}, Handler: gui.refresh},
		{Name: "Keybindings", Category: "Global", Keys: []any{'?'}, Handler: gui.showHelpHandler},
		{Name: "Command Palette", Category: "Global", Keys: []any{':'}, Handler: gui.openPalette, NoPalette: true},

		// Focus
		{Name: "Focus Notes", Category: "Focus", Keys: []any{'1'}, Handler: gui.focusNotes},
		{Name: "Focus Queries", Category: "Focus", Keys: []any{'2'}, Handler: gui.focusQueries},
		{Name: "Focus Tags", Category: "Focus", Keys: []any{'3'}, Handler: gui.focusTags},
		{Name: "Focus Preview", Category: "Focus", Keys: []any{'p'}, Handler: gui.focusPreview},
		{Name: "Focus Search Filter", Category: "Focus", Keys: []any{'0'}, Handler: gui.focusSearchFilter},

		// Tabs (palette-only, no keybindings)
		{Name: "Notes: All", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(0) }},
		{Name: "Notes: Today", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(1) }},
		{Name: "Notes: Recent", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(2) }},
		{Name: "Queries: Queries", Category: "Tabs", OnRun: func() error { return gui.switchQueriesTabByIndex(0) }},
		{Name: "Queries: Parents", Category: "Tabs", OnRun: func() error { return gui.switchQueriesTabByIndex(1) }},
		{Name: "Tags: All", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(0) }},
		{Name: "Tags: Global", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(1) }},
		{Name: "Tags: Inline", Category: "Tabs", OnRun: func() error { return gui.switchTagsTabByIndex(2) }},

		// Notes
		{Name: "View in Preview", Category: "Notes", Keys: []any{gocui.KeyEnter}, View: NotesView, Handler: gui.viewNoteInPreview, Context: NotesContext},
		{Name: "Open in Editor", Category: "Notes", Keys: []any{'E'}, View: NotesView, Handler: gui.editNote, Context: NotesContext},
		{Name: "Delete Note", Category: "Notes", Keys: []any{'d'}, View: NotesView, Handler: gui.deleteNote, Context: NotesContext},
		{Name: "Copy Note Path", Category: "Notes", Keys: []any{'y'}, View: NotesView, Handler: gui.copyNotePath, Context: NotesContext},
		{Name: "Add Tag", Category: "Notes", Keys: []any{'t'}, View: NotesView, Handler: gui.addGlobalTag, Context: NotesContext},
		{Name: "Remove Tag", Category: "Notes", Keys: []any{'T'}, View: NotesView, Handler: gui.removeTag, Context: NotesContext},
		{Name: "Set Parent", Category: "Notes", Keys: []any{'>'}, View: NotesView, Handler: gui.setParentDialog, Context: NotesContext},
		{Name: "Remove Parent", Category: "Notes", Keys: []any{'P'}, View: NotesView, Handler: gui.removeParent, Context: NotesContext},
		{Name: "Toggle Bookmark", Category: "Notes", Keys: []any{'b'}, View: NotesView, Handler: gui.toggleBookmark, Context: NotesContext},
		{Name: "Show Info", Category: "Notes", Keys: []any{'s'}, View: NotesView, Handler: gui.showInfoDialog, Context: NotesContext},

		// Tags
		{Name: "Filter by Tag", Category: "Tags", Keys: []any{gocui.KeyEnter}, View: TagsView, Handler: gui.filterByTag, Context: TagsContext},
		{Name: "Rename Tag", Category: "Tags", Keys: []any{'r'}, View: TagsView, Handler: gui.renameTag, Context: TagsContext},
		{Name: "Delete Tag", Category: "Tags", Keys: []any{'d'}, View: TagsView, Handler: gui.deleteTag, Context: TagsContext},

		// Queries
		{Name: "Run Query", Category: "Queries", Keys: []any{gocui.KeyEnter}, View: QueriesView, Handler: gui.runQuery, Context: QueriesContext},
		{Name: "Delete Query", Category: "Queries", Keys: []any{'d'}, View: QueriesView, Handler: gui.deleteQuery, Context: QueriesContext},

		// Preview
		{Name: "Delete Card", Category: "Preview", Keys: []any{'d'}, View: PreviewView, Handler: gui.deleteCardFromPreview, Context: PreviewContext},
		{Name: "Open in Editor", Category: "Preview", Keys: []any{'E'}, View: PreviewView, Handler: gui.openCardInEditor, Context: PreviewContext},
		{Name: "Append #done", Category: "Preview", Keys: []any{'D'}, View: PreviewView, Handler: gui.appendDone, Context: PreviewContext},
		{Name: "Move Card", Category: "Preview", Keys: []any{'m'}, View: PreviewView, Handler: gui.moveCardHandler, Context: PreviewContext},
		{Name: "Merge Notes", Category: "Preview", Keys: []any{'M'}, View: PreviewView, Handler: gui.mergeCardHandler, Context: PreviewContext},
		{Name: "Toggle Frontmatter", Category: "Preview", Keys: []any{'f'}, View: PreviewView, Handler: gui.toggleFrontmatter, Context: PreviewContext},
		{Name: "View Options", Category: "Preview", Keys: []any{'v'}, View: PreviewView, Handler: gui.viewOptionsDialog, Context: PreviewContext},
		{Name: "Set Parent", Category: "Preview", Keys: []any{'>'}, View: PreviewView, Handler: gui.setParentDialog, Context: PreviewContext},
		{Name: "Remove Parent", Category: "Preview", Keys: []any{'P'}, View: PreviewView, Handler: gui.removeParent, Context: PreviewContext},
		{Name: "Add Tag", Category: "Preview", Keys: []any{'t'}, View: PreviewView, Handler: gui.addGlobalTag, Context: PreviewContext},
		{Name: "Toggle Inline Tag", Category: "Preview", Keys: []any{gocui.KeyCtrlT}, View: PreviewView, Handler: gui.toggleInlineTag, Context: PreviewContext, KeyHint: "<c-t>"},
		{Name: "Remove Tag", Category: "Preview", Keys: []any{'T'}, View: PreviewView, Handler: gui.removeTag, Context: PreviewContext},
		{Name: "Toggle Bookmark", Category: "Preview", Keys: []any{'b'}, View: PreviewView, Handler: gui.toggleBookmark, Context: PreviewContext},
		{Name: "Show Info", Category: "Preview", Keys: []any{'s'}, View: PreviewView, Handler: gui.showInfoDialog, Context: PreviewContext},
		{Name: "Open Link", Category: "Preview", Keys: []any{'o'}, View: PreviewView, Handler: gui.openLink, Context: PreviewContext},
		{Name: "Toggle Todo", Category: "Preview", Keys: []any{'x'}, View: PreviewView, Handler: gui.toggleTodo, Context: PreviewContext},
		{Name: "Focus Note from Preview", Category: "Preview", Keys: []any{gocui.KeyEnter}, View: PreviewView, Handler: gui.focusNoteFromPreview, Context: PreviewContext},
		{Name: "Back", Category: "Preview", Keys: []any{gocui.KeyEsc}, View: PreviewView, Handler: gui.previewBack, NoPalette: true},

		// Preview (palette-only)
		{Name: "Toggle Title", Category: "Preview", Context: PreviewContext, OnRun: gui.wrap(gui.toggleTitle)},
		{Name: "Toggle Global Tags", Category: "Preview", Context: PreviewContext, OnRun: gui.wrap(gui.toggleGlobalTags)},
		{Name: "Toggle Markdown", Category: "Preview", Context: PreviewContext, OnRun: gui.wrap(gui.toggleMarkdown)},
		{Name: "Order Cards", Category: "Preview", Context: PreviewContext, OnRun: gui.orderCards},

		// Search Filter
		{Name: "Clear Search", Category: "Search", Keys: []any{'x'}, View: SearchFilterView, Handler: gui.clearSearch, Context: SearchFilterContext},
	}
}

// wrap converts a standard gocui handler into a closure for palette commands.
func (gui *Gui) wrap(fn func(*gocui.Gui, *gocui.View) error) func() error {
	return func() error {
		return fn(gui.g, nil)
	}
}

// dialogActive returns true when a dialog is open.
func (gui *Gui) dialogActive() bool {
	return gui.state.Dialog != nil && gui.state.Dialog.Active
}

// suppressDuringDialog wraps a handler to no-op when a dialog is active.
func (gui *Gui) suppressDuringDialog(fn func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if gui.dialogActive() {
			return nil
		}
		return fn(g, v)
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
