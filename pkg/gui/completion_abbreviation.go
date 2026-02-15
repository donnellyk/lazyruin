package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/commands"
)

// isAbbreviationCompletion returns true if the active completion is triggered by !.
func isAbbreviationCompletion(v *gocui.View, state *CompletionState) bool {
	if !state.Active {
		return false
	}
	content := v.TextArea.GetUnwrappedContent()
	start := state.TriggerStart
	return start < len(content) && content[start] == '!'
}

// extractParentPath finds a >path token in expansion text. The parent path
// starts at a whitespace-delimited ">" and extends to end-of-string or until
// the next recognized token prefix (#, [[). This allows parent names with spaces
// (e.g. ">My Project/Daily Log").
// Returns the remaining text (without the >path token) and the path (without > prefix).
func extractParentPath(text string) (remaining, parentPath string) {
	// Find > at a word boundary (start of string or preceded by whitespace)
	startIdx := -1
	for i := 0; i < len(text); i++ {
		if text[i] == '>' {
			if i == 0 || text[i-1] == ' ' {
				// Check it's not a bare > at end
				if i+1 < len(text) && text[i+1] != ' ' {
					startIdx = i
					break
				}
			}
		}
	}
	if startIdx < 0 {
		return text, ""
	}

	// Path extends from after > until we hit a recognized token prefix or end
	pathStart := startIdx + 1
	// Skip additional > for >> mode
	if pathStart < len(text) && text[pathStart] == '>' {
		pathStart++
	}
	endIdx := len(text)
	// Scan for next token prefix: # or [[ (preceded by space)
	for i := pathStart; i < len(text); i++ {
		if text[i] == ' ' && i+1 < len(text) {
			next := text[i+1]
			if next == '#' || (next == '[' && i+2 < len(text) && text[i+2] == '[') {
				endIdx = i
				break
			}
		}
	}

	parentPath = strings.TrimSpace(text[pathStart:endIdx])
	before := strings.TrimSpace(text[:startIdx])
	after := strings.TrimSpace(text[endIdx:])
	parts := []string{}
	if before != "" {
		parts = append(parts, before)
	}
	if after != "" {
		parts = append(parts, after)
	}
	remaining = strings.Join(parts, " ")
	return remaining, parentPath
}

// resolveParentPath resolves a /-delimited path to a CaptureParentInfo.
// The first segment matches a parent bookmark name; if no bookmark matches,
// it falls back to searching all notes by title (for >> mode).
// Subsequent segments drill into children by matching titles.
func (gui *Gui) resolveParentPath(path string) *CaptureParentInfo {
	segments := strings.Split(path, "/")
	if len(segments) == 0 || segments[0] == "" {
		return nil
	}

	// Find the bookmark matching the first segment
	var currentUUID string
	var titleParts []string
	segLower := strings.ToLower(segments[0])
	for _, p := range gui.state.Parents.Items {
		if strings.ToLower(p.Name) == segLower {
			currentUUID = p.UUID
			titleParts = append(titleParts, p.Title)
			break
		}
	}

	// Fallback: match by note title (for >> all-notes mode)
	if currentUUID == "" {
		for _, note := range gui.state.Notes.Items {
			if strings.ToLower(note.Title) == segLower {
				currentUUID = note.UUID
				titleParts = append(titleParts, note.Title)
				break
			}
		}
	}

	if currentUUID == "" {
		return nil
	}

	// Drill into children for subsequent segments
	for _, seg := range segments[1:] {
		if seg == "" {
			continue
		}
		children, err := gui.ruinCmd.Search.Search("parent:"+currentUUID, commands.SearchOptions{
			Limit: 50,
		})
		if err != nil {
			return nil
		}
		sl := strings.ToLower(seg)
		found := false
		for _, child := range children {
			if strings.ToLower(child.Title) == sl {
				currentUUID = child.UUID
				titleParts = append(titleParts, child.Title)
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return &CaptureParentInfo{
		UUID:  currentUUID,
		Title: titleParts[len(titleParts)-1],
	}
}

// acceptAbbreviationInCapture handles accepting an abbreviation completion in Capture mode.
// It expands the abbreviation, extracts any >path token to set the capture parent,
// and inserts the remaining text.
func (gui *Gui) acceptAbbreviationInCapture(v *gocui.View, state *CompletionState) {
	if !state.Active || len(state.Items) == 0 {
		return
	}

	item := state.Items[state.SelectedIndex]
	remaining, parentPath := extractParentPath(item.InsertText)

	// Backspace from cursor to trigger start
	cursorPos := viewCursorBytePos(v)
	charsToDelete := cursorPos - state.TriggerStart
	for range charsToDelete {
		v.TextArea.BackSpaceChar()
	}

	// Type the remaining text (without >path token)
	v.TextArea.TypeString(remaining)

	// Resolve parent path if present
	if parentPath != "" {
		if info := gui.resolveParentPath(parentPath); info != nil {
			gui.state.CaptureParent = info
		}
	}

	// Clear completion state
	state.Active = false
	state.Items = nil
	state.SelectedIndex = 0

	gui.updateCaptureFooter()
}
