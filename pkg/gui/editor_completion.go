package gui

import "github.com/jesseduffield/gocui"

// DrillFlags controls which drill-down behaviors a completionEditor supports.
type DrillFlags int

const (
	DrillParent   DrillFlags = 1 << iota // / drills into parent children (> trigger)
	DrillWikiLink                        // / or # drills into wiki-link headers ([[ trigger)
)

// completionEditor is a configurable gocui.Editor that handles completion
// navigation and drill-down for any popup with a CompletionState.
// It replaces the four near-identical editors (search, pick, inputPopup, snippet).
type completionEditor struct {
	gui        *Gui
	state      func() *CompletionState
	triggers   func() []CompletionTrigger
	drillFlags DrillFlags
}

func (e *completionEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	state := e.state()

	triggers := e.triggers()
	if triggers == nil {
		// No triggers configured — fall back to raw text editing
		return gocui.SimpleEditor(v, key, ch, mod)
	}

	if state.Active {
		switch key {
		case gocui.KeyArrowDown:
			completionDown(state)
			return true
		case gocui.KeyArrowUp:
			completionUp(state)
			return true
		case 0: // rune input
			if e.drillFlags&DrillWikiLink != 0 && (ch == '/' || ch == '#') && isWikiLinkCompletion(v, state) {
				e.gui.drillWikiLinkHeader(v, state)
				return true
			}
			if e.drillFlags&DrillParent != 0 && ch == '/' && isParentCompletion(v, state) {
				e.gui.drillParentChild(v, state, triggers)
				return true
			}
		}
	}

	handled := gocui.SimpleEditor(v, key, ch, mod)

	e.gui.updateCompletion(v, triggers, state)

	return handled
}
