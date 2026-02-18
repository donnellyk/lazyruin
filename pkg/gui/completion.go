package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/types"
)

// Type aliases for backward compatibility — canonical definitions live in types/.
type CompletionItem = types.CompletionItem
type CompletionTrigger = types.CompletionTrigger
type ParentDrillEntry = types.ParentDrillEntry
type CompletionState = types.CompletionState

// NewCompletionState returns an initialized CompletionState.
var NewCompletionState = types.NewCompletionState

// extractTokenAtCursor scans backward from cursorPos to find the current token
// (delimited by whitespace or start of string). Returns the token and its start position.
func extractTokenAtCursor(content string, cursorPos int) (string, int) {
	if cursorPos > len(content) {
		cursorPos = len(content)
	}
	start := cursorPos
	for start > 0 {
		ch := content[start-1]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		start--
	}
	return content[start:cursorPos], start
}

// lineContainsAt returns true if the current line (determined by cursorPos)
// contains an '@' character. Used to scope ambient date fallback.
func lineContainsAt(content string, cursorPos int) bool {
	cp := cursorPos
	if cp > len(content) {
		cp = len(content)
	}
	// Find line start
	lineStart := strings.LastIndex(content[:cp], "\n") + 1
	// Find line end
	lineEnd := strings.Index(content[cp:], "\n")
	if lineEnd == -1 {
		lineEnd = len(content)
	} else {
		lineEnd += cp
	}
	return strings.Contains(content[lineStart:lineEnd], "@")
}

// detectTrigger checks if the token at the cursor matches any trigger prefix.
// Returns the matching trigger and the filter text (portion after the prefix), or nil.
// As a fallback, scans backward for unclosed [[ to support bracket-style triggers
// whose filter text may contain spaces.
func detectTrigger(content string, cursorPos int, triggers []CompletionTrigger) (*CompletionTrigger, string, int) {
	token, tokenStart := extractTokenAtCursor(content, cursorPos)
	for i := range triggers {
		t := &triggers[i]
		if strings.HasPrefix(token, t.Prefix) {
			filter := token[len(t.Prefix):]
			return t, filter, tokenStart
		}
	}

	// Fallback: scan for unclosed [[ before cursor (allows spaces in filter)
	cp := cursorPos
	if cp > len(content) {
		cp = len(content)
	}
	if idx := strings.LastIndex(content[:cp], "[["); idx >= 0 {
		after := content[idx+2 : cp]
		if !strings.Contains(after, "]]") {
			for i := range triggers {
				if triggers[i].Prefix == "[[" {
					return &triggers[i], after, idx
				}
			}
		}
	}

	// Fallback: scan backward for date-style prefixes whose filter may contain spaces
	// (e.g. "created:next week", "before:last monday")
	datePrefixes := []string{"created:", "updated:", "before:", "after:", "between:", "@"}
	for _, dp := range datePrefixes {
		for i := cp - 1; i >= 0; i-- {
			ch := content[i]
			if ch == '\n' {
				break
			}
			// Check if dp starts at position i
			if i+len(dp) <= cp && content[i:i+len(dp)] == dp {
				// Verify word boundary: start of string or whitespace before
				if i == 0 || content[i-1] == ' ' || content[i-1] == '\t' || content[i-1] == '\n' {
					after := content[i+len(dp) : cp]
					for j := range triggers {
						if triggers[j].Prefix == dp {
							return &triggers[j], after, i
						}
					}
				}
				break
			}
		}
	}

	// Fallback: scan backward on the current line for > at a word boundary
	// (parent completion whose filter may contain spaces after drilling)
	for i := cp - 1; i >= 0; i-- {
		ch := content[i]
		if ch == '\n' {
			break
		}
		if ch == '>' && (i == 0 || content[i-1] == ' ' || content[i-1] == '\t' || content[i-1] == '\n') {
			after := content[i+1 : cp]
			// Exclude blockquotes: "> " with a space and no / is a quote, not a parent
			if len(after) > 0 && after[0] == ' ' {
				break
			}
			for j := range triggers {
				if triggers[j].Prefix == ">" {
					return &triggers[j], after, i
				}
			}
			break
		}
	}

	return nil, "", 0
}

