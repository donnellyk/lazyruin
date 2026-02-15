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
			if isParentCompletion(v, state) {
				e.gui.acceptParentCompletion(v, state)
				e.gui.renderCaptureTextArea(v)
				return true
			}
			e.gui.acceptCompletion(v, state, e.gui.captureTriggers())
			e.gui.renderCaptureTextArea(v)
			return true
		case 0: // rune input
			if (ch == '/' || ch == '#') && isWikiLinkCompletion(v, state) {
				e.gui.drillWikiLinkHeader(v, state)
				return true
			}
			if ch == '/' && isParentCompletion(v, state) {
				e.gui.drillParentChild(v, state, e.gui.captureTriggers())
				e.gui.renderCaptureTextArea(v)
				return true
			}
		}
	}

	// Intercept Enter for markdown continuation
	if key == gocui.KeyEnter && !state.Active {
		content := v.TextArea.GetUnwrappedContent()
		pos := viewCursorBytePos(v)
		line := currentLineAtPos(content, pos)

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
			e.gui.renderCaptureTextArea(v)
			e.gui.updateCompletion(v, e.gui.captureTriggers(), state)
			return true
		}
	}

	// Delegate to SimpleEditor for all other input
	handled := gocui.SimpleEditor(v, key, ch, mod)
	e.gui.renderCaptureTextArea(v)

	// After every keystroke, update completion and footer
	e.gui.updateCompletion(v, e.gui.captureTriggers(), state)
	e.gui.updateCaptureFooter()

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

	cursorPos := viewCursorBytePos(v)

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

	gui.renderCaptureTextArea(v)

	// Re-run completion — the filter now contains '#', triggering header mode
	gui.updateCompletion(v, gui.captureTriggers(), state)
}

// isParentCompletion returns true if the active completion is triggered by >.
func isParentCompletion(v *gocui.View, state *CompletionState) bool {
	if !state.Active {
		return false
	}
	content := v.TextArea.GetUnwrappedContent()
	start := state.TriggerStart
	return start < len(content) && content[start] == '>'
}

// drillParentChild commits the selected parent and transitions to showing its children.
// Pushes the selection onto the drill stack and rewrites the token as >Parent/Child/.../
// The triggers parameter determines which completion triggers are re-evaluated after drilling.
func (gui *Gui) drillParentChild(v *gocui.View, state *CompletionState, triggers []CompletionTrigger) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]
	state.ParentDrill = append(state.ParentDrill, ParentDrillEntry{
		Name: item.Label,
		UUID: item.Value,
	})

	content := v.TextArea.GetUnwrappedContent()
	cursorPos := viewCursorBytePos(v)

	// Backspace from cursor to trigger start
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	// Rebuild the path from the drill stack, preserving >> prefix
	var path strings.Builder
	prefix := ">"
	triggerEnd := state.TriggerStart + 2
	if triggerEnd <= len(content) && content[state.TriggerStart:triggerEnd] == ">>" {
		prefix = ">>"
	}
	path.WriteString(prefix)
	for _, entry := range state.ParentDrill {
		path.WriteString(entry.Name)
		path.WriteByte('/')
	}
	v.TextArea.TypeString(path.String())

	// Clear completion state
	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0

	v.RenderTextArea()

	// Re-run completion to show children
	gui.updateCompletion(v, triggers, state)
}

// acceptParentCompletion sets the selected note as the capture parent,
// removes the >... token from the content, and updates the footer.
func (gui *Gui) acceptParentCompletion(v *gocui.View, state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]

	gui.state.CaptureParent = &CaptureParentInfo{
		UUID:  item.Value,
		Title: item.Label,
	}

	cursorPos := viewCursorBytePos(v)

	// Remove the >... token from content
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	// Clear completion and drill state
	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0
	state.ParentDrill = nil

	v.RenderTextArea()
	gui.updateCaptureFooter()
}
