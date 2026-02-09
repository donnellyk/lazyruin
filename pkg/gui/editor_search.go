package gui

import "github.com/jesseduffield/gocui"

// searchEditor is a custom gocui.Editor for the search popup.
// When completion is active, arrow keys navigate suggestions and Tab accepts.
// Otherwise, it delegates to SimpleEditor for standard text editing.
type searchEditor struct {
	gui *Gui
}

func (e *searchEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.SearchCompletion

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
	e.gui.updateCompletion(v, e.gui.searchTriggers(), state)

	return handled
}
