package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

// CompletionItem represents a single suggestion in the completion dropdown.
type CompletionItem struct {
	Label              string // display text (e.g. "#project")
	InsertText         string // text to insert (e.g. "#project")
	Detail             string // right-aligned detail (e.g. "(5)")
	ContinueCompleting bool   // if true, don't add trailing space -- allows chaining into next trigger
	Value              string // opaque data (e.g. UUID) for use by accept handlers
	PrependToLine      bool   // if true, prepend InsertText to existing line content instead of replacing trigger
}

// CompletionTrigger defines a prefix that activates completion with a candidate provider.
type CompletionTrigger struct {
	Prefix     string
	Candidates func(filter string) []CompletionItem
}

// ParentDrillEntry records a parent selected during drill-down navigation.
type ParentDrillEntry struct {
	Name string
	UUID string
}

// CompletionState tracks the current state of a completion session.
type CompletionState struct {
	Active        bool
	TriggerStart  int // byte offset where the trigger token starts
	Items         []CompletionItem
	SelectedIndex int
	ParentDrill   []ParentDrillEntry // stack of drilled-into parents for > completion
}

// NewCompletionState returns an initialized CompletionState.
func NewCompletionState() *CompletionState {
	return &CompletionState{}
}

// Dismiss fully resets the completion state, hiding the suggestion dropdown.
func (s *CompletionState) Dismiss() {
	s.Active = false
	s.Items = nil
	s.SelectedIndex = 0
	s.TriggerStart = 0
	s.ParentDrill = nil
}

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
