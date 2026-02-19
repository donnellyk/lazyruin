package gui

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
	"github.com/muesli/reflow/wordwrap"
)

// inlineTagRe matches #tag patterns (hashtag followed by word characters).
var inlineTagRe = regexp.MustCompile(`#[\w-]+`)

func (gui *Gui) RenderPreview() {
	v := gui.views.Preview
	if v == nil {
		return
	}

	v.Clear()

	// Snapshot and clear link highlight — it only survives a single render
	// cycle. highlightNextLink/highlightPrevLink set it right before calling
	// renderPreview, so it's visible for this render but auto-clears for any
	// subsequent render triggered by other navigation.
	gui.contexts.Preview.RenderedLink = gui.contexts.Preview.HighlightedLink
	gui.contexts.Preview.HighlightedLink = -1

	switch gui.contexts.Preview.Mode {
	case context.PreviewModeCardList:
		gui.renderSeparatorCards(v)
	case context.PreviewModePickResults:
		gui.renderPickResults(v)
	}
}

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var sb strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			inEsc = true
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// visibleWidth returns the number of visible runes in a string, ignoring ANSI escape sequences.
func visibleWidth(s string) int {
	return len([]rune(stripAnsi(s)))
}

// isHeaderLine checks whether a rendered line is a markdown header.
func isHeaderLine(line string) bool {
	trimmed := strings.TrimLeft(stripAnsi(line), " ")
	return strings.HasPrefix(trimmed, "#")
}

// fprintPreviewLine writes a line to the preview view, applying a dim background
// highlight across the full view width when lineNum matches the current CursorLine.
// When a link is highlighted (HighlightedLink >= 0), only the link span is highlighted
// instead of the full line.
func (gui *Gui) fprintPreviewLine(v *gocui.View, line string, lineNum int, highlight bool) {
	if !highlight || lineNum != gui.contexts.Preview.CursorLine {
		fmt.Fprintln(v, line)
		return
	}

	// Check for link-only highlight (set by renderPreview snapshot)
	hl := gui.contexts.Preview.RenderedLink
	if hl >= 0 && hl < len(gui.contexts.Preview.Links) {
		link := gui.contexts.Preview.Links[hl]
		if link.Line == lineNum {
			fmt.Fprintln(v, highlightSpan(line, link.Col, link.Len))
			return
		}
	}

	// Full-line highlight
	width, _ := v.InnerSize()
	pad := width - visibleWidth(line)
	if pad < 0 {
		pad = 0
	}
	// Re-apply background after every ANSI reset so chroma formatting
	// doesn't clear our highlight mid-line.
	patched := strings.ReplaceAll(line, AnsiReset, AnsiReset+AnsiDimBg)
	fmt.Fprintf(v, "%s%s%s%s\n", AnsiDimBg, patched, strings.Repeat(" ", pad), AnsiReset)
}

// highlightSpan applies AnsiDimBg to a span of visible characters in an ANSI-decorated
// string. col and length are in visible-character units (ignoring ANSI escapes).
func highlightSpan(line string, col, length int) string {
	var sb strings.Builder
	runes := []rune(line)
	visPos := 0
	inEsc := false
	spanStart := col
	spanEnd := col + length

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inEsc {
			sb.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			sb.WriteRune(r)
			inEsc = true
			continue
		}

		// Visible character
		if visPos == spanStart {
			sb.WriteString(AnsiDimBg)
		}
		sb.WriteRune(r)
		visPos++
		if visPos == spanEnd {
			sb.WriteString(AnsiReset)
		}
	}
	// Safety: close highlight if line ended before spanEnd
	if visPos < spanEnd && visPos >= spanStart {
		sb.WriteString(AnsiReset)
	}
	return sb.String()
}

