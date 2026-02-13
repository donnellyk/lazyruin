package gui

import "github.com/jesseduffield/gocui"

// tagInputEditor is a custom gocui.Editor for the tag input popup.
// When completion is active, arrow keys navigate suggestions.
// Otherwise it delegates to SimpleEditor.
type tagInputEditor struct {
	gui *Gui
}

func (e *tagInputEditor) triggers() []CompletionTrigger {
	if e.gui.state.TagInputConfig == nil {
		return nil
	}
	return []CompletionTrigger{
		{Prefix: "#", Candidates: e.gui.state.TagInputConfig.Candidates},
	}
}

func (e *tagInputEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.TagInputCompletion

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
	e.gui.updateCompletion(v, e.triggers(), state)

	return handled
}
