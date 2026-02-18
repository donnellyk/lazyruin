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
func (gui *Gui) commands() []Command {
	return []Command{
		// Global
		{Name: "Quit", Category: "Global", Keys: []any{'q', gocui.KeyCtrlC}, Handler: gui.quit},
		{Name: "Search", Category: "Global", Keys: []any{'/'}, Handler: gui.openSearch},
		{Name: "Pick", Category: "Global", Keys: []any{'p', '\\'}, Handler: gui.openPick},
		{Name: "New Note", Category: "Global", Keys: []any{'n'}, Handler: gui.newNote},
		{Name: "Refresh", Category: "Global", Keys: []any{gocui.KeyCtrlR}, Handler: gui.refresh},
		{Name: "Keybindings", Category: "Global", Keys: []any{'?'}, Handler: gui.showHelpHandler},
		{Name: "Command Palette", Category: "Global", Keys: []any{':'}, Handler: gui.openPalette, NoPalette: true},
		{Name: "Calendar", Category: "Global", Keys: []any{'c'}, Handler: gui.openCalendar},
		{Name: "Contributions", Category: "Global", Keys: []any{'C'}, Handler: gui.openContrib},

		// Focus
		{Name: "Focus Notes", Category: "Focus", Keys: []any{'1'}, Handler: gui.focusNotes},
		{Name: "Focus Queries", Category: "Focus", Keys: []any{'2'}, Handler: gui.focusQueries},
		{Name: "Focus Tags", Category: "Focus", Keys: []any{'3'}, Handler: gui.focusTags},
		{Name: "Focus Preview", Category: "Focus", Keys: []any{}, Handler: gui.focusPreview},
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
		{Name: "View in Preview", Category: "Notes", Keys: []any{gocui.KeyEnter}, Views: []string{NotesView}, Handler: gui.viewNoteInPreview, Contexts: []ContextKey{NotesContext}},
		{Name: "Open in Editor", Category: "Notes", Keys: []any{'E'}, Views: []string{NotesView}, Handler: gui.editNote, Contexts: []ContextKey{NotesContext}},
		{Name: "Delete Note", Category: "Notes", Keys: []any{'d'}, Views: []string{NotesView}, Handler: gui.deleteNote, Contexts: []ContextKey{NotesContext}},
		{Name: "Copy Note Path", Category: "Notes", Keys: []any{'y'}, Views: []string{NotesView}, Handler: gui.copyNotePath, Contexts: []ContextKey{NotesContext}},

		// Note Actions (shared Notes + Preview)
		{Name: "Add Tag", Category: "Note Actions", Keys: []any{'t'}, Views: []string{NotesView, PreviewView}, Handler: gui.addGlobalTag, Contexts: []ContextKey{NotesContext, PreviewContext}},
		{Name: "Remove Tag", Category: "Note Actions", Keys: []any{'T'}, Views: []string{NotesView, PreviewView}, Handler: gui.removeTag, Contexts: []ContextKey{NotesContext, PreviewContext}},
		{Name: "Set Parent", Category: "Note Actions", Keys: []any{'>'}, Views: []string{NotesView, PreviewView}, Handler: gui.setParentDialog, Contexts: []ContextKey{NotesContext, PreviewContext}},
		{Name: "Remove Parent", Category: "Note Actions", Keys: []any{'P'}, Views: []string{NotesView, PreviewView}, Handler: gui.removeParent, Contexts: []ContextKey{NotesContext, PreviewContext}},
		{Name: "Toggle Bookmark", Category: "Note Actions", Keys: []any{'b'}, Views: []string{NotesView, PreviewView}, Handler: gui.toggleBookmark, Contexts: []ContextKey{NotesContext, PreviewContext}},
		{Name: "Show Info", Category: "Note Actions", Keys: []any{'s'}, Views: []string{NotesView, PreviewView}, Handler: gui.preview.showInfoDialog, Contexts: []ContextKey{NotesContext, PreviewContext}},

		// Tags — keybindings migrated to TagsController; palette entries generated from controller bindings

		// Queries
		{Name: "Run Query", Category: "Queries", Keys: []any{gocui.KeyEnter}, Views: []string{QueriesView}, Handler: gui.runQuery, Contexts: []ContextKey{QueriesContext}},
		{Name: "Delete Query", Category: "Queries", Keys: []any{'d'}, Views: []string{QueriesView}, Handler: gui.deleteQuery, Contexts: []ContextKey{QueriesContext}},

		// Preview
		{Name: "Delete Card", Category: "Preview", Keys: []any{'d'}, Views: []string{PreviewView}, Handler: gui.preview.deleteCardFromPreview, Contexts: []ContextKey{PreviewContext}},
		{Name: "Open in Editor", Category: "Preview", Keys: []any{'E'}, Views: []string{PreviewView}, Handler: gui.preview.openCardInEditor, Contexts: []ContextKey{PreviewContext}},
		{Name: "Toggle #done", Category: "Preview", Keys: []any{'D'}, Views: []string{PreviewView}, Handler: gui.preview.appendDone, Contexts: []ContextKey{PreviewContext}},
		{Name: "Move Card", Category: "Preview", Keys: []any{'m'}, Views: []string{PreviewView}, Handler: gui.preview.moveCardHandler, Contexts: []ContextKey{PreviewContext}},
		{Name: "Merge Notes", Category: "Preview", Keys: []any{'M'}, Views: []string{PreviewView}, Handler: gui.preview.mergeCardHandler, Contexts: []ContextKey{PreviewContext}},
		{Name: "Toggle Frontmatter", Category: "Preview", Keys: []any{'f'}, Views: []string{PreviewView}, Handler: gui.preview.toggleFrontmatter, Contexts: []ContextKey{PreviewContext}},
		{Name: "View Options", Category: "Preview", Keys: []any{'v'}, Views: []string{PreviewView}, Handler: gui.preview.viewOptionsDialog, Contexts: []ContextKey{PreviewContext}},
		{Name: "Toggle Inline Tag", Category: "Preview", Keys: []any{gocui.KeyCtrlT}, Views: []string{PreviewView}, Handler: gui.preview.toggleInlineTag, Contexts: []ContextKey{PreviewContext}, KeyHint: "<c-t>"},
		{Name: "Toggle Inline Date", Category: "Preview", Keys: []any{gocui.KeyCtrlD}, Views: []string{PreviewView}, Handler: gui.preview.toggleInlineDate, Contexts: []ContextKey{PreviewContext}, KeyHint: "<c-d>"},
		{Name: "Open Link", Category: "Preview", Keys: []any{'o'}, Views: []string{PreviewView}, Handler: gui.preview.openLink, Contexts: []ContextKey{PreviewContext}},
		{Name: "Toggle Todo", Category: "Preview", Keys: []any{'x'}, Views: []string{PreviewView}, Handler: gui.preview.toggleTodo, Contexts: []ContextKey{PreviewContext}},
		{Name: "Focus Note from Preview", Category: "Preview", Keys: []any{gocui.KeyEnter}, Views: []string{PreviewView}, Handler: gui.preview.focusNoteFromPreview, Contexts: []ContextKey{PreviewContext}},
		{Name: "Back", Category: "Preview", Keys: []any{gocui.KeyEsc}, Views: []string{PreviewView}, Handler: gui.preview.previewBack, NoPalette: true},
		{Name: "Go Back", Category: "Preview", Keys: []any{'['}, Views: []string{PreviewView}, Handler: gui.preview.navBack, Contexts: []ContextKey{PreviewContext}},
		{Name: "Go Forward", Category: "Preview", Keys: []any{']'}, Views: []string{PreviewView}, Handler: gui.preview.navForward, Contexts: []ContextKey{PreviewContext}},

		// Preview (palette-only)
		{Name: "Toggle Title", Category: "Preview", Contexts: []ContextKey{PreviewContext}, OnRun: gui.wrap(gui.preview.toggleTitle)},
		{Name: "Toggle Global Tags", Category: "Preview", Contexts: []ContextKey{PreviewContext}, OnRun: gui.wrap(gui.preview.toggleGlobalTags)},
		{Name: "Toggle Markdown", Category: "Preview", Contexts: []ContextKey{PreviewContext}, OnRun: gui.wrap(gui.preview.toggleMarkdown)},
		{Name: "Order Cards", Category: "Preview", Contexts: []ContextKey{PreviewContext}, OnRun: gui.preview.orderCards},
		{Name: "View History", Category: "Preview", Contexts: []ContextKey{PreviewContext}, OnRun: gui.preview.showNavHistory},

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
