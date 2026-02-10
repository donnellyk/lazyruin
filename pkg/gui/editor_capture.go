package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

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
		case 0: // rune input
			if (ch == '/' || ch == '#') && isWikiLinkCompletion(v, state) {
				e.gui.drillWikiLinkHeader(v, state)
				return true
			}
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

// isWikiLinkCompletion returns true if the active completion is triggered by [[.
func isWikiLinkCompletion(v *gocui.View, state *CompletionState) bool {
	if !state.Active {
		return false
	}
	content := v.TextArea.GetUnwrappedContent()
	start := state.TriggerStart
	return start+2 <= len(content) && content[start:start+2] == "[["
}

// drillWikiLinkHeader commits the selected note title and transitions to header mode.
// It replaces the current trigger token with [[Title# so the completion system
// re-detects the filter as containing '#' and switches to header candidates.
func (gui *Gui) drillWikiLinkHeader(v *gocui.View, state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]
	// The item.Label is the note title (without [[ or ]])
	noteTitle := item.Label

	// If the label contains '#', we're already in header mode — ignore the drill.
	if strings.Contains(noteTitle, "#") {
		return
	}

	content := v.TextArea.GetUnwrappedContent()
	cx, cy := v.TextArea.GetCursorXY()
	cursorPos := cursorBytePos(content, cx, cy)

	// Backspace from cursor to trigger start
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	// Insert [[Title#
	v.TextArea.TypeString("[[" + noteTitle + "#")

	// Clear completion state
	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0

	v.RenderTextArea()

	// Re-run completion — the filter now contains '#', triggering header mode
	gui.updateCompletion(v, gui.captureTriggers(), state)
}
