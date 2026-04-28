package gui

import (
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// isTriggerBoundary reports whether position i in content is a valid
// start-of-trigger boundary. Trigger characters (e.g. `#`, `@`) only fire
// completion when preceded by a boundary so we don't match inside words.
//
// Boundaries are: start-of-string, whitespace, `!` (for negation like
// `!#tag`), and option separators `=`, `,`, `|` (so completion fires
// inside embed option values like `![[pick: ... | filter=#tag`).
func isTriggerBoundary(content string, i int) bool {
	if i == 0 {
		return true
	}
	switch content[i-1] {
	case ' ', '\t', '\n', '!', '=', ',', '|':
		return true
	}
	return false
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

// lineContainsAt returns true if the current line (determined by cursorPos)
// contains an '@' character. Used to scope ambient date fallback.
func lineContainsAt(content string, cursorPos int) bool {
	cp := min(cursorPos, len(content))
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
func detectTrigger(content string, cursorPos int, triggers []types.CompletionTrigger) (*types.CompletionTrigger, string, int) {
	token, tokenStart := extractTokenAtCursor(content, cursorPos)
	for i := range triggers {
		t := &triggers[i]
		if strings.HasPrefix(token, t.Prefix) {
			filter := token[len(t.Prefix):]
			return t, filter, tokenStart
		}
	}

	cp := min(cursorPos, len(content))

	// Fallback: scan backward on the current line for `#` at a trigger
	// boundary (start-of-line, whitespace, `!` for negation, or `=`, `,`,
	// `|` for use inside embed option values like `![[pick:... | filter=#tag`).
	// Must run before the `![[` fallback so it can fire inside embed queries.
	for i := cp - 1; i >= 0; i-- {
		ch := content[i]
		if ch == '\n' {
			break
		}
		if ch == '#' {
			if isTriggerBoundary(content, i) {
				for j := range triggers {
					if triggers[j].Prefix == "#" {
						after := content[i+1 : cp]
						// Require the after-text to be a plausible tag segment
						// (no spaces) to avoid matching headings or mid-word `#`.
						if !strings.ContainsAny(after, " \t") {
							return &triggers[j], after, i
						}
					}
				}
			}
			break
		}
	}

	// Fallback: scan backward for date-style prefixes whose filter may contain
	// spaces (e.g. "created:next week", "before:last monday"). Runs before
	// the `![[` scan so date/`@` completion can fire inside embed queries
	// like `![[pick: ... | filter=@today` without being intercepted by the
	// embed fallback.
	datePrefixes := []string{"created:", "updated:", "before:", "after:", "between:", "@"}
	for _, dp := range datePrefixes {
		for i := cp - 1; i >= 0; i-- {
			ch := content[i]
			if ch == '\n' {
				break
			}
			// Check if dp starts at position i
			if i+len(dp) <= cp && content[i:i+len(dp)] == dp {
				if isTriggerBoundary(content, i) {
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

	// Fallback: scan for unclosed ![[ on the current line (embed trigger,
	// longer prefix — must win over the plain `[[` scan below). Limited to
	// the current line because the reference spec requires dynamic embeds to
	// be standalone lines; matching across newlines would confuse completion
	// dispatch with insideDynamicEmbed which is also line-scoped.
	lineStart := strings.LastIndexByte(content[:cp], '\n') + 1
	if idx := strings.LastIndex(content[lineStart:cp], "![["); idx >= 0 {
		absIdx := lineStart + idx
		after := content[absIdx+3 : cp]
		if !strings.Contains(after, "]]") {
			for i := range triggers {
				if triggers[i].Prefix == "![[" {
					return &triggers[i], after, absIdx
				}
			}
		}
	}

	// Fallback: scan for unclosed [[ before cursor (allows spaces in filter)
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
// is active and updates the types.CompletionState accordingly.
func (gui *Gui) updateCompletion(v *gocui.View, triggers []types.CompletionTrigger, state *types.CompletionState) {
	content := v.TextArea.GetUnwrappedContent()
	cursorPos := viewCursorBytePos(v)

	// Dynamic-embed aware dispatch. Inside an unclosed `![[...`:
	//   - options phase (`| key=value`): offer option key/value candidates
	//   - `query:` / `compose:` types: offer saved-query / bookmark candidates
	//   - `search:` / `pick:` or no type yet: fall through to normal triggers.
	es := insideDynamicEmbed(content, cursorPos)
	if es.inEmbed {
		if items, start, ok := gui.dynamicEmbedCandidates(content, cursorPos, es); ok {
			if len(items) > 0 {
				state.Active = true
				state.TriggerStart = start
				state.Items = items
				if state.SelectedIndex >= len(items) {
					state.SelectedIndex = 0
				}
				return
			}
			// Embed dispatch fired but returned no items. Still dismiss so
			// we don't leave a stale dropdown open.
			state.Active = false
			state.Items = nil
			state.SelectedIndex = 0
			return
		}
	}

	trigger, filter, tokenStart := detectTrigger(content, cursorPos, triggers)
	if trigger != nil {
		items := trigger.Candidates(filter)
		if len(items) > 0 {
			state.Active = true
			state.TriggerStart = tokenStart
			state.Items = items
			if state.SelectedIndex >= len(items) || state.Items[state.SelectedIndex].IsHeader {
				state.SelectedIndex = types.FirstSelectable(items)
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
func (gui *Gui) acceptCompletion(v *gocui.View, state *types.CompletionState, triggers []types.CompletionTrigger) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]
	if item.IsHeader {
		return
	}

	if item.PrependToLine {
		gui.acceptPrependCompletion(v, state, item)
	} else {
		gui.acceptReplaceCompletion(v, state, item, triggers)
	}
}

// acceptReplaceCompletion is the standard completion: replace the trigger token with InsertText.
func (gui *Gui) acceptReplaceCompletion(v *gocui.View, state *types.CompletionState, item types.CompletionItem, triggers []types.CompletionTrigger) {
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
func (gui *Gui) acceptPrependCompletion(v *gocui.View, state *types.CompletionState, item types.CompletionItem) {
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
func (gui *Gui) completionEsc(state func() *types.CompletionState, onClose func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if s := state(); s.Active {
			s.Dismiss()
			return nil
		}
		return onClose(g, v)
	}
}

// completionTab returns a Tab handler that accepts completion if active.
func (gui *Gui) completionTab(state func() *types.CompletionState, triggers func() []types.CompletionTrigger) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if s := state(); s.Active && len(s.Items) > 0 {
			gui.acceptCompletion(v, s, triggers())
		}
		return nil
	}
}

// completionEnter returns an Enter handler that accepts completion if active,
// otherwise calls onSubmit.
func (gui *Gui) completionEnter(state func() *types.CompletionState, triggers func() []types.CompletionTrigger, onSubmit func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if s := state(); s.Active && len(s.Items) > 0 {
			gui.acceptCompletion(v, s, triggers())
			return nil
		}
		return onSubmit(g, v)
	}
}

// completionDown moves the selection down in the completion list, skipping headers.
func completionDown(state *types.CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}
	for i := state.SelectedIndex + 1; i < len(state.Items); i++ {
		if !state.Items[i].IsHeader {
			state.SelectedIndex = i
			return
		}
	}
}

// completionUp moves the selection up in the completion list, skipping headers.
func completionUp(state *types.CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}
	for i := state.SelectedIndex - 1; i >= 0; i-- {
		if !state.Items[i].IsHeader {
			state.SelectedIndex = i
			return
		}
	}
}

// captureTab handles Tab in the capture popup, accepting the active completion.
func (gui *Gui) captureTab(g *gocui.Gui, v *gocui.View) error {
	state := gui.contexts.Capture.Completion
	if state.Active {
		// Check for action items (e.g. /scratchpad)
		if len(state.Items) > 0 {
			selected := state.Items[state.SelectedIndex]
			if selected.Value == "action:scratchpad" {
				state.Dismiss()
				gui.clearTriggerToken(v, state)
				gui.renderCaptureTextArea(v)
				return gui.helpers.Scratchpad().OpenBrowserForInsert(func(text string) {
					if gui.views.Capture != nil {
						gui.views.Capture.TextArea.TypeString(text)
					}
				})
			}
		}
		if isParentCompletion(v, state) {
			gui.acceptParentCompletion(v, state)
		} else {
			gui.acceptCompletion(v, state, gui.captureTriggers())
		}
		gui.renderCaptureTextArea(v)
	}
	return nil
}

// clearTriggerToken removes the trigger text (e.g. "/scratchpad") from the editor
// without inserting replacement text.
func (gui *Gui) clearTriggerToken(v *gocui.View, state *types.CompletionState) {
	cursorPos := viewCursorBytePos(v)
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}
}
