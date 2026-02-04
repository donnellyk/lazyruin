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
			fmt.Fprintf(v, "\x1b[44;37m%s\x1b[0m\n", line1)
			fmt.Fprintf(v, "\x1b[44;37m%s\x1b[0m\n", line2)
		} else {
			fmt.Fprintln(v, line1)
			fmt.Fprintln(v, line2)
		}

		// fmt.Fprintln(v, "")
	}

	v.SetOrigin(0, 0)
}

func (gui *Gui) renderQueries() {
	v := gui.views.Queries
	if v == nil {
		return
	}

	v.Clear()

	if len(gui.state.Queries.Items) == 0 {
		fmt.Fprintln(v, " No saved queries.")
		return
	}

	for _, query := range gui.state.Queries.Items {
		prefix := "  "

		// Line 1
		fmt.Fprintf(v, "%s%s\n", prefix, query.Name)

		// Line 2
		queryStr := query.Query
		if len(queryStr) > 25 {
			queryStr = queryStr[:22] + "..."
		}
		fmt.Fprintf(v, "    %s\n", queryStr)
	}

	// Sync view cursor with selection for highlight (2 lines per query)
	v.SetCursor(0, gui.state.Queries.SelectedIndex*2)
	v.SetOrigin(0, 0)
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

	// Sync view cursor with selection for highlight
	v.SetCursor(0, gui.state.Tags.SelectedIndex)
	v.SetOrigin(0, 0)
}

func (gui *Gui) renderPreview() {
	v := gui.views.Preview
	if v == nil {
		return
	}

	v.Clear()

	// Clean up old card views first
	gui.cleanupCardViews()

	if gui.state.Preview.Mode == PreviewModeCardList {
		gui.renderCardViews()
	} else {
		gui.renderSingleNotes(v)
	}
}

// cleanupCardViews removes any existing card views
func (gui *Gui) cleanupCardViews() {
	for _, name := range gui.state.Preview.CardViewNames {
		gui.g.DeleteView(name)
	}
	gui.state.Preview.CardViewNames = nil
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

	// Apply content filtering based on toggle state
	content := gui.filterContent(note.Content)

	lines := strings.Split(content, "\n")
	start := gui.state.Preview.ScrollOffset
	if start >= len(lines) {
		start = 0
	}

	for i := start; i < len(lines); i++ {
		fmt.Fprintln(v, lines[i])
	}
}

// padOrTruncate ensures a string is exactly 'length' display characters
func padOrTruncate(s string, length int) string {
	// Replace tabs with spaces for consistent width
	s = strings.ReplaceAll(s, "\t", "    ")

	runes := []rune(s)
	if len(runes) > length {
		if length > 3 {
			return string(runes[:length-3]) + "..."
		}
		return string(runes[:length])
	}
	return s + strings.Repeat(" ", length-len(runes))
}

// renderCardViews creates actual gocui views for each card
func (gui *Gui) renderCardViews() {
	cards := gui.state.Preview.Cards
	if len(cards) == 0 {
		fmt.Fprintln(gui.views.Preview, "No matching notes.")
		return
	}

	// Get the preview panel's position
	x0, y0, x1, y1 := gui.views.Preview.Dimensions()

	// Card dimensions
	cardHeight := 8 // Height of each card view
	if gui.state.Preview.ShowFrontmatter {
		cardHeight += 2
	}
	cardWidth := x1 - x0 - 2 // Leave margin inside preview

	// Starting position for first card (inside preview panel)
	startX := x0 + 1
	startY := y0 + 1
	currentY := startY

	for i, note := range cards {
		// Stop if we run out of vertical space
		if currentY+cardHeight > y1-1 {
			break
		}

		selected := i == gui.state.Preview.SelectedCardIndex
		viewName := fmt.Sprintf("card-%d", i)

		// Create the card view
		cardView, err := gui.g.SetView(viewName, startX, currentY, startX+cardWidth, currentY+cardHeight, 0)
		if err != nil && err.Error() != "unknown view" {
			continue
		}

		// Track for cleanup
		gui.state.Preview.CardViewNames = append(gui.state.Preview.CardViewNames, viewName)

		// Configure the card view
		title := note.Title
		if title == "" {
			title = "Untitled"
		}
		cardView.Title = " " + title + " "
		cardView.Footer = fmt.Sprintf("%s  %s", note.TagsString(), note.ShortDate())
		cardView.Wrap = true
		setRoundedCorners(cardView)

		// Set frame color based on selection
		if selected {
			cardView.FrameColor = gocui.ColorGreen
			cardView.TitleColor = gocui.ColorGreen
		} else {
			cardView.FrameColor = gocui.ColorDefault
			cardView.TitleColor = gocui.ColorDefault
		}

		// Render card content
		cardView.Clear()

		// Frontmatter if enabled
		if gui.state.Preview.ShowFrontmatter {
			fmt.Fprintf(cardView, "uuid: %s\n", note.UUID)
			fmt.Fprintf(cardView, "created: %s\n", note.Created.Format("2006-01-02"))
		}

		// Content preview
		content := note.Content
		if content == "" {
			content, _ = gui.loadNoteContent(note.Path)
		}
		content = gui.filterContent(content)
		contentLines := strings.Split(content, "\n")
		maxLines := 4
		for j, l := range contentLines {
			if j >= maxLines {
				fmt.Fprintln(cardView, "...")
				break
			}
			fmt.Fprintln(cardView, l)
		}

		// Move to next card position
		currentY += cardHeight + 1
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
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

// filterContent applies display filters based on preview toggle state
func (gui *Gui) filterContent(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skipNextEmpty := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle global tags (lines that are only hashtags at the start)
		if !gui.state.Preview.ShowGlobalTags && i < 3 {
			if isGlobalTagLine(trimmed) {
				skipNextEmpty = true
				continue
			}
		}

		// Handle title (first H1 heading)
		if !gui.state.Preview.ShowTitle {
			if strings.HasPrefix(trimmed, "# ") && len(result) == 0 {
				skipNextEmpty = true
				continue
			}
		}

		// Skip empty lines after stripped content
		if skipNextEmpty && trimmed == "" {
			skipNextEmpty = false
			continue
		}
		skipNextEmpty = false

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// isGlobalTagLine checks if a line contains only hashtags (global tags)
func isGlobalTagLine(line string) bool {
	if line == "" {
		return false
	}
	// Line should be space-separated hashtags like "#tag1 #tag2"
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !strings.HasPrefix(part, "#") {
			return false
		}
	}
	return true
}