// buildCardContent returns the rendered lines for a single card's body content.
func (gui *Gui) BuildCardContent(note models.Note, contentWidth int) []string {
	content := note.Content
	if content == "" {
		content, _ = gui.loadNoteContent(note.Path)
	}

	var lines []string

	if gui.contexts.Preview.ShowFrontmatter {
		if fm, err := gui.loadNoteFrontmatter(note.Path); err == nil && fm != "" {
			lines = append(lines, " "+AnsiDim+"---"+AnsiReset)
			for _, fl := range strings.Split(fm, "\n") {
				lines = append(lines, " "+AnsiDim+fl+AnsiReset)
			}
			lines = append(lines, " "+AnsiDim+"---"+AnsiReset)
		}
	}

	if gui.contexts.Preview.RenderMarkdown {
		rendered := gui.renderMarkdown(content, contentWidth)
		for _, rl := range strings.Split(rendered, "\n") {
			lines = append(lines, " "+rl)
		}
	} else {
		for _, l := range strings.Split(content, "\n") {
			for _, wl := range wrapLine(l, contentWidth) {
				lines = append(lines, " "+wl)
			}
		}
	}

	// Trim visually empty lines from start and end (strip ANSI before checking,
	// since rendered markdown lines contain escape codes even when visually blank)
	for len(lines) > 0 && strings.TrimSpace(stripAnsi(lines[0])) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(stripAnsi(lines[len(lines)-1])) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// renderSeparatorCards renders cards using separator lines instead of frames
func (gui *Gui) renderSeparatorCards(v *gocui.View) {
	cards := gui.contexts.Preview.Cards
	if len(cards) == 0 {
		fmt.Fprintln(v, "No matching notes.")
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)

	isActive := gui.contextMgr.Current() == "preview"
	selectedStartLine := 0
	selectedEndLine := 0
	currentLine := 0
	gui.contexts.Preview.CardLineRanges = make([][2]int, len(cards))
	gui.contexts.Preview.HeaderLines = gui.contexts.Preview.HeaderLines[:0]

	for i, note := range cards {
		selected := isActive && i == gui.contexts.Preview.SelectedCardIndex
		gui.contexts.Preview.CardLineRanges[i][0] = currentLine

		if selected {
			selectedStartLine = currentLine
		}

		// Upper separator with title (and "Temporarily Moved" badge for multi-card)
		title := note.Title
		if title == "" {
			title = "Untitled"
		}
		upperRight := ""
		if gui.contexts.Preview.TemporarilyMoved[i] && len(cards) > 1 {
			upperRight = " Temporarily Moved "
		}
		gui.fprintPreviewLine(v, gui.buildSeparatorLine(true, " "+title+" ", upperRight, width, selected), currentLine, isActive)
		currentLine++

		// Card body content
		for _, line := range gui.BuildCardContent(note, contentWidth) {
			if isHeaderLine(line) {
				gui.contexts.Preview.HeaderLines = append(gui.contexts.Preview.HeaderLines, currentLine)
			}
			gui.fprintPreviewLine(v, line, currentLine, isActive)
			currentLine++
		}

		// Lower separator with date, global tags, and parent (if set)
		var parentLabel string
		if note.Parent != "" {
			parentLabel = gui.resolveParentLabel(note.Parent)
		}
		rightText := ""
		if meta := models.JoinDot(note.ShortDate(), note.GlobalTagsString(), parentLabel); meta != "" {
			rightText = " " + meta + " "
		}
		gui.fprintPreviewLine(v, gui.buildSeparatorLine(false, "", rightText, width, selected), currentLine, isActive)
		currentLine++

		gui.contexts.Preview.CardLineRanges[i][1] = currentLine
		if selected {
			selectedEndLine = currentLine
		}

		// Blank line between cards (except last)
		if i < len(cards)-1 {
			gui.fprintPreviewLine(v, "", currentLine, isActive)
			currentLine++
		}
	}

	// Scroll to keep cursor/card visible, including borders when at card edges
	_, viewHeight := v.InnerSize()
	originY := gui.contexts.Preview.ScrollOffset
	if isActive {
		cl := gui.contexts.Preview.CursorLine
		idx := gui.contexts.Preview.SelectedCardIndex
		// If cursor is on the first content line of a card, show the upper separator too
		showFrom := cl
		showTo := cl
		if idx < len(gui.contexts.Preview.CardLineRanges) {
			r := gui.contexts.Preview.CardLineRanges[idx]
			if cl == r[0]+1 {
				showFrom = r[0] // include upper separator
			}
			if cl == r[1]-2 {
				showTo = r[1] - 1 // include lower separator
			}
		}
		if showFrom < originY {
			originY = showFrom
		} else if showTo >= originY+viewHeight {
			originY = showTo - viewHeight + 1
		}
	} else {
		if selectedStartLine < originY {
			originY = selectedStartLine
		} else if selectedEndLine > originY+viewHeight {
			originY = selectedEndLine - viewHeight
		}
	}
	gui.contexts.Preview.ScrollOffset = originY
	v.SetOrigin(0, originY)
}

// buildSeparatorLine creates a separator line with optional left and right text
func (gui *Gui) buildSeparatorLine(upper bool, leftText, rightText string, width int, highlight bool) string {
	dim := AnsiDim
	green := AnsiGreen
	reset := AnsiReset

	sep := "─"
	leftLen := len([]rune(leftText))
	rightLen := len([]rune(rightText))

	// Calculate fill length
	fillLen := width - leftLen - rightLen - 4 // 4 for leading/trailing separator chars
	if fillLen < 0 {
		fillLen = 0
	}

	var sb strings.Builder
	if highlight {
		sb.WriteString(green)
	}
	sb.WriteString(dim)
	if upper {
		sb.WriteString("╭")
	} else {
		sb.WriteString("╰")
	}
	sb.WriteString(sep)
	sb.WriteString(leftText)
	for i := 0; i < fillLen; i++ {
		sb.WriteString(sep)
	}
	sb.WriteString(rightText)
	sb.WriteString(sep)
	if upper {
		sb.WriteString("╮")
	} else {
		sb.WriteString("╯")
	}
	sb.WriteString(reset)

	return sb.String()
}

// resolveParentLabel returns a display name for a parent UUID by checking
// loaded parent bookmarks, then loaded notes, then falling back to a truncated UUID.
func (gui *Gui) resolveParentLabel(uuid string) string {
	for _, bm := range gui.contexts.Queries.Parents {
		if bm.UUID == uuid {
			return bm.Name
		}
	}
	for _, note := range gui.contexts.Notes.Items {
		if note.UUID == uuid {
			return note.Title
		}
	}
	// Fallback: show truncated UUID
	if len(uuid) > 8 {
		return uuid[:8] + "..."
	}
	return uuid
}

// renderPickResults renders line-level pick results grouped by note title
func (gui *Gui) renderPickResults(v *gocui.View) {
	results := gui.contexts.Preview.PickResults
	if len(results) == 0 {
		fmt.Fprintln(v, "No matching lines.")
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)

	isActive := gui.contextMgr.Current() == "preview"
	selectedStartLine := 0
	selectedEndLine := 0
	currentLine := 0
	gui.contexts.Preview.CardLineRanges = make([][2]int, len(results))
	gui.contexts.Preview.HeaderLines = gui.contexts.Preview.HeaderLines[:0]

	for i, result := range results {
		selected := isActive && i == gui.contexts.Preview.SelectedCardIndex
		gui.contexts.Preview.CardLineRanges[i][0] = currentLine

		if selected {
			selectedStartLine = currentLine
		}

		// Separator header with note title
		title := result.Title
		if title == "" {
			title = "Untitled"
		}
		gui.fprintPreviewLine(v, gui.buildSeparatorLine(true, " "+title+" ", "", width, selected), currentLine, isActive)
		currentLine++

		// Render each match line
		for _, match := range result.Matches {
			lineNum := fmt.Sprintf("%02d", match.Line)
			prefix := fmt.Sprintf("  L%s: ", lineNum)
			prefixLen := len(prefix)
			highlighted := gui.highlightMarkdown(match.Content)
			wrapped := wordwrap.String(highlighted, contentWidth-prefixLen)
			indent := strings.Repeat(" ", prefixLen)
			for j, line := range strings.Split(strings.TrimRight(wrapped, "\n"), "\n") {
				var formatted string
				if j == 0 {
					formatted = fmt.Sprintf("  %sL%s:%s %s", AnsiDim, lineNum, AnsiReset, line)
				} else {
					formatted = indent + line
				}
				gui.fprintPreviewLine(v, formatted, currentLine, isActive)
				currentLine++
			}
		}

		// Lower separator
		matchCount := fmt.Sprintf(" %d matches ", len(result.Matches))
		gui.fprintPreviewLine(v, gui.buildSeparatorLine(false, "", matchCount, width, selected), currentLine, isActive)
		currentLine++

		gui.contexts.Preview.CardLineRanges[i][1] = currentLine
		if selected {
			selectedEndLine = currentLine
		}

		// Blank line between groups (except last)
		if i < len(results)-1 {
			gui.fprintPreviewLine(v, "", currentLine, isActive)
			currentLine++
		}
	}

	// Scroll to keep cursor/group visible
	_, viewHeight := v.InnerSize()
	originY := gui.contexts.Preview.ScrollOffset
	if isActive {
		cl := gui.contexts.Preview.CursorLine
		if cl < originY {
			originY = cl
		} else if cl >= originY+viewHeight {
			originY = cl - viewHeight + 1
		}
	} else {
		if selectedStartLine < originY {
			originY = selectedStartLine
		} else if selectedEndLine > originY+viewHeight {
			originY = selectedEndLine - viewHeight
		}
	}
	gui.contexts.Preview.ScrollOffset = originY
	v.SetOrigin(0, originY)
}

// loadNoteFrontmatter returns the raw YAML frontmatter block (without the --- delimiters).
func (gui *Gui) loadNoteFrontmatter(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	raw := string(data)
	if !strings.HasPrefix(raw, "---") {
		return "", nil
	}
	rest := raw[3:]
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return "", nil
	}
	fm := strings.TrimPrefix(rest[:idx], "\n")
	return fm, nil
}

func (gui *Gui) loadNoteContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)

	// Strip YAML frontmatter if present
	if strings.HasPrefix(content, "---") {
		// Find the closing ---
		rest := content[3:]
		if idx := strings.Index(rest, "\n---"); idx != -1 {
			content = strings.TrimLeft(rest[idx+4:], "\n")
		}
	}

	return content, nil
}
