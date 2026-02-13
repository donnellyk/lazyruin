package gui

import "github.com/jesseduffield/gocui"

// parentInputEditor is a custom gocui.Editor for the parent input popup.
// When completion is active, arrow keys navigate suggestions, Enter accepts
// the parent, and / drills into children. Otherwise it delegates to SimpleEditor.
type parentInputEditor struct {
	gui *Gui
}

func (e *parentInputEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.ParentInputCompletion

	if state.Active {
		switch key {
		case gocui.KeyArrowDown:
			completionDown(state)
			return true
		case gocui.KeyArrowUp:
			completionUp(state)
			return true
		}

		// Drill into children with /
		if key == 0 && ch == '/' && isParentCompletion(v, state) {
			e.gui.drillParentChild(v, state, e.gui.parentInputTriggers())
			return true
		}
	}

	// Delegate to SimpleEditor for all other input
	handled := gocui.SimpleEditor(v, key, ch, mod)

	// After every keystroke, update completion
	e.gui.updateCompletion(v, e.gui.parentInputTriggers(), state)

	return handled
}
