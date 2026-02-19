package gui

import (
	"fmt"
	"kvnd/lazyruin/pkg/gui/types"
	"strings"

	"github.com/jesseduffield/gocui"
)

// cursorBytePos converts a 2D cursor position (cx, cy) into a byte offset
// within the given content string (split by "\n").
func cursorBytePos(content string, cx, cy int) int {
	lines := strings.Split(content, "\n")
	pos := 0
	for i := 0; i < cy && i < len(lines); i++ {
		pos += len(lines[i]) + 1 // +1 for newline
	}
	if cy < len(lines) {
		lineLen := len(lines[cy])
		if cx > lineLen {
			cx = lineLen
		}
		pos += cx
	}
	if pos > len(content) {
		pos = len(content)
	}
	return pos
}

// viewCursorBytePos returns the cursor's byte offset in the unwrapped content.
// It handles both AutoWrap (wrapped cursor coordinates) and non-AutoWrap views.
func viewCursorBytePos(v *gocui.View) int {
	cx, cy := v.TextArea.GetCursorXY()
	unwrapped := v.TextArea.GetUnwrappedContent()

	if !v.TextArea.AutoWrap {
		return cursorBytePos(unwrapped, cx, cy)
	}

	// AutoWrap: cursor coordinates are in wrapped space.
	// Use GetContent() (includes soft newlines) to find the wrapped byte position,
	// then map back to unwrapped content.
	wrapped := v.TextArea.GetContent()
	wPos := cursorBytePos(wrapped, cx, cy)

	// Walk both strings; soft newlines only exist in wrapped content.
	wi, ui := 0, 0
	for wi < wPos && ui < len(unwrapped) {
		if wrapped[wi] == unwrapped[ui] {
			wi++
			ui++
		} else {
			wi++ // soft newline
		}
	}
	return ui
}

const maxSuggestionItems = 6

// renderSuggestionView creates or updates a suggestion dropdown view at the given position.
// It returns the view name so the caller can manage it.
func (gui *Gui) renderSuggestionView(g *gocui.Gui, viewName string, state *types.CompletionState, x0, y0, maxWidth int) error {
	if !state.Active || len(state.Items) == 0 {
		g.DeleteView(viewName)
		return nil
	}

	// Calculate dimensions
	itemCount := min(len(state.Items), maxSuggestionItems)

	// Find max label and detail widths for column alignment
	maxLabelW := 0
	maxDetailW := 0
	for _, item := range state.Items {
		if lw := len([]rune(item.Label)); lw > maxLabelW {
			maxLabelW = lw
		}
		if dw := len([]rune(item.Detail)); dw > maxDetailW {
			maxDetailW = dw
		}
	}

	// width = " " + label column + gap + detail column + " " + frame
	width := max(1+maxLabelW+2+maxDetailW+1+2, 20)
	if width > maxWidth {
		width = maxWidth
	}

	x1 := x0 + width
	y1 := y0 + itemCount + 1 // +1 for border

	v, err := g.SetView(viewName, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Frame = true
	v.FrameColor = gocui.ColorYellow
	v.TitleColor = gocui.ColorYellow
	setRoundedCorners(v)

	v.Clear()

	innerWidth := max(width-2, 10)

	// Determine visible window for scrolling
	startIdx := 0
	if state.SelectedIndex >= maxSuggestionItems {
		startIdx = state.SelectedIndex - maxSuggestionItems + 1
	}
	endIdx := startIdx + itemCount
	if endIdx > len(state.Items) {
		endIdx = len(state.Items)
		startIdx = max(endIdx-itemCount, 0)
	}

	// Detail column starts at a fixed position
	detailCol := innerWidth - maxDetailW - 1 // 1 for trailing space

	for i := startIdx; i < endIdx; i++ {
		item := state.Items[i]
		selected := i == state.SelectedIndex

		label := " " + item.Label
		detail := item.Detail

		// Pad label to reach detail column
		labelRunes := len([]rune(label))
		pad := max(detailCol-labelRunes, 1)

		line := label + strings.Repeat(" ", pad) + AnsiDim + detail + AnsiReset
		// Pad to full width for highlight (account for ANSI not taking visual space)
		visualLen := labelRunes + pad + len([]rune(detail))
		line = line + strings.Repeat(" ", max(innerWidth-visualLen, 0))

		if selected {
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line, AnsiReset)
		} else {
			fmt.Fprintln(v, line)
		}
	}

	g.SetViewOnTop(viewName)
	return nil
}
