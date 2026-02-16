package gui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/models"
)

// listItem holds the formatted lines for a single list item.
// Lines[0] is always rendered plain when selected, Lines[1:] are dim when unselected.
// Lines must not contain ANSI codes that would conflict with selection highlighting.
type listItem struct {
	Lines []string
}

// renderList is a shared helper for rendering list panels with selection highlighting.
// It handles clear, empty state, per-item formatting via the builder callback,
// selection highlighting with blue background, and scroll management.
// The builder receives the item index and whether it's currently selected.
func renderList(v *gocui.View, itemCount int, selectedIndex int, isActive bool, linesPerItem int, emptyMsg string, builder func(index int, selected bool) listItem) {
	v.Clear()

	if itemCount == 0 {
		fmt.Fprintln(v, emptyMsg)
		return
	}

	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	pad := func(s string) string {
		return s + strings.Repeat(" ", max(0, width-len([]rune(s))))
	}

	for i := range itemCount {
		selected := isActive && i == selectedIndex
		item := builder(i, selected)

		if selected {
			for _, line := range item.Lines {
				fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line), AnsiReset)
			}
		} else {
			for j, line := range item.Lines {
				if j == 0 {
					fmt.Fprintln(v, line)
				} else {
					fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line, AnsiReset)
				}
			}
		}
	}

	_, viewHeight := v.InnerSize()
	selLine := selectedIndex * linesPerItem
	scrollListView(v, selLine, linesPerItem, viewHeight)
}

func (gui *Gui) renderNotes() {
	v := gui.views.Notes
	if v == nil {
		return
	}

	width := v.InnerWidth()
	if width < 10 {
		width = 30
	}

	renderList(v, len(gui.state.Notes.Items), gui.state.Notes.SelectedIndex,
		gui.state.CurrentContext == NotesContext, 3,
		"\n No notes found.\n\n Press 'n' to create a new note\n or '/' to search",
		func(i int, _ bool) listItem {
			note := gui.state.Notes.Items[i]

			title := note.Title
			if title == "" {
				title = note.Path
			}
			titleRunes := []rune(title)
			if len(titleRunes) > width-1 {
				title = strings.TrimRight(string(titleRunes[:width-4]), " ") + "..."
			}

			snippet := note.FirstLine()
			snippetRunes := []rune(snippet)
			if len(snippetRunes) > width-2 {
				snippet = strings.TrimRight(string(snippetRunes[:width-5]), " ") + "..."
			}

			meta := models.JoinDot(note.ShortDate(), note.TagsString())
			metaRunes := []rune(meta)
			if len(metaRunes) > width-3 {
				meta = string(metaRunes[:width-6]) + "..."
			}

			return listItem{Lines: []string{
				" " + title,
				"  " + snippet,
				"  " + meta,
			}}
		})
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

	renderList(v, len(gui.state.Queries.Items), gui.state.Queries.SelectedIndex,
		gui.state.CurrentContext == QueriesContext, 2,
		" No saved queries.",
		func(i int, _ bool) listItem {
			query := gui.state.Queries.Items[i]
			queryStr := query.Query
			if len(queryStr) > 25 {
				queryStr = queryStr[:22] + "..."
			}
			return listItem{Lines: []string{
				"  " + query.Name,
				"    " + queryStr,
			}}
		})
}

func (gui *Gui) renderParents() {
	v := gui.views.Queries
	if v == nil {
		return
	}

	width, _ := v.Size()
	if width < 10 {
		width = 30
	}

	renderList(v, len(gui.state.Parents.Items), gui.state.Parents.SelectedIndex,
		gui.state.CurrentContext == QueriesContext, 2,
		" No parent bookmarks.",
		func(i int, _ bool) listItem {
			parent := gui.state.Parents.Items[i]
			title := parent.Title
			if len(title) > width-6 {
				title = title[:width-9] + "..."
			}
			return listItem{Lines: []string{
				"  " + parent.Name,
				"    " + title,
			}}
		})
}

func (gui *Gui) renderTags() {
	v := gui.views.Tags
	if v == nil {
		return
	}

	items := gui.filteredTagItems()
	renderList(v, len(items), gui.state.Tags.SelectedIndex,
		gui.state.CurrentContext == TagsContext, 1,
		" No tags found.",
		func(i int, selected bool) listItem {
			tag := items[i]
			name := tag.Name
			if len(name) > 0 && name[0] != '#' {
				name = "#" + name
			}
			count := fmt.Sprintf("(%d)", tag.Count)
			if selected {
				return listItem{Lines: []string{
					fmt.Sprintf(" %s %s", name, count),
				}}
			}
			return listItem{Lines: []string{
				fmt.Sprintf(" %s %s%s%s", name, AnsiDim, count, AnsiReset),
			}}
		})
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
