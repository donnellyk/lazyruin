package gui

import "github.com/jesseduffield/gocui"

// captureEditor is a custom gocui.Editor for the capture popup.
// When completion is active, arrow keys navigate suggestions and Enter/Tab accepts.
// Otherwise, Enter continues markdown list/quote prefixes, and other keys are
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
		case gocui.KeyEnter:
			e.gui.acceptCompletion(v, state, e.gui.captureTriggers())
			return true
		}
	}

	// Intercept Enter for markdown continuation
	if key == gocui.KeyEnter && !state.Active {
		content := v.TextArea.GetUnwrappedContent()
		_, cy := v.TextArea.GetCursorXY()
		line := currentLine(content, cy)

		if cont := markdownContinuation(line); cont != nil {
			if cont.Empty {
				// Empty list item -- clear the prefix instead of continuing
				for range len(cont.Prefix) {
					v.TextArea.BackSpaceChar()
				}
			} else {
				// Insert newline + continuation prefix
				v.TextArea.TypeString("\n" + cont.Prefix)
			}
			v.RenderTextArea()
			e.gui.updateCompletion(v, e.gui.captureTriggers(), state)
			return true
		}
	}

	// Delegate to SimpleEditor for all other input
	handled := gocui.SimpleEditor(v, key, ch, mod)

	// After every keystroke, update completion
	e.gui.updateCompletion(v, e.gui.captureTriggers(), state)

	return handled
}
