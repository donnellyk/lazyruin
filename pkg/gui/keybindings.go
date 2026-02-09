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
	if err := gui.setupSearchFilterKeybindings(); err != nil {
		return err
	}
	if err := gui.setupCaptureKeybindings(); err != nil {
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
	if err := gui.g.SetKeybinding("", '0', gocui.ModNone, gui.focusSearchFilter); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", 'p', gocui.ModNone, gui.focusPreview); err != nil {
		return err
	}

	// Search
	if err := gui.g.SetKeybinding("", '/', gocui.ModNone, gui.openSearch); err != nil {
		return err
	}

	// New note
	if err := gui.g.SetKeybinding("", 'n', gocui.ModNone, gui.newNote); err != nil {
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

	// Mouse wheel scrolling (only acts on preview when hovering over it)
	if err := gui.g.SetKeybinding("", gocui.MouseWheelDown, gocui.ModNone, gui.previewScrollDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", gocui.MouseWheelUp, gocui.ModNone, gui.previewScrollUp); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupNotesKeybindings() error {
	view := NotesView

	// Mouse click to focus and select
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.notesClick); err != nil {
		return err
	}

	// Tab click to switch tabs
	if err := gui.g.SetTabClickBinding(view, gui.switchNotesTabByIndex); err != nil {
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
	if err := gui.g.SetKeybinding(view, 'E', gocui.ModNone, gui.editNotesInPreview); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'd', gocui.ModNone, gui.deleteNote); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'y', gocui.ModNone, gui.copyNotePath); err != nil {
		return err
	}

	// Mouse wheel scrolls viewport (selection-aware)
	if err := gui.g.SetKeybinding(view, gocui.MouseWheelDown, gocui.ModNone, gui.notesWheelDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.MouseWheelUp, gocui.ModNone, gui.notesWheelUp); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupQueriesKeybindings() error {
	view := QueriesView

	// Mouse click to focus and select
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.queriesClick); err != nil {
		return err
	}

	// Tab click to switch tabs
	if err := gui.g.SetTabClickBinding(view, gui.switchQueriesTabByIndex); err != nil {
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

	// Mouse wheel scrolls viewport (selection-aware)
	if err := gui.g.SetKeybinding(view, gocui.MouseWheelDown, gocui.ModNone, gui.queriesWheelDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.MouseWheelUp, gocui.ModNone, gui.queriesWheelUp); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupTagsKeybindings() error {
	view := TagsView

	// Mouse click to focus and select
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.tagsClick); err != nil {
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

	// Mouse wheel scrolls viewport (selection-aware)
	if err := gui.g.SetKeybinding(view, gocui.MouseWheelDown, gocui.ModNone, gui.tagsWheelDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.MouseWheelUp, gocui.ModNone, gui.tagsWheelUp); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupPreviewKeybindings() error {
	view := PreviewView

	// Mouse click to focus and select card
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.previewClick); err != nil {
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
	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.focusNoteFromPreview); err != nil {
		return err
	}

	// Edit mode actions (guarded by EditMode in handlers)
	if err := gui.g.SetKeybinding(view, 'd', gocui.ModNone, gui.deleteCardFromPreview); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'J', gocui.ModNone, gui.moveCardDown); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'K', gocui.ModNone, gui.moveCardUp); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'm', gocui.ModNone, gui.mergeCardHandler); err != nil {
		return err
	}

	// Display toggles
	if err := gui.g.SetKeybinding(view, 'f', gocui.ModNone, gui.toggleFrontmatter); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 't', gocui.ModNone, gui.toggleTitle); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, 'T', gocui.ModNone, gui.toggleGlobalTags); err != nil {
		return err
	}
	return nil
}

func (gui *Gui) setupSearchKeybindings() error {
	view := SearchView

	if err := gui.g.SetKeybinding(view, gocui.KeyEnter, gocui.ModNone, gui.searchEnter); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyEsc, gocui.ModNone, gui.searchEsc); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyTab, gocui.ModNone, gui.searchTab); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupCaptureKeybindings() error {
	view := CaptureView

	if err := gui.g.SetKeybinding(view, gocui.KeyCtrlS, gocui.ModNone, gui.submitCapture); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyEsc, gocui.ModNone, gui.cancelCapture); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding(view, gocui.KeyTab, gocui.ModNone, gui.captureTab); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) setupSearchFilterKeybindings() error {
	view := SearchFilterView

	// Mouse click clears search
	if err := gui.g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, gui.clearSearch); err != nil {
		return err
	}

	// Clear search with x
	if err := gui.g.SetKeybinding(view, 'x', gocui.ModNone, gui.clearSearch); err != nil {
		return err
	}

	return nil
}
