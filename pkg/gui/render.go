package gui

import (
	"fmt"
	"os"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) renderNotes() {
	v := gui.views.Notes
	if v == nil {
		return
	}

	v.Clear()

	if len(gui.state.Notes.Items) == 0 {
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, " No notes found.")
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, " Press 'n' to create a new note")
		fmt.Fprintln(v, " or '/' to search")
		return
	}

	// Get view width for full-line highlighting
	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	// Check if this panel is focused for highlighting
	isActive := gui.state.CurrentContext == NotesContext

	for i, note := range gui.state.Notes.Items {
		selected := isActive && i == gui.state.Notes.SelectedIndex

		// Line 1 - Title
		title := note.Title
		if title == "" {
			title = note.Path
		}
		if len(title) > width-2 {
			title = title[:width-5] + "..."
		}
		line1 := " " + title

		// Line 2 - Date and tags
		date := note.ShortDate()
		tags := note.TagsString()
		maxTagLen := width - len(date) - 7 // "  " + " · " + some buffer
		if maxTagLen > 0 && len(tags) > maxTagLen {
			tags = tags[:maxTagLen-3] + "..."
		}
		line2 := fmt.Sprintf("   %s · %s", date, tags)

		if selected {
			// Pad lines to full width for complete highlight
			line1 = line1 + strings.Repeat(" ", width-len(line1))
			line2 = line2 + strings.Repeat(" ", width-len(line2))
			// Blue background, white text
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line1, AnsiReset)
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line2, AnsiReset)
		} else {
			fmt.Fprintln(v, line1)
			fmt.Fprintln(v, line2)
		}

		// fmt.Fprintln(v, "")
	}

	// Scroll to keep selection visible (2 lines per note)
	_, viewHeight := v.Size()
	selLine := gui.state.Notes.SelectedIndex * 2
	scrollListView(v, selLine, 2, viewHeight)
}

func (gui *Gui) renderQueries() {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		gui.renderParents()
		return
	}
	gui.renderQueriesList()
}

func (gui *Gui) renderQueriesList() {
	v := gui.views.Queries
	if v == nil {
		return
	}

	v.Clear()

	if len(gui.state.Queries.Items) == 0 {
		fmt.Fprintln(v, " No saved queries.")
		return
	}

	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	isActive := gui.state.CurrentContext == QueriesContext

	for i, query := range gui.state.Queries.Items {
		selected := isActive && i == gui.state.Queries.SelectedIndex

		line1 := "  " + query.Name
		queryStr := query.Query
		if len(queryStr) > 25 {
			queryStr = queryStr[:22] + "..."
		}
		line2 := "    " + queryStr

		if selected {
			line1 = line1 + strings.Repeat(" ", max(0, width-len(line1)))
			line2 = line2 + strings.Repeat(" ", max(0, width-len(line2)))
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line1, AnsiReset)
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line2, AnsiReset)
		} else {
			fmt.Fprintln(v, line1)
			fmt.Fprintln(v, line2)
		}
	}

	_, viewHeight := v.Size()
	selLine := gui.state.Queries.SelectedIndex * 2
	scrollListView(v, selLine, 2, viewHeight)
}

func (gui *Gui) renderParents() {
	v := gui.views.Queries
	if v == nil {
		return
	}

	v.Clear()

	if len(gui.state.Parents.Items) == 0 {
		fmt.Fprintln(v, " No parent bookmarks.")
		return
	}

	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	isActive := gui.state.CurrentContext == QueriesContext

	for i, parent := range gui.state.Parents.Items {
		selected := isActive && i == gui.state.Parents.SelectedIndex

		line1 := "  " + parent.Name
		title := parent.Title
		if len(title) > width-6 {
			title = title[:width-9] + "..."
		}
		line2 := "    " + title

		if selected {
			line1 = line1 + strings.Repeat(" ", max(0, width-len(line1)))
			line2 = line2 + strings.Repeat(" ", max(0, width-len(line2)))
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line1, AnsiReset)
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line2, AnsiReset)
		} else {
			fmt.Fprintln(v, line1)
			fmt.Fprintln(v, line2)
		}
	}

	_, viewHeight := v.Size()
	selLine := gui.state.Parents.SelectedIndex * 2
	scrollListView(v, selLine, 2, viewHeight)
}

