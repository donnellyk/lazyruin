package gui

import "github.com/jesseduffield/gocui"

// inputPopupEditor is a custom gocui.Editor for the generic input popup.
// When completion is active, arrow keys navigate suggestions and / drills
// into parent children. Otherwise it delegates to SimpleEditor.
type inputPopupEditor struct {
	gui *Gui
}

func (e *inputPopupEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.InputPopupCompletion
	config := e.gui.state.InputPopupConfig
	if config == nil {
		return false
	}

	triggers := config.Triggers()

	if state.Active {
		switch key {
		case gocui.KeyArrowDown:
			completionDown(state)
			return true
		case gocui.KeyArrowUp:
			completionUp(state)
			return true
		}

		// Drill into children with / (only fires when trigger is >)
		if key == 0 && ch == '/' && isParentCompletion(v, state) {
			e.gui.drillParentChild(v, state, triggers)
			return true
		}
	}

	// Delegate to SimpleEditor for all other input
	handled := gocui.SimpleEditor(v, key, ch, mod)

	// After every keystroke, update completion
	e.gui.updateCompletion(v, triggers, state)

	return handled
}
