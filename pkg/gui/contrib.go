package gui

import (
	"fmt"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"
)

// openContrib opens the contribution chart dialog.
func (gui *Gui) openContrib(g *gocui.Gui, v *gocui.View) error {
	if gui.state.popupActive() {
		return nil
	}

	now := time.Now()
	if gui.state.Contrib == nil {
		gui.state.Contrib = &ContribState{
			SelectedDate: now.Format("2006-01-02"),
		}
	}

	gui.contribLoadData()
	gui.contribRefreshNotes()
	gui.pushContext(ContribCtx)
	return nil
}

// closeContrib closes the contribution chart dialog.
func (gui *Gui) closeContrib() {
	gui.g.DeleteView(ContribGridView)
	gui.g.DeleteView(ContribNotesView)
	gui.popContext()
}

// contribLoadData loads note counts for the past year.
func (gui *Gui) contribLoadData() {
	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	query := fmt.Sprintf("between:%s,%s", start.Format("2006-01-02"), now.Format("2006-01-02"))

	notes, err := gui.ruinCmd.Search.Search(query, commands.SearchOptions{
		Limit: 5000,
	})
	if err != nil {
		gui.state.Contrib.DayCounts = make(map[string]int)
		return
	}

	counts := make(map[string]int)
	for _, n := range notes {
		day := n.Created.Format("2006-01-02")
		counts[day]++
	}
	gui.state.Contrib.DayCounts = counts
}

// contribRefreshNotes fetches notes for the selected date.
func (gui *Gui) contribRefreshNotes() {
	s := gui.state.Contrib
	s.Notes = gui.fetchNotesForDate(s.SelectedDate)
	s.NoteIndex = 0
}

// contribMoveDay moves the selected date by delta days.
func (gui *Gui) contribMoveDay(delta int) {
	s := gui.state.Contrib
	t, _ := time.ParseInLocation("2006-01-02", s.SelectedDate, time.Local)
	t = t.AddDate(0, 0, delta)
	s.SelectedDate = t.Format("2006-01-02")
	gui.contribRefreshNotes()
}