func (gui *Gui) renderTags() {
	v := gui.views.Tags
	if v == nil {
		return
	}

	v.Clear()

	if len(gui.state.Tags.Items) == 0 {
		fmt.Fprintln(v, " No tags found.")
		return
	}

	for _, tag := range gui.state.Tags.Items {
		prefix := " "

		name := tag.Name
		if len(name) > 0 && name[0] != '#' {
			name = "#" + name
		}
		count := fmt.Sprintf("(%d)", tag.Count)

		fmt.Fprintf(v, "%s%s %s\n", prefix, name, count)
	}

	// Scroll to keep selection visible (1 line per tag)
	_, viewHeight := v.Size()
	selLine := gui.state.Tags.SelectedIndex
	scrollListView(v, selLine, 1, viewHeight)
	_, oy := v.Origin()
	v.SetCursor(0, selLine-oy)
}

func (gui *Gui) renderPreview() {
	v := gui.views.Preview
	if v == nil {
		return
	}

	v.Clear()

	if gui.state.Preview.Mode == PreviewModeCardList {
		gui.renderSeparatorCards(v)
	} else {
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
		fmt.Fprintln(v, strings.Repeat("─", 40))
		fmt.Fprintln(v, "")
	}

	lines := strings.Split(note.Content, "\n")
	start := gui.state.Preview.ScrollOffset
	if start >= len(lines) {
		start = 0
	}

	for i := start; i < len(lines); i++ {
		fmt.Fprintln(v, lines[i])
	}
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

	// Content width
	contentWidth := width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Track line positions for scrolling and click mapping
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
		upperSep := gui.buildSeparatorLine(true, " "+title+" ", "", width, selected)
		fmt.Fprintln(v, upperSep)
		currentLine++

		// Frontmatter if enabled
		if gui.state.Preview.ShowFrontmatter {
			fmt.Fprintf(v, "uuid: %s\n", note.UUID)
			fmt.Fprintf(v, "created: %s\n", note.Created.Format("2006-01-02"))
			currentLine += 2
		}

		// Full content with wrapping
		content := note.Content
		if content == "" {
			content, _ = gui.loadNoteContent(note.Path)
		}
		for _, l := range strings.Split(content, "\n") {
			wrapped := wrapLine(l, contentWidth)
			for _, wl := range wrapped {
				fmt.Fprintln(v, " "+wl)
				currentLine++
			}
		}

		// Lower separator with tags and date
		tags := note.TagsString()
		date := note.ShortDate()
		lowerSep := gui.buildSeparatorLine(false, "", " "+date+" · "+tags+" ", width, selected)
		fmt.Fprintln(v, lowerSep)
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

	// Only scroll when the selected card isn't fully visible
	_, viewHeight := v.InnerSize()
	originY := gui.state.Preview.ScrollOffset
	if selectedStartLine < originY {
		// Selected card starts above the viewport — scroll up
		originY = selectedStartLine
	} else if selectedEndLine > originY+viewHeight {
		// Selected card ends below the viewport — scroll down just enough
		originY = selectedEndLine - viewHeight
	}
	gui.state.Preview.ScrollOffset = originY
	v.SetOrigin(0, originY)
}

// wrapLine breaks a line into chunks that fit within the given width,
// wrapping at word boundaries when possible.
func wrapLine(s string, width int) []string {
	s = strings.ReplaceAll(s, "\t", "    ")
	runes := []rune(s)
	if len(runes) <= width {
		return []string{s}
	}
	var lines []string
	for len(runes) > width {
		// Find the last space at or before the width limit
		breakAt := -1
		for i := width; i >= 0; i-- {
			if runes[i] == ' ' {
				breakAt = i
				break
			}
		}
		if breakAt <= 0 {
			// No space found; hard break at width
			lines = append(lines, string(runes[:width]))
			runes = runes[width:]
		} else {
			lines = append(lines, string(runes[:breakAt]))
			runes = runes[breakAt+1:] // skip the space
		}
	}
	if len(runes) > 0 {
		lines = append(lines, string(runes))
	}
	return lines
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

// scrollListView sets the origin of a list view to keep selLine visible.
func scrollListView(v *gocui.View, selLine, itemHeight, viewHeight int) {
	_, currentOrigin := v.Origin()
	origin := currentOrigin

	// If selected item ends beyond viewport, scroll down
	if selLine+itemHeight > origin+viewHeight {
		origin = selLine + itemHeight - viewHeight
	}
	// If selected item starts above viewport, scroll up
	if selLine < origin {
		origin = selLine
	}

	v.SetOrigin(0, origin)
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
