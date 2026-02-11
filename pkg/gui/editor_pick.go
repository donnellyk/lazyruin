package gui

import "github.com/jesseduffield/gocui"

// pickEditor is a custom gocui.Editor for the pick popup.
// When completion is active, arrow keys navigate suggestions.
// Otherwise, it delegates to SimpleEditor for standard text editing.
type pickEditor struct {
	gui *Gui
}

func (e *pickEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.PickCompletion

	if state.Active {
		switch key {
		case gocui.KeyArrowDown:
			completionDown(state)
			return true
		case gocui.KeyArrowUp:
			completionUp(state)
			return true
		}
	}

	// Delegate to SimpleEditor for all other input
	handled := gocui.SimpleEditor(v, key, ch, mod)

	// After every keystroke, update completion
	e.gui.updateCompletion(v, e.gui.pickTriggers(), state)

	return handled
}
