package gui

import (
	"fmt"
	"os"
	"strings"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
	"github.com/muesli/reflow/wordwrap"
)

func (gui *Gui) renderPreview() {
	v := gui.views.Preview
	if v == nil {
		return
	}

	v.Clear()

	switch gui.state.Preview.Mode {
	case PreviewModeCardList:
		gui.renderSeparatorCards(v)
	case PreviewModePickResults:
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
func (gui *Gui) fprintPreviewLine(v *gocui.View, line string, lineNum int, highlight bool) {
	if highlight && lineNum == gui.state.Preview.CursorLine {
		width, _ := v.InnerSize()
		pad := width - visibleWidth(line)
		if pad < 0 {
			pad = 0
		}
		// Re-apply background after every ANSI reset so chroma formatting
		// doesn't clear our highlight mid-line.
		patched := strings.ReplaceAll(line, AnsiReset, AnsiReset+AnsiDimBg)
		fmt.Fprintf(v, "%s%s%s%s\n", AnsiDimBg, patched, strings.Repeat(" ", pad), AnsiReset)
	} else {
		fmt.Fprintln(v, line)
	}
}

// buildCardContent returns the rendered lines for a single card's body content.
func (gui *Gui) buildCardContent(note models.Note, contentWidth int) []string {
	content := note.Content
	if content == "" {
		content, _ = gui.loadNoteContent(note.Path)
	}

	var lines []string

	if gui.state.Preview.ShowFrontmatter {
		if fm, err := gui.loadNoteFrontmatter(note.Path); err == nil && fm != "" {
			lines = append(lines, " "+AnsiDim+"---"+AnsiReset)
			for _, fl := range strings.Split(fm, "\n") {
				lines = append(lines, " "+AnsiDim+fl+AnsiReset)
			}
			lines = append(lines, " "+AnsiDim+"---"+AnsiReset)
		}
	}

	if gui.state.Preview.RenderMarkdown {
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
	cards := gui.state.Preview.Cards
	if len(cards) == 0 {
		fmt.Fprintln(v, "No matching notes.")
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)

	isActive := gui.state.CurrentContext == PreviewContext
	selectedStartLine := 0
	selectedEndLine := 0
	currentLine := 0
	gui.state.Preview.CardLineRanges = make([][2]int, len(cards))
	gui.state.Preview.HeaderLines = gui.state.Preview.HeaderLines[:0]

	for i, note := range cards {
		selected := isActive && i == gui.state.Preview.SelectedCardIndex
		gui.state.Preview.CardLineRanges[i][0] = currentLine

		if selected {
			selectedStartLine = currentLine
		}

		// Upper separator with title (and "Temporarily Moved" badge for multi-card)
		title := note.Title
		if title == "" {
			title = "Untitled"
		}
		upperRight := ""
		if gui.state.Preview.TemporarilyMoved[i] && len(cards) > 1 {
			upperRight = " Temporarily Moved "
		}
		gui.fprintPreviewLine(v, gui.buildSeparatorLine(true, " "+title+" ", upperRight, width, selected), currentLine, isActive)
		currentLine++

		// Card body content
		for _, line := range gui.buildCardContent(note, contentWidth) {
			if isHeaderLine(line) {
				gui.state.Preview.HeaderLines = append(gui.state.Preview.HeaderLines, currentLine)
			}
			gui.fprintPreviewLine(v, line, currentLine, isActive)
			currentLine++
		}

		// Lower separator with tags and date
		tags := note.TagsString()
		date := note.ShortDate()
		gui.fprintPreviewLine(v, gui.buildSeparatorLine(false, "", " "+date+" · "+tags+" ", width, selected), currentLine, isActive)
		currentLine++

		gui.state.Preview.CardLineRanges[i][1] = currentLine
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
	originY := gui.state.Preview.ScrollOffset
	if isActive {
		cl := gui.state.Preview.CursorLine
		idx := gui.state.Preview.SelectedCardIndex
		// If cursor is on the first content line of a card, show the upper separator too
		showFrom := cl
		showTo := cl
		if idx < len(gui.state.Preview.CardLineRanges) {
			r := gui.state.Preview.CardLineRanges[idx]
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
	gui.state.Preview.ScrollOffset = originY
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

// renderPickResults renders line-level pick results grouped by note title
func (gui *Gui) renderPickResults(v *gocui.View) {
	results := gui.state.Preview.PickResults
	if len(results) == 0 {
		fmt.Fprintln(v, "No matching lines.")
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	contentWidth := max(width-2, 10)

	isActive := gui.state.CurrentContext == PreviewContext
	selectedStartLine := 0
	selectedEndLine := 0
	currentLine := 0
	gui.state.Preview.CardLineRanges = make([][2]int, len(results))
	gui.state.Preview.HeaderLines = gui.state.Preview.HeaderLines[:0]

	for i, result := range results {
		selected := isActive && i == gui.state.Preview.SelectedCardIndex
		gui.state.Preview.CardLineRanges[i][0] = currentLine

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

		gui.state.Preview.CardLineRanges[i][1] = currentLine
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
	originY := gui.state.Preview.ScrollOffset
	if isActive {
		cl := gui.state.Preview.CursorLine
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
	gui.state.Preview.ScrollOffset = originY
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
