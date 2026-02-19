package gui

import (
	"fmt"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
)

// Calendar focus: 0 = grid, 1 = notes, 2 = input
const (
	calFocusGrid  = 0
	calFocusNotes = 1
	calFocusInput = 2
)

// createCalendarViews creates the calendar grid, input, and note list views.
func (gui *Gui) createCalendarViews(g *gocui.Gui, maxX, maxY int) error {
	s := gui.contexts.Calendar.State
	totalWidth := 34
	gridHeight := 11 // border + padding + header + separator + 6 rows + padding(implicit) + border
	inputHeight := 3 // 2 content lines + shared border
	notesHeight := 12
	if totalWidth > maxX-4 {
		totalWidth = maxX - 4
	}

	totalHeight := gridHeight + inputHeight + notesHeight
	x0, y0, x1, _ := centerRect(maxX, maxY, totalWidth, totalHeight)

	// --- Input view (1 line high, at top) ---
	inputY1 := y0 + inputHeight
	iv, iErr := g.SetView(CalendarInputView, x0, y0, x1, inputY1, 0)
	if iErr != nil && iErr.Error() != "unknown view" {
		return iErr
	}
	iv.Editable = true
	iv.Editor = gocui.EditorFunc(gocui.SimpleEditor)
	iv.Wrap = false
	setRoundedCorners(iv)

	if s.Focus == calFocusInput {
		iv.FrameColor = gocui.ColorGreen
		iv.TitleColor = gocui.ColorGreen
	} else {
		iv.FrameColor = gocui.ColorDefault
		iv.TitleColor = gocui.ColorDefault
	}

	// Show dimmed placeholder when empty and not focused
	if s.Focus != calFocusInput && iv.TextArea.GetContent() == "" {
		iv.Clear()
		fmt.Fprintf(iv, "%s/ jump to date%s", AnsiDim, AnsiReset)
	}

	g.SetViewOnTop(CalendarInputView)

	// --- Grid view ---
	gridY1 := inputY1 + gridHeight
	gv, err := g.SetView(CalendarGridView, x0, inputY1, x1, gridY1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	monthName := time.Month(s.Month).String()
	gv.Title = fmt.Sprintf(" %s %d ", monthName, s.Year)
	selectedTime := gui.helpers.Calendar().SelectedTime()
	gv.Footer = fmt.Sprintf(" %s ", selectedTime.Format("Mon, Jan 02"))
	gv.Editable = false
	setRoundedCorners(gv)

	if s.Focus == calFocusGrid {
		gv.FrameColor = gocui.ColorGreen
		gv.TitleColor = gocui.ColorGreen
	} else {
		gv.FrameColor = gocui.ColorDefault
		gv.TitleColor = gocui.ColorDefault
	}

	gui.renderCalendarGrid(gv)
	g.SetViewOnTop(CalendarGridView)

	// --- Notes view ---
	notesY0 := gridY1
	notesY1 := notesY0 + notesHeight
	if notesY1 >= maxY {
		notesY1 = maxY - 1
	}

	nv, nErr := g.SetView(CalendarNotesView, x0, notesY0, x1, notesY1, 0)
	if nErr != nil && nErr.Error() != "unknown view" {
		return nErr
	}

	noteCount := len(s.Notes)
	if noteCount == 1 {
		nv.Title = " 1 note "
	} else {
		nv.Title = fmt.Sprintf(" %d notes ", noteCount)
	}
	setRoundedCorners(nv)

	if s.Focus == calFocusNotes {
		nv.FrameColor = gocui.ColorGreen
		nv.TitleColor = gocui.ColorGreen
	} else {
		nv.FrameColor = gocui.ColorDefault
		nv.TitleColor = gocui.ColorDefault
	}

	renderDateNoteList(nv, s.Notes, s.NoteIndex, s.Focus == calFocusNotes)

	g.SetViewOnTop(CalendarNotesView)

	// Set focus
	switch s.Focus {
	case calFocusGrid:
		g.Cursor = false
		g.SetCurrentView(CalendarGridView)
	case calFocusInput:
		g.Cursor = true
		g.SetCurrentView(CalendarInputView)
	case calFocusNotes:
		g.Cursor = false
		g.SetCurrentView(CalendarNotesView)
	}

	return nil
}

// renderCalendarGrid renders the month grid into the view.
func (gui *Gui) renderCalendarGrid(v *gocui.View) {
	v.Clear()
	s := gui.contexts.Calendar.State
	now := time.Now()
	today := now.Day()
	todayMonth := int(now.Month())
	todayYear := now.Year()

	// The grid is 22 visible chars wide: 1 leading space + 7 columns * 3 chars
	gridWidth := 22
	innerWidth, _ := v.InnerSize()
	leftPad := strings.Repeat(" ", max(0, (innerWidth-gridWidth)/2))

	// Top padding
	fmt.Fprintln(v)

	// Header
	fmt.Fprintf(v, "%s Su Mo Tu We Th Fr Sa\n", leftPad)

	// Separator
	fmt.Fprintf(v, "%s %s%s%s\n", leftPad, AnsiDim, strings.Repeat("─", gridWidth-1), AnsiReset)

	// First day of month
	first := time.Date(s.Year, time.Month(s.Month), 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(first.Weekday()) // 0=Sunday
	daysInMonth := daysIn(s.Year, s.Month)

	// Previous month days for filling
	prevMonthDays := 0
	if s.Month == 1 {
		prevMonthDays = daysIn(s.Year-1, 12)
	} else {
		prevMonthDays = daysIn(s.Year, s.Month-1)
	}

	// Render 6 rows
	day := 1
	nextMonthDay := 1
	for row := range 6 {
		var line strings.Builder
		line.WriteString(leftPad)
		line.WriteString(" ")
		for col := range 7 {
			cellIdx := row*7 + col
			if cellIdx < startWeekday {
				d := prevMonthDays - startWeekday + cellIdx + 1
				fmt.Fprintf(&line, "%s%3d%s", AnsiDim, d, AnsiReset)
			} else if day <= daysInMonth {
				if day == s.SelectedDay {
					fmt.Fprintf(&line, "%s%3d%s", AnsiBlueBgWhite, day, AnsiReset)
				} else if day == today && s.Month == todayMonth && s.Year == todayYear {
					fmt.Fprintf(&line, "%s%3d%s", AnsiBoldWhite, day, AnsiReset)
				} else {
					fmt.Fprintf(&line, "%3d", day)
				}
				day++
			} else {
				fmt.Fprintf(&line, "%s%3d%s", AnsiDim, nextMonthDay, AnsiReset)
				nextMonthDay++
			}
		}
		fmt.Fprintln(v, line.String())
	}
}

// daysIn returns the number of days in the given month.
func daysIn(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()
}
