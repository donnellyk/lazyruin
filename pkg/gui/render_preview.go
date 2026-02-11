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
	default:
		gui.renderSingleNotes(v)
	}
}

func (gui *Gui) renderSingleNotes(v *gocui.View) {
	if len(gui.state.Notes.Items) == 0 {
		fmt.Fprintln(v, "No note selected.")
		return
	}

	idx := gui.state.Notes.SelectedIndex
	if idx >= len(gui.state.Notes.Items) {
		return
	}

	note := gui.state.Notes.Items[idx]

	if note.Content == "" {
		content, err := gui.loadNoteContent(note.Path)
		if err != nil {
			fmt.Fprintf(v, "Error loading note: %v", err)
			return
		}
		gui.state.Notes.Items[idx].Content = content
		note.Content = content
	}

	// Show frontmatter metadata if enabled
	if gui.state.Preview.ShowFrontmatter {
		fmt.Fprintf(v, "uuid: %s\n", note.UUID)
		fmt.Fprintf(v, "created: %s\n", note.Created.Format("2006-01-02 15:04"))
		fmt.Fprintf(v, "updated: %s\n", note.Updated.Format("2006-01-02 15:04"))
		fmt.Fprintf(v, "tags: %v\n", note.Tags)
		if len(note.InlineTags) > 0 {
			fmt.Fprintf(v, "inline-tags: %v\n", note.InlineTags)
		}
		fmt.Fprintln(v, strings.Repeat("─", 40))
		fmt.Fprintln(v, "")
	}

	content := note.Content
	if gui.state.Preview.RenderMarkdown {
		width, _ := v.InnerSize()
		if width < 10 {
			width = 40
		}
		rendered := gui.renderMarkdown(content, width-2)
		lines := strings.Split(rendered, "\n")
		start := gui.state.Preview.ScrollOffset
		if start >= len(lines) {
			start = 0
		}
		for i := start; i < len(lines); i++ {
			fmt.Fprintln(v, lines[i])
		}
	} else {
		lines := strings.Split(content, "\n")
		start := gui.state.Preview.ScrollOffset
		if start >= len(lines) {
			start = 0
		}
		for i := start; i < len(lines); i++ {
			fmt.Fprintln(v, lines[i])
		}
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
		lines = append(lines,
			fmt.Sprintf("uuid: %s", note.UUID),
			fmt.Sprintf("created: %s", note.Created.Format("2006-01-02")),
		)
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

	for i, note := range cards {
		selected := isActive && i == gui.state.Preview.SelectedCardIndex
		gui.state.Preview.CardLineRanges[i][0] = currentLine

		if selected {
			selectedStartLine = currentLine
		}

		// Upper separator with title
		title := note.Title
		if title == "" {
			title = "Untitled"
		}
		fmt.Fprintln(v, gui.buildSeparatorLine(true, " "+title+" ", "", width, selected))
		currentLine++

		// Card body content
		for _, line := range gui.buildCardContent(note, contentWidth) {
			fmt.Fprintln(v, line)
			currentLine++
		}

		// Lower separator with tags and date
		tags := note.TagsString()
		date := note.ShortDate()
		fmt.Fprintln(v, gui.buildSeparatorLine(false, "", " "+date+" · "+tags+" ", width, selected))
		currentLine++

		gui.state.Preview.CardLineRanges[i][1] = currentLine
		if selected {
			selectedEndLine = currentLine
		}

		// Blank line between cards (except last)
		if i < len(cards)-1 {
			fmt.Fprintln(v, "")
			currentLine++
		}
	}

	// Scroll to keep selected card visible
	_, viewHeight := v.InnerSize()
	originY := gui.state.Preview.ScrollOffset
	if selectedStartLine < originY {
		originY = selectedStartLine
	} else if selectedEndLine > originY+viewHeight {
		originY = selectedEndLine - viewHeight
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
		fmt.Fprintln(v, gui.buildSeparatorLine(true, " "+title+" ", "", width, selected))
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
				if j == 0 {
					fmt.Fprintf(v, "  %sL%s:%s %s\n", AnsiDim, lineNum, AnsiReset, line)
				} else {
					fmt.Fprintf(v, "%s%s\n", indent, line)
				}
				currentLine++
			}
		}

		// Lower separator
		matchCount := fmt.Sprintf(" %d matches ", len(result.Matches))
		fmt.Fprintln(v, gui.buildSeparatorLine(false, "", matchCount, width, selected))
		currentLine++

		gui.state.Preview.CardLineRanges[i][1] = currentLine
		if selected {
			selectedEndLine = currentLine
		}

		// Blank line between groups (except last)
		if i < len(results)-1 {
			fmt.Fprintln(v, "")
			currentLine++
		}
	}

	// Scroll to keep selected group visible
	_, viewHeight := v.InnerSize()
	originY := gui.state.Preview.ScrollOffset
	if selectedStartLine < originY {
		originY = selectedStartLine
	} else if selectedEndLine > originY+viewHeight {
		originY = selectedEndLine - viewHeight
	}
	gui.state.Preview.ScrollOffset = originY
	v.SetOrigin(0, originY)
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
