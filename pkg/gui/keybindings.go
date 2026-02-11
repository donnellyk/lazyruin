package gui

import "github.com/jesseduffield/gocui"

type binding struct {
	view    string
	key     any
	handler func(*gocui.Gui, *gocui.View) error
}

// registerBindings registers a slice of keybindings.
func (gui *Gui) registerBindings(bindings []binding) error {
	for _, b := range bindings {
		if err := gui.g.SetKeybinding(b.view, b.key, gocui.ModNone, b.handler); err != nil {
			return err
		}
	}
	return nil
}

// setupKeybindings configures all keyboard shortcuts.
func (gui *Gui) setupKeybindings() error {
	if err := gui.registerBindings(gui.globalBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.notesBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.queriesBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.tagsBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.previewBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.searchBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.captureBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.searchFilterBindings()); err != nil {
		return err
	}
	if err := gui.registerBindings(gui.pickBindings()); err != nil {
		return err
	}
	if err := gui.setupDialogKeybindings(); err != nil {
		return err
	}

	// Tab click bindings (different signature, can't be table-driven)
	if err := gui.g.SetTabClickBinding(NotesView, gui.switchNotesTabByIndex); err != nil {
		return err
	}
	if err := gui.g.SetTabClickBinding(QueriesView, gui.switchQueriesTabByIndex); err != nil {
		return err
	}
	if err := gui.g.SetTabClickBinding(TagsView, gui.switchTagsTabByIndex); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) globalBindings() []binding {
	return []binding{
		{"", 'q', gui.quit},
		{"", gocui.KeyCtrlC, gui.quit},
		{"", gocui.KeyTab, gui.nextPanel},
		{"", gocui.KeyBacktab, gui.prevPanel},
		{"", '1', gui.focusNotes},
		{"", '2', gui.focusQueries},
		{"", '3', gui.focusTags},
		{"", '0', gui.focusSearchFilter},
		{"", 'p', gui.focusPreview},
		{"", '/', gui.openSearch},
		{"", 'n', gui.newNote},
		{"", gocui.KeyCtrlR, gui.refresh},
		{"", '?', gui.showHelpHandler},
		{"", gocui.MouseWheelDown, gui.previewScrollDown},
		{"", gocui.MouseWheelUp, gui.previewScrollUp},
		{"", '\\', gui.openPick},
	}
}

func (gui *Gui) notesBindings() []binding {
	v := NotesView
	return []binding{
		{v, gocui.MouseLeft, gui.notesClick},
		{v, 'j', gui.notesDown},
		{v, 'k', gui.notesUp},
		{v, gocui.KeyArrowDown, gui.notesDown},
		{v, gocui.KeyArrowUp, gui.notesUp},
		{v, 'g', gui.notesTop},
		{v, 'G', gui.notesBottom},
		{v, gocui.KeyEnter, gui.editNote},
		{v, 'e', gui.editNote},
		{v, 'E', gui.editNotesInPreview},
		{v, 'd', gui.deleteNote},
		{v, 'y', gui.copyNotePath},
		{v, gocui.MouseWheelDown, gui.notesWheelDown},
		{v, gocui.MouseWheelUp, gui.notesWheelUp},
	}
}

func (gui *Gui) queriesBindings() []binding {
	v := QueriesView
	return []binding{
		{v, gocui.MouseLeft, gui.queriesClick},
		{v, 'j', gui.queriesDown},
		{v, 'k', gui.queriesUp},
		{v, gocui.KeyArrowDown, gui.queriesDown},
		{v, gocui.KeyArrowUp, gui.queriesUp},
		{v, gocui.KeyEnter, gui.runQuery},
		{v, 'd', gui.deleteQuery},
		{v, gocui.MouseWheelDown, gui.queriesWheelDown},
		{v, gocui.MouseWheelUp, gui.queriesWheelUp},
	}
}

func (gui *Gui) tagsBindings() []binding {
	v := TagsView
	return []binding{
		{v, gocui.MouseLeft, gui.tagsClick},
		{v, 'j', gui.tagsDown},
		{v, 'k', gui.tagsUp},
		{v, gocui.KeyArrowDown, gui.tagsDown},
		{v, gocui.KeyArrowUp, gui.tagsUp},
		{v, gocui.KeyEnter, gui.filterByTag},
		{v, 'r', gui.renameTag},
		{v, 'd', gui.deleteTag},
		{v, gocui.MouseWheelDown, gui.tagsWheelDown},
		{v, gocui.MouseWheelUp, gui.tagsWheelUp},
	}
}

func (gui *Gui) previewBindings() []binding {
	v := PreviewView
	return []binding{
		{v, gocui.MouseLeft, gui.previewClick},
		{v, 'j', gui.previewDown},
		{v, 'k', gui.previewUp},
		{v, gocui.KeyEsc, gui.previewBack},
		{v, gocui.KeyEnter, gui.focusNoteFromPreview},
		{v, 'd', gui.deleteCardFromPreview},
		{v, 'm', gui.moveCardHandler},
		{v, 'M', gui.mergeCardHandler},
		{v, 'f', gui.toggleFrontmatter},
		{v, 't', gui.toggleTitle},
		{v, 'T', gui.toggleGlobalTags},
		{v, 'M', gui.toggleMarkdown}, // overwrites mergeCardHandler; same as original
	}
}

func (gui *Gui) searchBindings() []binding {
	v := SearchView
	return []binding{
		{v, gocui.KeyEnter, gui.searchEnter},
		{v, gocui.KeyEsc, gui.searchEsc},
		{v, gocui.KeyTab, gui.searchTab},
	}
}

func (gui *Gui) captureBindings() []binding {
	v := CaptureView
	return []binding{
		{v, gocui.KeyCtrlS, gui.submitCapture},
		{v, gocui.KeyEsc, gui.cancelCapture},
		{v, gocui.KeyTab, gui.captureTab},
	}
}

func (gui *Gui) pickBindings() []binding {
	v := PickView
	return []binding{
		{v, gocui.KeyEnter, gui.pickEnter},
		{v, gocui.KeyEsc, gui.pickEsc},
		{v, gocui.KeyTab, gui.pickTab},
		{v, gocui.KeyCtrlA, gui.togglePickAny},
	}
}

func (gui *Gui) searchFilterBindings() []binding {
	v := SearchFilterView
	return []binding{
		{v, gocui.MouseLeft, gui.clearSearch},
		{v, 'x', gui.clearSearch},
	}
}
