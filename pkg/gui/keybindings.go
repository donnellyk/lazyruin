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
		views := cmd.Views
		if views == nil {
			views = []string{""}
			handler = gui.suppressDuringDialog(handler)
		}
		for _, view := range views {
			h := handler
			if isMainPanelView(view) {
				h = gui.suppressDuringDialog(handler)
			}
			for _, key := range cmd.Keys {
				if err := gui.g.SetKeybinding(view, key, gocui.ModNone, h); err != nil {
					return err
				}
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
		gui.inputPopupBindings,
		gui.snippetEditorBindings,
		gui.paletteBindings,
		gui.calendarBindings,
		gui.contribBindings,
	}
	for _, fn := range navBindings {
		bindings := fn()
		for i, b := range bindings {
			if b.view == "" || isMainPanelView(b.view) {
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
	if err := gui.g.SetTabClickBinding(NotesView, gui.suppressTabClickDuringDialog(gui.switchNotesTabByIndex)); err != nil {
		return err
	}
	if err := gui.g.SetTabClickBinding(QueriesView, gui.suppressTabClickDuringDialog(gui.switchQueriesTabByIndex)); err != nil {
		return err
	}
	if err := gui.g.SetTabClickBinding(TagsView, gui.suppressTabClickDuringDialog(gui.switchTagsTabByIndex)); err != nil {
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
		{v, gocui.KeyArrowDown, gui.previewDown},
		{v, gocui.KeyArrowUp, gui.previewUp},
		{v, 'J', gui.previewCardDown},
		{v, 'K', gui.previewCardUp},
		{v, '}', gui.previewNextHeader},
		{v, '{', gui.previewPrevHeader},
		{v, 'l', gui.highlightNextLink},
		{v, 'L', gui.highlightPrevLink},
	}
}

func (gui *Gui) searchBindings() []binding {
	v := SearchView
	searchState := func() *CompletionState { return gui.state.SearchCompletion }
	return []binding{
		{v, gocui.KeyEnter, gui.completionEnter(searchState, gui.searchTriggers, gui.executeSearch)},
		{v, gocui.KeyEsc, gui.completionEsc(searchState, gui.cancelSearch)},
		{v, gocui.KeyTab, gui.completionTab(searchState, gui.searchTriggers)},
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
	pickState := func() *CompletionState { return gui.state.PickCompletion }
	return []binding{
		{v, gocui.KeyEnter, gui.completionEnter(pickState, gui.pickTriggers, gui.executePick)},
		{v, gocui.KeyEsc, gui.completionEsc(pickState, gui.cancelPick)},
		{v, gocui.KeyTab, gui.completionTab(pickState, gui.pickTriggers)},
		{v, gocui.KeyCtrlA, gui.togglePickAny},
	}
}

func (gui *Gui) inputPopupBindings() []binding {
	v := InputPopupView
	return []binding{
		{v, gocui.KeyEnter, gui.inputPopupEnter},
		{v, gocui.KeyEsc, gui.inputPopupEsc},
		{v, gocui.KeyTab, gui.inputPopupTab},
	}
}

func (gui *Gui) snippetEditorBindings() []binding {
	nv := SnippetNameView
	ev := SnippetExpansionView
	return []binding{
		{nv, gocui.KeyEsc, gui.snippetEditorEsc},
		{nv, gocui.KeyTab, gui.snippetEditorTab},
		{nv, gocui.KeyEnter, gui.snippetEditorTab},
		{nv, gocui.MouseLeft, gui.snippetEditorClickName},
		{ev, gocui.KeyEsc, gui.snippetEditorEsc},
		{ev, gocui.KeyTab, gui.snippetEditorTab},
		{ev, gocui.KeyEnter, gui.snippetEditorEnter},
		{ev, gocui.MouseLeft, gui.snippetEditorClickExpansion},
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

func (gui *Gui) calendarBindings() []binding {
	gv := CalendarGridView
	iv := CalendarInputView
	nv := CalendarNotesView

	return []binding{
		// Grid navigation
		{gv, 'h', gui.calendarGridLeft},
		{gv, 'l', gui.calendarGridRight},
		{gv, 'k', gui.calendarGridUp},
		{gv, 'j', gui.calendarGridDown},
		{gv, gocui.KeyArrowLeft, gui.calendarGridLeft},
		{gv, gocui.KeyArrowRight, gui.calendarGridRight},
		{gv, gocui.KeyArrowUp, gui.calendarGridUp},
		{gv, gocui.KeyArrowDown, gui.calendarGridDown},
		{gv, gocui.KeyEnter, gui.calendarGridEnter},
		{gv, gocui.KeyEsc, gui.calendarEsc},
		{gv, gocui.KeyTab, gui.calendarTab},
		{gv, gocui.KeyBacktab, gui.calendarBacktab},
		{gv, '/', gui.calendarFocusInput},
		{gv, gocui.MouseLeft, gui.calendarGridClick},
		// Input view
		{iv, gocui.KeyEnter, gui.calendarInputEnter},
		{iv, gocui.KeyEsc, gui.calendarInputEsc},
		{iv, gocui.KeyTab, gui.calendarTab},
		{iv, gocui.KeyBacktab, gui.calendarBacktab},
		{iv, gocui.MouseLeft, gui.calendarInputClick},
		// Note list navigation
		{nv, 'j', gui.calendarNoteDown},
		{nv, 'k', gui.calendarNoteUp},
		{nv, gocui.KeyArrowDown, gui.calendarNoteDown},
		{nv, gocui.KeyArrowUp, gui.calendarNoteUp},
		{nv, gocui.KeyEnter, gui.calendarNoteEnter},
		{nv, gocui.KeyEsc, gui.calendarEsc},
		{nv, gocui.KeyTab, gui.calendarTab},
		{nv, gocui.KeyBacktab, gui.calendarBacktab},
		{nv, '/', gui.calendarFocusInput},
	}
}

func (gui *Gui) contribBindings() []binding {
	gv := ContribGridView
	nv := ContribNotesView
	return []binding{
		// Grid navigation (h/l = weeks/columns, j/k = days/rows)
		{gv, 'h', gui.contribGridLeft},
		{gv, 'l', gui.contribGridRight},
		{gv, 'k', gui.contribGridUp},
		{gv, 'j', gui.contribGridDown},
		{gv, gocui.KeyArrowLeft, gui.contribGridLeft},
		{gv, gocui.KeyArrowRight, gui.contribGridRight},
		{gv, gocui.KeyArrowUp, gui.contribGridUp},
		{gv, gocui.KeyArrowDown, gui.contribGridDown},
		{gv, gocui.KeyEnter, gui.contribGridEnter},
		{gv, gocui.KeyEsc, gui.contribEsc},
		{gv, gocui.KeyTab, gui.contribTab},
		// Note list navigation
		{nv, 'j', gui.contribNoteDown},
		{nv, 'k', gui.contribNoteUp},
		{nv, gocui.KeyArrowDown, gui.contribNoteDown},
		{nv, gocui.KeyArrowUp, gui.contribNoteUp},
		{nv, gocui.KeyEnter, gui.contribNoteEnter},
		{nv, gocui.KeyEsc, gui.contribEsc},
		{nv, gocui.KeyTab, gui.contribTab},
	}
}
