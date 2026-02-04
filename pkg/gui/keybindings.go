package gui

import "github.com/jesseduffield/gocui"

// setupKeybindings configures all keyboard shortcuts.
func (gui *Gui) setupKeybindings() error {
	if err := gui.setupGlobalKeybindings(); err != nil {
		return err
	}
	if err := gui.setupNotesKeybindings(); err != nil {
		return err
	}
	if err := gui.setupQueriesKeybindings(); err != nil {
		return err
	}
	if err := gui.setupTagsKeybindings(); err != nil {
		return err
	}
	if err := gui.setupPreviewKeybindings(); err != nil {
		return err
	}
	if err := gui.setupSearchKeybindings(); err != nil {
		return err
	}
	if err := gui.setupDialogKeybindings(); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupGlobalKeybindings() error {
	// Quit
	if err := gui.g.SetKeybinding("", 'q', gocui.ModNone, gui.quit); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, gui.quit); err != nil {
		return err
	}

	// Panel navigation
	if err := gui.g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, gui.nextPanel); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", gocui.KeyBacktab, gocui.ModNone, gui.prevPanel); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", '1', gocui.ModNone, gui.focusNotes); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", '2', gocui.ModNone, gui.focusQueries); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", '3', gocui.ModNone, gui.focusTags); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", 'p', gocui.ModNone, gui.focusPreview); err != nil {
		return err
	}

	// Search
	if err := gui.g.SetKeybinding("", '/', gocui.ModNone, gui.openSearch); err != nil {
		return err
	}

	// Refresh
	if err := gui.g.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, gui.refresh); err != nil {
		return err
	}

	// Help
	if err := gui.g.SetKeybinding("", '?', gocui.ModNone, gui.showHelpHandler); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupNotesKeybindings() error {
	view := NotesView

	// Mouse click to focus
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.focusNotes); err != nil {
		return err
	}

	// Navigation
	if err := gui.g.SetKeybinding(view, 'j', gocui.ModNone, gui.notesDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'k', gocui.ModNone, gui.notesUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyArrowDown, gocui.ModNone, gui.notesDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyArrowUp, gocui.ModNone, gui.notesUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'g', gocui.ModNone, gui.notesTop); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'G', gocui.ModNone, gui.notesBottom); err != nil {
		return err
	}

	// Actions
	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.editNote); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'e', gocui.ModNone, gui.editNote); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'n', gocui.ModNone, gui.newNote); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'd', gocui.ModNone, gui.deleteNote); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'y', gocui.ModNone, gui.copyNotePath); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupQueriesKeybindings() error {
	view := QueriesView

	// Mouse click to focus
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.focusQueries); err != nil {
		return err
	}

	// Navigation
	if err := gui.g.SetKeybinding(view, 'j', gocui.ModNone, gui.queriesDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'k', gocui.ModNone, gui.queriesUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyArrowDown, gocui.ModNone, gui.queriesDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyArrowUp, gocui.ModNone, gui.queriesUp); err != nil {
		return err
	}

	// Actions
	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.runQuery); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'd', gocui.ModNone, gui.deleteQuery); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupTagsKeybindings() error {
	view := TagsView

	// Mouse click to focus
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.focusTags); err != nil {
		return err
	}

	// Navigation
	if err := gui.g.SetKeybinding(view, 'j', gocui.ModNone, gui.tagsDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'k', gocui.ModNone, gui.tagsUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyArrowDown, gocui.ModNone, gui.tagsDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyArrowUp, gocui.ModNone, gui.tagsUp); err != nil {
		return err
	}

	// Actions
	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.filterByTag); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'r', gocui.ModNone, gui.renameTag); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'd', gocui.ModNone, gui.deleteTag); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupPreviewKeybindings() error {
	view := PreviewView

	// Mouse click to focus
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.focusPreview); err != nil {
		return err
	}

	// Navigation
	if err := gui.g.SetKeybinding(view, 'j', gocui.ModNone, gui.previewDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'k', gocui.ModNone, gui.previewUp); err != nil {
		return err
	}

	// Actions
	if err := gui.g.SetKeybinding(view, gocui.KeyEsc, gocui.ModNone, gui.previewBack); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'f', gocui.ModNone, gui.toggleFrontmatter); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.focusNoteFromPreview); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupSearchKeybindings() error {
	view := SearchView

	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.executeSearch); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyEsc, gocui.ModNone, gui.cancelSearch); err != nil {
		return err
	}

	return nil
}
