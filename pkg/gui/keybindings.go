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
	// Register command-table bindings (actions driven by commands())
	for _, cmd := range gui.commands() {
		if cmd.Handler == nil || len(cmd.Keys) == 0 {
			continue
		}
		handler := cmd.Handler
		if cmd.View == "" {
			handler = gui.suppressDuringDialog(handler)
		}
		for _, key := range cmd.Keys {
			if err := gui.g.SetKeybinding(cmd.View, key, gocui.ModNone, handler); err != nil {
				return err
			}
		}
	}

	// Navigation and infrastructure bindings (not user-facing commands)
	navBindings := []func() []binding{
		gui.globalNavBindings,
		gui.notesNavBindings,
		gui.queriesNavBindings,
		gui.tagsNavBindings,
		gui.previewNavBindings,
		gui.searchBindings,
		gui.captureBindings,
		gui.pickBindings,
		gui.paletteBindings,
	}
	for _, fn := range navBindings {
		bindings := fn()
		for i, b := range bindings {
			if b.view == "" {
				bindings[i].handler = gui.suppressDuringDialog(b.handler)
			}
		}
		if err := gui.registerBindings(bindings); err != nil {
			return err
		}
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

func (gui *Gui) globalNavBindings() []binding {
	return []binding{
		{"", gocui.KeyTab, gui.nextPanel},
		{"", gocui.KeyBacktab, gui.prevPanel},
		{"", gocui.MouseWheelDown, gui.previewScrollDown},
		{"", gocui.MouseWheelUp, gui.previewScrollUp},
	}
}

func (gui *Gui) notesNavBindings() []binding {
	v := NotesView
	return []binding{
		{v, gocui.MouseLeft, gui.notesClick},
		{v, 'j', gui.notesDown},
		{v, 'k', gui.notesUp},
		{v, gocui.KeyArrowDown, gui.notesDown},
		{v, gocui.KeyArrowUp, gui.notesUp},
		{v, 'g', gui.notesTop},
		{v, 'G', gui.notesBottom},
		{v, gocui.MouseWheelDown, gui.notesWheelDown},
		{v, gocui.MouseWheelUp, gui.notesWheelUp},
	}
}

func (gui *Gui) queriesNavBindings() []binding {
	v := QueriesView
	return []binding{
		{v, gocui.MouseLeft, gui.queriesClick},
		{v, 'j', gui.queriesDown},
		{v, 'k', gui.queriesUp},
		{v, gocui.KeyArrowDown, gui.queriesDown},
		{v, gocui.KeyArrowUp, gui.queriesUp},
		{v, gocui.MouseWheelDown, gui.queriesWheelDown},
		{v, gocui.MouseWheelUp, gui.queriesWheelUp},
	}
}

func (gui *Gui) tagsNavBindings() []binding {
	v := TagsView
	return []binding{
		{v, gocui.MouseLeft, gui.tagsClick},
		{v, 'j', gui.tagsDown},
		{v, 'k', gui.tagsUp},
		{v, gocui.KeyArrowDown, gui.tagsDown},
		{v, gocui.KeyArrowUp, gui.tagsUp},
		{v, gocui.MouseWheelDown, gui.tagsWheelDown},
		{v, gocui.MouseWheelUp, gui.tagsWheelUp},
	}
}

func (gui *Gui) previewNavBindings() []binding {
	v := PreviewView
	return []binding{
		{v, gocui.MouseLeft, gui.previewClick},
		{v, 'j', gui.previewDown},
		{v, 'k', gui.previewUp},
		{v, 'J', gui.previewCardDown},
		{v, 'K', gui.previewCardUp},
		{v, ']', gui.previewNextHeader},
		{v, '[', gui.previewPrevHeader},
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

func (gui *Gui) paletteBindings() []binding {
	v := PaletteView
	lv := PaletteListView
	return []binding{
		{v, gocui.KeyEnter, gui.paletteEnter},
		{v, gocui.KeyEsc, gui.paletteEsc},
		{lv, gocui.MouseLeft, gui.paletteListClick},
	}
}
