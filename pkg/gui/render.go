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

	for i, note := range gui.state.Notes.Items {
		prefix := "  "
		if i == gui.state.Notes.SelectedIndex {
			prefix = "> "
		}

		// Line 1
		title := note.Title
		if title == "" {
			title = note.Path // TODO: Change this. Most paths will be gibberish timestamps
		}

		title = title[:30] // TODO: unsafe truncation
		fmt.Fprintf(v, "%s%s\n", prefix, title)

		// Line 2
		date := note.ShortDate()
		tags := note.TagsString()
		if len(tags) > 20 {
			tags = tags[:20] + "…"
		}
		fmt.Fprintf(v, "  %s · %s\n", date, tags)

		fmt.Fprintln(v, "")
	}
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

	for i, query := range gui.state.Queries.Items {
		prefix := "  "
		if i == gui.state.Queries.SelectedIndex {
			prefix = "> "
		}

		// Line 1
		fmt.Fprintf(v, "%s%s\n", prefix, query.Name)

		// Line 2
		queryStr := query.Query
		if len(queryStr) > 25 {
			queryStr = queryStr[:25]
		}
		fmt.Fprintf(v, "    %s\n", queryStr)
	}
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

	for i, tag := range gui.state.Tags.Items {
		prefix := "  "
		if i == gui.state.Tags.SelectedIndex {
			prefix = "> "
		}

		name := "#" + tag.Name
		count := fmt.Sprintf("(%d)", tag.Count)

		fmt.Fprintf(v, "%s%-20s %s\n", prefix, name, count)
	}
}

func (gui *Gui) renderPreview() {
	v := gui.views.Preview
	if v == nil {
		return
	}

	v.Clear()

	if gui.state.Preview.Mode == PreviewModeCardList {
		gui.renderCardList(v)
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

		// TODO: This looks weird, the double setting of state. Look into it.
		gui.state.Notes.Items[idx].Content = content
		note.Content = content
	}

	// TODO: Implement proper frontmatter display (ie. show file as-is or strip it out)
	if gui.state.Preview.ShowFrontmatter {
		fmt.Fprintf(v, "uuid: %s\n", note.UUID)
		fmt.Fprintf(v, "created: %s\n", note.Created.Format("2006-01-02 15:04"))
		fmt.Fprintf(v, "updated: %s\n", note.Updated.Format("2006-01-02 15:04"))
		fmt.Fprintf(v, "tags: %v\n", note.Tags)
		fmt.Fprintln(v, strings.Repeat("─", 40))
		fmt.Fprintln(v, "")
	}

	// TODO: Seems like a wild way to render lines of text...
	lines := strings.Split(note.Content, "\n")
	start := gui.state.Preview.ScrollOffset
	if start >= len(lines) {
		start = 0
	}

	for i := start; i < len(lines); i++ {
		fmt.Fprintln(v, lines[i])
	}
}

func (gui *Gui) renderCardList(v *gocui.View) {
	cards := gui.state.Preview.Cards
	if len(cards) == 0 {
		fmt.Fprintln(v, "No matching notes.")
		return
	}

	width, _ := v.Size()
	if width < 10 {
		width = 40
	}

	for i, note := range cards {
		selected := i == gui.state.Preview.SelectedCardIndex

		// Card border style
		borderChar := "─"
		cornerTL := "┌"
		cornerTR := "┐"
		cornerBL := "└"
		cornerBR := "┘"
		side := "│"

		if selected {
			borderChar = "═"
			cornerTL = "╔"
			cornerTR = "╗"
			cornerBL = "╚"
			cornerBR = "╝"
			side = "║"
		}

		// Title bar
		title := note.Title
		if len(title) > width-10 {
			title = title[:width-13] + "..."
		}

		titleLine := fmt.Sprintf("%s─ %s ", cornerTL, title)
		padding := width - len(titleLine) - 1
		if padding < 0 {
			padding = 0
		}
		titleLine += strings.Repeat(borderChar, padding) + cornerTR
		fmt.Fprintln(v, titleLine)

		// Frontmatter if enabled
		if gui.state.Preview.ShowFrontmatter {
			fmt.Fprintf(v, "%s uuid: %s\n", side, note.UUID)
			fmt.Fprintf(v, "%s created: %s\n", side, note.Created.Format("2006-01-02"))
		}

		// Content preview (first few lines)
		content := note.Content
		if content == "" {
			content, _ = gui.loadNoteContent(note.Path)
		}

		contentLines := strings.Split(content, "\n")
		maxLines := 4
		for j, line := range contentLines {
			if j >= maxLines {
				fmt.Fprintf(v, "%s ...\n", side)
				break
			}
			if len(line) > width-4 {
				line = line[:width-7] + "..."
			}
			fmt.Fprintf(v, "%s %s\n", side, line)
		}

		// Footer
		tags := note.TagsString()
		date := note.ShortDate()
		footer := fmt.Sprintf("%s %s  %s", side, tags, date)
		fmt.Fprintln(v, footer)

		// Bottom border
		fmt.Fprintln(v, cornerBL+strings.Repeat(borderChar, width-2)+cornerBR)
		fmt.Fprintln(v, "")
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

	return string(data), nil
}
