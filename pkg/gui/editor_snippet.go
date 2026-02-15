package gui

import "github.com/jesseduffield/gocui"

// snippetExpansionEditor is a custom gocui.Editor for the snippet expansion field.
// When completion is active, arrow keys navigate suggestions and / drills
// into parent children. Otherwise it delegates to SimpleEditor.
type snippetExpansionEditor struct {
	gui *Gui
}

func (e *snippetExpansionEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.gui.state.SnippetEditorCompletion
	triggers := e.gui.snippetExpansionTriggers()

	if state.Active {
		switch key {
		case gocui.KeyArrowDown:
			completionDown(state)
			return true
		case gocui.KeyArrowUp:
			completionUp(state)
			return true
		case 0: // rune input
			if (ch == '/' || ch == '#') && isWikiLinkCompletion(v, state) {
				e.gui.drillWikiLinkHeader(v, state)
				return true
			}
			if ch == '/' && isParentCompletion(v, state) {
				e.gui.drillParentChild(v, state, triggers)
				return true
			}
		}
	}

	handled := gocui.SimpleEditor(v, key, ch, mod)

	e.gui.updateCompletion(v, triggers, state)

	return handled
}