// updateCompletion is called after every keystroke. It checks whether a trigger
// is active and updates the CompletionState accordingly.
func (gui *Gui) updateCompletion(v *gocui.View, triggers []CompletionTrigger, state *CompletionState) {
	content := v.TextArea.GetUnwrappedContent()
	cursorPos := viewCursorBytePos(v)

	trigger, filter, tokenStart := detectTrigger(content, cursorPos, triggers)
	if trigger != nil {
		items := trigger.Candidates(filter)
		if len(items) > 0 {
			state.Active = true
			state.TriggerStart = tokenStart
			state.Items = items
			if state.SelectedIndex >= len(items) {
				state.SelectedIndex = 0
			}
			return
		}
	}

	// Fallback: ambient candidates (e.g. date parsing) when no trigger matched.
	// Only activates when @ appears on the current line, scoping suggestions
	// to contexts where the user has signaled date intent.
	if state.FallbackCandidates != nil {
		token, tStart := extractTokenAtCursor(content, cursorPos)
		if token != "" && lineContainsAt(content, cursorPos) {
			if items := state.FallbackCandidates(token); len(items) > 0 {
				state.Active = true
				state.TriggerStart = tStart
				state.Items = items
				if state.SelectedIndex >= len(items) {
					state.SelectedIndex = 0
				}
				return
			}
		}
	}

	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0
}

// acceptCompletion replaces the current trigger token with the selected item's InsertText.
// When triggers is non-nil and the item has ContinueCompleting set, completion is
// re-run immediately so the inserted prefix can chain into its own trigger.
func (gui *Gui) acceptCompletion(v *gocui.View, state *CompletionState, triggers []CompletionTrigger) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]

	if item.PrependToLine {
		gui.acceptPrependCompletion(v, state, item)
	} else {
		gui.acceptReplaceCompletion(v, state, item, triggers)
	}
}

// acceptReplaceCompletion is the standard completion: replace the trigger token with InsertText.
func (gui *Gui) acceptReplaceCompletion(v *gocui.View, state *CompletionState, item CompletionItem, triggers []CompletionTrigger) {
	cursorPos := viewCursorBytePos(v)

	// Calculate how many chars to backspace (from cursorPos back to TriggerStart)
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	if item.ContinueCompleting {
		v.TextArea.TypeString(item.InsertText)
	} else {
		v.TextArea.TypeString(item.InsertText + " ")
	}

	state.Dismiss()

	v.RenderTextArea()

	// Re-run completion so a chained trigger can activate immediately
	if item.ContinueCompleting && triggers != nil {
		gui.updateCompletion(v, triggers, state)
	}
}

// acceptPrependCompletion prepends InsertText to the current line content,
// removing the trigger token. Used for line-prefix items like headings and bullets.
func (gui *Gui) acceptPrependCompletion(v *gocui.View, state *CompletionState, item CompletionItem) {
	content := v.TextArea.GetUnwrappedContent()
	cursorPos := viewCursorBytePos(v)

	// Find line boundaries around the trigger
	lineStart := strings.LastIndex(content[:state.TriggerStart], "\n") + 1
	lineEnd := strings.Index(content[cursorPos:], "\n")
	if lineEnd == -1 {
		lineEnd = len(content)
	} else {
		lineEnd += cursorPos
	}

	// Extract line content without the trigger token
	beforeTrigger := content[lineStart:state.TriggerStart]
	afterCursor := content[cursorPos:lineEnd]
	lineContent := strings.TrimSpace(beforeTrigger + afterCursor)

	// Delete from cursor back to trigger start
	for range cursorPos - state.TriggerStart {
		v.TextArea.BackSpaceChar()
	}
	// Delete forward to end of line
	for range []rune(afterCursor) {
		v.TextArea.DeleteChar()
	}
	// Delete backward to start of line
	for range []rune(beforeTrigger) {
		v.TextArea.BackSpaceChar()
	}

	// Type the new line: prefix + existing content
	if lineContent != "" {
		v.TextArea.TypeString(item.InsertText + " " + lineContent)
	} else {
		v.TextArea.TypeString(item.InsertText + " ")
	}

	state.Dismiss()

	v.RenderTextArea()
}

// completionEsc returns an Esc handler that dismisses completion if active,
// otherwise calls onClose.
func (gui *Gui) completionEsc(state func() *CompletionState, onClose func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if s := state(); s.Active {
			s.Dismiss()
			return nil
		}
		return onClose(g, v)
	}
}

// completionTab returns a Tab handler that accepts completion if active.
func (gui *Gui) completionTab(state func() *CompletionState, triggers func() []CompletionTrigger) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if s := state(); s.Active && len(s.Items) > 0 {
			gui.acceptCompletion(v, s, triggers())
		}
		return nil
	}
}

// completionEnter returns an Enter handler that accepts completion if active,
// otherwise calls onSubmit.
func (gui *Gui) completionEnter(state func() *CompletionState, triggers func() []CompletionTrigger, onSubmit func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if s := state(); s.Active && len(s.Items) > 0 {
			gui.acceptCompletion(v, s, triggers())
			return nil
		}
		return onSubmit(g, v)
	}
}

// completionDown moves the selection down in the completion list.
func completionDown(state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}
	if state.SelectedIndex < len(state.Items)-1 {
		state.SelectedIndex++
	}
}

// completionUp moves the selection up in the completion list.
func completionUp(state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}
	if state.SelectedIndex > 0 {
		state.SelectedIndex--
	}
}
