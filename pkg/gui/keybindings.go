package gui

import (
	"fmt"
	"sort"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/types"
)

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
	// Infrastructure bindings not owned by any controller.
	navBindings := []func() []binding{
		gui.globalNavBindings,
		// notesNavBindings removed — notes bindings registered via NotesController
		// queriesNavBindings removed — queries bindings registered via QueriesController
		// tagsNavBindings removed — tags bindings registered via TagsController
		// previewNavBindings removed — preview bindings registered via PreviewController
		// searchBindings removed — bindings registered via SearchController
		// captureBindings removed — bindings registered via CaptureController
		// pickBindings removed — bindings registered via PickController
		// inputPopupBindings removed — bindings registered via InputPopupController
		// Tab/Backtab removed — bindings registered via GlobalController
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

	// Clear Search binding (SearchFilterView-specific, no controller home yet)
	if err := gui.g.SetKeybinding(SearchFilterView, 'x', gocui.ModNone, gui.clearSearch); err != nil {
		return err
	}

	if err := gui.setupDialogKeybindings(); err != nil {
		return err
	}

	// Register context/controller bindings (all migrated panels)
	if err := gui.registerContextBindings(); err != nil {
		return err
	}

	// Tab click bindings (different signature, can't be registered via controllers)
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

// registerContextBindings registers keybindings from all migrated contexts.
// This bridges the new controller/context system into gocui's keybinding API.
func (gui *Gui) registerContextBindings() error {
	opts := types.KeybindingsOpts{}

	for _, ctx := range gui.contexts.All() {
		viewNames := ctx.GetViewNames()
		kind := ctx.GetKind()

		for _, b := range ctx.GetKeybindings(opts) {
			binding := b
			// Skip palette-only bindings (Key == nil = no keybinding, just palette entry)
			if binding.Key == nil {
				continue
			}
			handler := func(g *gocui.Gui, v *gocui.View) error {
				// Suppress main/side panel bindings during popups, but allow
				// popup contexts to handle their own keybindings.
				if gui.overlayActive() && kind != types.PERSISTENT_POPUP && kind != types.TEMPORARY_POPUP {
					return nil
				}
				if binding.GetDisabledReason != nil {
					if reason := binding.GetDisabledReason(); reason != nil {
						return nil
					}
				}
				return binding.Handler()
			}

			for _, viewName := range viewNames {
				if err := gui.g.SetKeybinding(viewName, binding.Key, binding.Mod, handler); err != nil {
					return err
				}
			}
		}

		// Register mouse bindings as regular gocui keybindings (same mechanism
		// used by all other panels — gocui treats mouse events as keys).
		for _, mb := range ctx.GetMouseKeybindings(opts) {
			mouseBind := mb
			for _, viewName := range viewNames {
				handler := func(g *gocui.Gui, v *gocui.View) error {
					if gui.overlayActive() {
						return nil
					}
					return mouseBind.Handler(gocui.ViewMouseBindingOpts{})
				}
				if err := gui.g.SetKeybinding(viewName, mouseBind.Key, gocui.ModNone, handler); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (gui *Gui) globalNavBindings() []binding {
	// Tab/Backtab migrated to GlobalController.
	return []binding{
		{"", gocui.MouseWheelDown, gui.preview.previewScrollDown},
		{"", gocui.MouseWheelUp, gui.preview.previewScrollUp},
	}
}

// notesNavBindings removed — notes navigation is now handled by NotesController.
// queriesNavBindings removed — queries navigation is now handled by QueriesController.
// tagsNavBindings removed — tags navigation is now handled by TagsController.
// previewNavBindings removed — preview navigation is now handled by PreviewController.
// searchBindings removed — search bindings are now handled by SearchController.
// captureBindings removed — capture bindings are now handled by CaptureController.
// pickBindings removed — pick bindings are now handled by PickController.
// inputPopupBindings removed — input popup bindings are now handled by InputPopupController.

// DumpBindings returns a stable sorted list of all registered controller bindings
// for debugging and regression diffing. Use with --debug-bindings flag.
func (gui *Gui) DumpBindings() []string {
	opts := types.KeybindingsOpts{}
	var entries []string
	for _, ctx := range gui.contexts.All() {
		for _, b := range ctx.GetKeybindings(opts) {
			keyStr := ""
			if b.Key != nil {
				keyStr = keyDisplayString(b.Key)
			}
			for _, viewName := range ctx.GetViewNames() {
				entry := fmt.Sprintf("%-12s %-16s %-8s %s", string(ctx.GetKey()), viewName, keyStr, b.ID)
				entries = append(entries, entry)
			}
		}
	}
	sort.Strings(entries)
	return entries
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
