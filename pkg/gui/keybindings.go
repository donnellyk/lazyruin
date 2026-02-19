package gui

import (
	"fmt"
	"sort"

	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// setupKeybindings configures all keyboard shortcuts.
func (gui *Gui) setupKeybindings() error {
	// Clear Search binding (SearchFilterView-specific)
	if err := gui.g.SetKeybinding(SearchFilterView, 'x', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		gui.helpers.Search().ClearSearch()
		return nil
	}); err != nil {
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
	if err := gui.g.SetTabClickBinding(NotesView, gui.suppressTabClickDuringDialog(
		func(idx int) error { return gui.helpers.Notes().SwitchNotesTabByIndex(idx) },
	)); err != nil {
		return err
	}
	if err := gui.g.SetTabClickBinding(QueriesView, gui.suppressTabClickDuringDialog(
		func(idx int) error { return gui.helpers.Queries().SwitchQueriesTabByIndex(idx) },
	)); err != nil {
		return err
	}
	if err := gui.g.SetTabClickBinding(TagsView, gui.suppressTabClickDuringDialog(
		func(idx int) error { return gui.helpers.Tags().SwitchTagsTabByIndex(idx) },
	)); err != nil {
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

			// If ViewName is set on the binding, register only for that view.
			// Otherwise register for all context views.
			views := viewNames
			if binding.ViewName != "" {
				views = []string{binding.ViewName}
			}
			for _, viewName := range views {
				if err := gui.g.SetKeybinding(viewName, binding.Key, binding.Mod, handler); err != nil {
					return err
				}
			}
		}

		// Register mouse bindings. If ViewMouseBinding.ViewName is set, register
		// only for that view; otherwise register for all context views.
		// Popup context mouse bindings are not suppressed (same rule as keyboard).
		for _, mb := range ctx.GetMouseKeybindings(opts) {
			mouseBind := mb
			mouseViews := viewNames
			if mouseBind.ViewName != "" {
				mouseViews = []string{mouseBind.ViewName}
			}
			for _, viewName := range mouseViews {
				handler := func(g *gocui.Gui, v *gocui.View) error {
					if gui.overlayActive() && kind != types.PERSISTENT_POPUP && kind != types.TEMPORARY_POPUP {
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
