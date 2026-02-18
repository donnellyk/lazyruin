package gui

import (
	"fmt"
	"time"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
)

// showError displays an error message in the status bar for 3 seconds, then restores it.
func (gui *Gui) showError(err error) {
	if gui.views.Status == nil || err == nil {
		return
	}
	gui.views.Status.Clear()
	fmt.Fprintf(gui.views.Status, " %sError: %s%s", AnsiYellow, err.Error(), AnsiReset)

	go func() {
		time.Sleep(3 * time.Second)
		gui.g.Update(func(g *gocui.Gui) error {
			gui.updateStatusBar()
			return nil
		})
	}()
}

func (gui *Gui) updateStatusBar() {
	if gui.views.Status == nil {
		return
	}

	gui.views.Status.Clear()

	def := gui.contextHintDefs()
	hints := def.statusBar
	if hints == nil {
		hints = def.hints
	}

	cyan := AnsiCyan
	reset := AnsiReset
	for i, h := range hints {
		if i > 0 {
			fmt.Fprint(gui.views.Status, " | ")
		}
		fmt.Fprintf(gui.views.Status, "%s: %s%s%s", h.action, cyan, h.key, reset)
	}
}

// notesTabIndex returns the index for the current notes tab.
// Delegates to NotesContext.
func (gui *Gui) notesTabIndex() int {
	return gui.contexts.Notes.TabIndex()
}

// updateNotesTab syncs the gocui view's TabIndex with the current tab
func (gui *Gui) updateNotesTab() {
	if gui.views.Notes != nil {
		gui.views.Notes.TabIndex = gui.notesTabIndex()
	}
}

// queriesTabIndex returns the index for the current queries tab.
// Delegates to QueriesContext.
func (gui *Gui) queriesTabIndex() int {
	return gui.contexts.Queries.TabIndex()
}

// updateQueriesTab syncs the gocui view's TabIndex with the current queries tab
func (gui *Gui) updateQueriesTab() {
	if gui.views.Queries != nil {
		gui.views.Queries.TabIndex = gui.queriesTabIndex()
	}
}

// tagsTabIndex returns the index for the current tags tab.
// Delegates to TagsContext.
func (gui *Gui) tagsTabIndex() int {
	return gui.contexts.Tags.TabIndex()
}

var tagsTabsNew = []context.TagsTab{context.TagsTabAll, context.TagsTabGlobal, context.TagsTabInline}

// updateTagsTab syncs the gocui view's TabIndex with the current tags tab.
func (gui *Gui) updateTagsTab() {
	if gui.views.Tags != nil {
		gui.views.Tags.TabIndex = gui.tagsTabIndex()
	}
}
