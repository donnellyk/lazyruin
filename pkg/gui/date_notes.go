package gui

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// renderDateNoteList renders a note list into a gocui view using the standard
// 3-line format (title, snippet, date+tags) with selection highlighting.
func renderDateNoteList(v *gocui.View, notes []models.Note, selectedIndex int, focused bool) {
	v.Clear()

	if len(notes) == 0 {
		fmt.Fprintln(v, " No notes for this date")
		return
	}

	width, _ := v.InnerSize()
	if width < 10 {
		width = 30
	}

	pad := func(s string) string {
		return s + strings.Repeat(" ", max(0, width-len([]rune(s))))
	}

	for i, note := range notes {
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

		lines := []string{
			" " + title,
			"  " + snippet,
			"  " + meta,
		}

		selected := focused && i == selectedIndex
		if selected {
			for _, line := range lines {
				fmt.Fprintf(v, "%s%s%s\n", AnsiBlueBgWhite, pad(line), AnsiReset)
			}
		} else {
			fmt.Fprintln(v, lines[0])
			for _, line := range lines[1:] {
				fmt.Fprintf(v, "%s%s%s\n", AnsiDim, line, AnsiReset)
			}
		}
	}

	_, viewHeight := v.InnerSize()
	selLine := selectedIndex * 3
	scrollListView(v, selLine, 3, viewHeight)
}