// createContribViews creates the contribution chart views.
func (gui *Gui) createContribViews(g *gocui.Gui, maxX, maxY int) error {
	s := gui.state.Contrib

	// Calculate width based on available space
	// Each cell is 2 chars wide (◼ + space), plus 5 for row labels, plus 2 for borders
	weekCols := (maxX - 10 - 2 - 5) / 2 // available for cells
	if weekCols > 52 {
		weekCols = 52
	}
	if weekCols < 10 {
		weekCols = 10
	}
	s.WeekCount = weekCols
	gridWidth := weekCols*2 + 5 + 2 // cells + labels + borders

	gridHeight := 11 // 1 month header + 7 day rows + 1 legend + border
	notesHeight := 12

	x0, y0, x1, _ := centerRect(maxX, maxY, gridWidth, gridHeight+notesHeight)
	gridY1 := y0 + gridHeight

	// Grid view
	gv, err := g.SetView(ContribGridView, x0, y0, x1, gridY1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	gv.Title = " Contributions "
	t, _ := time.ParseInLocation("2006-01-02", s.SelectedDate, time.Local)
	noteCount := len(s.Notes)
	gv.Footer = fmt.Sprintf(" %s · %d notes ", t.Format("Mon, Jan 02"), noteCount)
	setRoundedCorners(gv)

	if s.Focus == 0 {
		gv.FrameColor = gocui.ColorGreen
		gv.TitleColor = gocui.ColorGreen
	} else {
		gv.FrameColor = gocui.ColorDefault
		gv.TitleColor = gocui.ColorDefault
	}

	gui.renderContribGrid(gv)

	g.SetViewOnTop(ContribGridView)

	// Notes view
	notesY0 := gridY1
	notesY1 := notesY0 + notesHeight
	if notesY1 >= maxY {
		notesY1 = maxY - 1
	}

	nv, nErr := g.SetView(ContribNotesView, x0, notesY0, x1, notesY1, 0)
	if nErr != nil && nErr.Error() != "unknown view" {
		return nErr
	}

	if noteCount == 1 {
		nv.Title = " 1 note "
	} else {
		nv.Title = fmt.Sprintf(" %d notes ", noteCount)
	}
	setRoundedCorners(nv)

	if s.Focus == 1 {
		nv.FrameColor = gocui.ColorGreen
		nv.TitleColor = gocui.ColorGreen
	} else {
		nv.FrameColor = gocui.ColorDefault
		nv.TitleColor = gocui.ColorDefault
	}

	renderDateNoteList(nv, s.Notes, s.NoteIndex, s.Focus == 1)

	g.SetViewOnTop(ContribNotesView)

	if s.Focus == 0 {
		g.SetCurrentView(ContribGridView)
	} else {
		g.SetCurrentView(ContribNotesView)
	}

	return nil
}

// renderContribGrid renders the contribution heatmap grid.
func (gui *Gui) renderContribGrid(v *gocui.View) {
	v.Clear()
	s := gui.state.Contrib

	now := time.Now()
	// End date is end of current week (Saturday)
	endDate := now
	for endDate.Weekday() != time.Saturday {
		endDate = endDate.AddDate(0, 0, 1)
	}

	// Start date is weekCount weeks back from endDate
	startDate := endDate.AddDate(0, 0, -(s.WeekCount-1)*7)
	// Align to Sunday
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	// Build a grid: weeks[weekIdx][dayOfWeek] = date
	type cell struct {
		date  string
		count int
	}
	weeks := make([][]cell, s.WeekCount)
	d := startDate
	for w := range s.WeekCount {
		weeks[w] = make([]cell, 7)
		for dow := range 7 {
			dateStr := d.Format("2006-01-02")
			weeks[w][dow] = cell{date: dateStr, count: s.DayCounts[dateStr]}
			d = d.AddDate(0, 0, 1)
		}
	}

	// Month labels row (each column is 2 chars wide)
	var monthLine strings.Builder
	monthLine.WriteString("     ") // row label padding
	lastMonth := -1
	charsWritten := 0
	for w := range s.WeekCount {
		colPos := w * 2
		if charsWritten > colPos {
			// Previous label still occupies this column
			continue
		}
		dt, _ := time.ParseInLocation("2006-01-02", weeks[w][0].date, time.Local)
		m := int(dt.Month())
		if m != lastMonth && dt.Day() <= 7 {
			name := dt.Format("Jan")
			// Pad to reach this column position
			if charsWritten < colPos {
				monthLine.WriteString(strings.Repeat(" ", colPos-charsWritten))
				charsWritten = colPos
			}
			monthLine.WriteString(name)
			charsWritten += len(name)
			lastMonth = m
		}
	}
	fmt.Fprintln(v, monthLine.String())

	// Day rows (Sunday=0 through Saturday=6)
	dayLabels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for dow := range 7 {
		var line strings.Builder
		// Row label: show Mon, Wed, Fri
		switch dow {
		case 1:
			line.WriteString(" Mon ")
		case 3:
			line.WriteString(" Wed ")
		case 5:
			line.WriteString(" Fri ")
		default:
			_ = dayLabels // suppress unused
			line.WriteString("     ")
		}

		for w := range s.WeekCount {
			c := weeks[w][dow]
			// Don't show future dates
			dt, _ := time.ParseInLocation("2006-01-02", c.date, time.Local)
			if dt.After(now) {
				line.WriteString("  ")
				continue
			}

			if c.date == s.SelectedDate {
				fmt.Fprintf(&line, "%s◼%s ", AnsiBlueBgWhite, AnsiReset)
			} else {
				line.WriteString(contribChar(c.count))
			}
		}
		fmt.Fprintln(v, line.String())
	}

	// Legend row
	fmt.Fprintf(v, "  %s◼%s = 0  %s◼%s = 1  %s◼%s = 2  %s◼%s = 3+\n",
		AnsiDim, AnsiReset,
		AnsiGreen1, AnsiReset,
		AnsiGreen2, AnsiReset,
		AnsiGreen3, AnsiReset,
	)
}

// contribChar returns the colored block character (with trailing space) for a given note count.
func contribChar(count int) string {
	switch count {
	case 0:
		return fmt.Sprintf("%s◼%s ", AnsiDim, AnsiReset)
	case 1:
		return fmt.Sprintf("%s◼%s ", AnsiGreen1, AnsiReset)
	case 2:
		return fmt.Sprintf("%s◼%s ", AnsiGreen2, AnsiReset)
	default:
		return fmt.Sprintf("%s◼%s ", AnsiGreen3, AnsiReset)
	}
}

// Contribution chart keybinding handlers

func (gui *Gui) contribGridLeft(g *gocui.Gui, v *gocui.View) error {
	gui.contribMoveDay(-7) // left = prev week (column)
	return nil
}

func (gui *Gui) contribGridRight(g *gocui.Gui, v *gocui.View) error {
	gui.contribMoveDay(7) // right = next week (column)
	return nil
}

func (gui *Gui) contribGridUp(g *gocui.Gui, v *gocui.View) error {
	gui.contribMoveDay(-1) // up = prev day (row)
	return nil
}

func (gui *Gui) contribGridDown(g *gocui.Gui, v *gocui.View) error {
	gui.contribMoveDay(1) // down = next day (row)
	return nil
}

func (gui *Gui) contribGridEnter(g *gocui.Gui, v *gocui.View) error {
	gui.contribLoadInPreview()
	return nil
}

func (gui *Gui) contribEsc(g *gocui.Gui, v *gocui.View) error {
	gui.closeContrib()
	return nil
}

func (gui *Gui) contribTab(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Contrib
	if s.Focus == 0 {
		s.Focus = 1
	} else {
		s.Focus = 0
	}
	return nil
}

func (gui *Gui) contribNoteDown(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Contrib
	if s.NoteIndex < len(s.Notes)-1 {
		s.NoteIndex++
	}
	return nil
}

func (gui *Gui) contribNoteUp(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Contrib
	if s.NoteIndex > 0 {
		s.NoteIndex--
	}
	return nil
}

func (gui *Gui) contribNoteEnter(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Contrib
	if len(s.Notes) == 0 {
		return nil
	}
	gui.contribLoadNoteInPreview(s.NoteIndex)
	return nil
}

// contribLoadInPreview loads all notes for the selected date into preview.
func (gui *Gui) contribLoadInPreview() {
	s := gui.state.Contrib
	if len(s.Notes) == 0 {
		gui.closeContrib()
		return
	}

	notes, err := gui.ruinCmd.Search.Search("created:"+s.SelectedDate, commands.SearchOptions{
		Sort:           "created",
		Limit:          100,
		IncludeContent: true,
		StripTitle:     true,
	})
	if err != nil || len(notes) == 0 {
		gui.closeContrib()
		return
	}

	date := s.SelectedDate
	gui.helpers.Preview().PushNavHistory()
	gui.contexts.Preview.Cards = notes
	gui.contexts.Preview.SelectedCardIndex = 0
	gui.contexts.Preview.ScrollOffset = 0
	gui.contexts.Preview.Mode = PreviewModeCardList
	gui.closeContrib()
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Contrib: " + date + " "
	}
	gui.setContext(PreviewContext)
	gui.renderPreview()
}

// contribLoadNoteInPreview loads a single note into preview.
func (gui *Gui) contribLoadNoteInPreview(index int) {
	s := gui.state.Contrib
	if index >= len(s.Notes) {
		return
	}
	note := s.Notes[index]

	full, err := gui.ruinCmd.Search.Get(note.UUID, commands.SearchOptions{
		IncludeContent: true,
		StripTitle:     true,
	})
	if err != nil || full == nil {
		return
	}

	title := full.Title
	gui.helpers.Preview().PushNavHistory()
	gui.contexts.Preview.Cards = []models.Note{*full}
	gui.contexts.Preview.SelectedCardIndex = 0
	gui.contexts.Preview.ScrollOffset = 0
	gui.contexts.Preview.Mode = PreviewModeCardList
	gui.closeContrib()
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " " + title + " "
	}
	gui.setContext(PreviewContext)
	gui.renderPreview()
}
