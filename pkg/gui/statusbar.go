package gui

import (
	"fmt"
	"time"

	"kvnd/lazyruin/pkg/gui/context"

	"github.com/jesseduffield/gocui"
)

// showError displays an error message in the status bar for 3 seconds, then restores it.
func (gui *Gui) ShowError(err error) {
	if gui.views.Status == nil || err == nil {
		return
	}
	gui.views.Status.Clear()
	fmt.Fprintf(gui.views.Status, " %sError: %s%s", AnsiYellow, err.Error(), AnsiReset)

	go func() {
		time.Sleep(3 * time.Second)
		gui.g.Update(func(g *gocui.Gui) error {
			gui.UpdateStatusBar()
			return nil
		})
	}()
}

func (gui *Gui) UpdateStatusBar() {
	if gui.views.Status == nil {
		return
	}

	gui.views.Status.Clear()

	for i, h := range gui.statusBarHints() {
		if i > 0 {
			fmt.Fprint(gui.views.Status, " | ")
		}
		fmt.Fprintf(gui.views.Status, "%s: %s%s%s", h.action, AnsiCyan, h.key, AnsiReset)
	}
}

// notesTabIndex returns the index for the current notes tab.
// Delegates to "notes".
func (gui *Gui) notesTabIndex() int {
	return gui.contexts.Notes.TabIndex()
}

// updateNotesTab syncs the gocui view's TabIndex with the current tab
func (gui *Gui) UpdateNotesTab() {
	if gui.views.Notes != nil {
		gui.views.Notes.TabIndex = gui.notesTabIndex()
	}
}

// queriesTabIndex returns the index for the current queries tab.
// Delegates to "queries".
func (gui *Gui) queriesTabIndex() int {
	return gui.contexts.Queries.TabIndex()
}

// updateQueriesTab syncs the gocui view's TabIndex with the current queries tab
func (gui *Gui) UpdateQueriesTab() {
	if gui.views.Queries != nil {
		gui.views.Queries.TabIndex = gui.queriesTabIndex()
	}
}

// tagsTabIndex returns the index for the current tags tab.
// Delegates to "tags".
func (gui *Gui) tagsTabIndex() int {
	return gui.contexts.Tags.TabIndex()
}

var tagsTabsNew = []context.TagsTab{context.TagsTabAll, context.TagsTabGlobal, context.TagsTabInline}

// updateTagsTab syncs the gocui view's TabIndex with the current tags tab.
func (gui *Gui) UpdateTagsTab() {
	if gui.views.Tags != nil {
		gui.views.Tags.TabIndex = gui.tagsTabIndex()
	}
}
