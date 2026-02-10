package gui

import (
	"fmt"
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

	// Get usable content width (excludes frame)
	width := v.InnerWidth()
	if width < 10 {
		width = 30
	}

	// Check if this panel is focused for highlighting
	isActive := gui.state.CurrentContext == NotesContext

	for i, note := range gui.state.Notes.Items {
		selected := isActive && i == gui.state.Notes.SelectedIndex

		// Line 1 - Title (1-space prefix)
		title := note.Title
		if title == "" {
			title = note.Path
		}
		titleRunes := []rune(title)
		if len(titleRunes) > width-1 {
			title = strings.TrimRight(string(titleRunes[:width-4]), " ") + "..."
		}
		line1 := " " + title

		// Line 2 - First line of content (2-space prefix)
		snippet := note.FirstLine()
		snippetRunes := []rune(snippet)
		if len(snippetRunes) > width-2 {
			snippet = strings.TrimRight(string(snippetRunes[:width-5]), " ") + "..."
		}
		line2 := "  " + snippet

		// Line 3 - Date and tags (2-space prefix)
		date := note.ShortDate()
		tags := note.TagsString()
		maxTagLen := width - len(date) - 5 // "  " + " · "
		tagRunes := []rune(tags)
		if maxTagLen > 0 && len(tagRunes) > maxTagLen {
			tags = string(tagRunes[:maxTagLen-3]) + "..."
		}
		line3 := fmt.Sprintf("  %s · %s", date, tags)

		if selected {
			pad := func(s string) string {
				return s + strings.Repeat(" ", max(0, width-len(s)))
			}
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line1), AnsiReset)
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line2), AnsiReset)
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line3), AnsiReset)
		} else {
			fmt.Fprintln(v, line1)
			fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line2, AnsiReset)
			fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line3, AnsiReset)
		}

	}

	// Scroll to keep selection visible (3 lines per note)
	_, viewHeight := v.InnerSize()
	selLine := gui.state.Notes.SelectedIndex * 3
	scrollListView(v, selLine, 3, viewHeight)
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
			fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line2, AnsiReset)
		}
	}

	_, viewHeight := v.InnerSize()
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
			fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line2, AnsiReset)
		}
	}

	_, viewHeight := v.InnerSize()
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

	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	isActive := gui.state.CurrentContext == TagsContext

	for i, tag := range gui.state.Tags.Items {
		selected := isActive && i == gui.state.Tags.SelectedIndex

		name := tag.Name
		if len(name) > 0 && name[0] != '#' {
			name = "#" + name
		}
		count := fmt.Sprintf("(%d)", tag.Count)

		if selected {
			line := fmt.Sprintf(" %s %s", name, count)
			line = line + strings.Repeat(" ", max(0, width-len(line)))
			fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, line, AnsiReset)
		} else {
			fmt.Fprintf(v, " %s %s%s%s\n", name, AnsiDim, count, AnsiReset)
		}
	}

	// Scroll to keep selection visible (1 line per tag)
	_, viewHeight := v.InnerSize()
	selLine := gui.state.Tags.SelectedIndex
	scrollListView(v, selLine, 1, viewHeight)
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

// scrollViewport scrolls a list view's origin by delta lines without
// constraining the selection to stay visible. Keyboard navigation uses
// scrollListView instead, which does keep the selection on-screen.
func scrollViewport(v *gocui.View, delta int) {
	_, oy := v.Origin()
	newOy := oy + delta
	if newOy < 0 {
		newOy = 0
	}
	v.SetOrigin(0, newOy)
}

// listClickIndex returns the item index for a mouse click in a list view
// with the given item height (lines per item).
func listClickIndex(v *gocui.View, itemHeight int) int {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	return (cy + oy) / itemHeight
}
