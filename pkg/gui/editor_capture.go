package gui

import "github.com/jesseduffield/gocui"

// captureEditor is a custom gocui.Editor for the capture popup.
// When completion is active, arrow keys navigate suggestions and Enter/Tab accepts.
// Otherwise, Enter inserts a newline (multi-line editing) and other keys are
// handled by SimpleEditor.
type captureEditor struct {
	gui *Gui
}

func (e *captureEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.CaptureCompletion

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
	e.gui.updateCompletion(v, e.gui.captureTriggers(), state)

	return handled
}
