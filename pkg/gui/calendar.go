package gui

import (
	"fmt"
	"strings"
	"time"

	anytime "github.com/ijt/go-anytime"
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"
)

// Calendar focus: 0 = grid, 1 = notes, 2 = input
const (
	calFocusGrid  = 0
	calFocusNotes = 1
	calFocusInput = 2
)

// openCalendar opens the calendar dialog.
func (gui *Gui) openCalendar(g *gocui.Gui, v *gocui.View) error {
	if !gui.openOverlay(OverlayCalendar) {
		return nil
	}

	now := time.Now()
	if gui.state.Calendar == nil {
		gui.state.Calendar = &CalendarState{
			Year:        now.Year(),
			Month:       int(now.Month()),
			SelectedDay: now.Day(),
		}
	}

	gui.calendarRefreshNotes()
	return nil
}

// closeCalendar closes the calendar dialog.
func (gui *Gui) closeCalendar() {
	gui.closeOverlay()
	gui.g.DeleteView(CalendarGridView)
	gui.g.DeleteView(CalendarInputView)
	gui.g.DeleteView(CalendarNotesView)
	gui.g.Cursor = false
	gui.g.SetCurrentView(gui.contextToView(gui.state.currentContext()))
}

// calendarSelectedDate returns the currently selected date as YYYY-MM-DD.
func (gui *Gui) calendarSelectedDate() string {
	s := gui.state.Calendar
	return fmt.Sprintf("%04d-%02d-%02d", s.Year, s.Month, s.SelectedDay)
}

// calendarRefreshNotes fetches notes for the currently selected date.
func (gui *Gui) calendarRefreshNotes() {
	s := gui.state.Calendar
	s.Notes = gui.fetchNotesForDate(gui.calendarSelectedDate())
	s.NoteIndex = 0
}

// calendarSelectedTime returns the selected date as a time.Time.
func (gui *Gui) calendarSelectedTime() time.Time {
	s := gui.state.Calendar
	return time.Date(s.Year, time.Month(s.Month), s.SelectedDay, 0, 0, 0, 0, time.Local)
}

// calendarMoveDay moves the selected day by delta days, crossing month boundaries.
func (gui *Gui) calendarMoveDay(delta int) {
	t := gui.calendarSelectedTime().AddDate(0, 0, delta)
	s := gui.state.Calendar
	s.Year = t.Year()
	s.Month = int(t.Month())
	s.SelectedDay = t.Day()
	gui.calendarRefreshNotes()
}

// calendarSetDate sets the calendar to the given date directly.
func (gui *Gui) calendarSetDate(t time.Time) {
	s := gui.state.Calendar
	s.Year = t.Year()
	s.Month = int(t.Month())
	s.SelectedDay = t.Day()
	gui.calendarRefreshNotes()
}

// createCalendarViews creates the calendar grid, input, and note list views.
func (gui *Gui) createCalendarViews(g *gocui.Gui, maxX, maxY int) error {
	s := gui.state.Calendar
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
	selectedTime := gui.calendarSelectedTime()
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
	s := gui.state.Calendar
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

// calendarGridClick handles mouse clicks on the calendar grid to select a date.
func (gui *Gui) calendarGridClick(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Calendar
	s.Focus = calFocusGrid

	_, cy := v.Cursor()
	_, oy := v.Origin()
	row := cy + oy

	// Rows: 0 = padding, 1 = header, 2 = separator, 3-8 = week rows
	if row < 3 || row > 8 {
		return nil
	}

	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	absX := cx + ox

	innerWidth, _ := v.InnerSize()
	gridWidth := 22
	leftPadLen := max(0, (innerWidth-gridWidth)/2)

	contentX := absX - leftPadLen - 1
	if contentX < 0 || contentX >= 21 {
		return nil
	}

	col := contentX / 3
	if col > 6 {
		col = 6
	}

	weekRow := row - 3
	first := time.Date(s.Year, time.Month(s.Month), 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(first.Weekday())
	daysInMonth := daysIn(s.Year, s.Month)

	cellIdx := weekRow*7 + col
	day := cellIdx - startWeekday + 1

	if day >= 1 && day <= daysInMonth {
		s.SelectedDay = day
		gui.calendarRefreshNotes()
	} else {
		t := time.Date(s.Year, time.Month(s.Month), day, 0, 0, 0, 0, time.Local)
		gui.calendarSetDate(t)
	}

	return nil
}

// calendarInputClick focuses the input view when clicked.
func (gui *Gui) calendarInputClick(g *gocui.Gui, v *gocui.View) error {
	gui.state.Calendar.Focus = calFocusInput
	// Clear placeholder on focus
	v.Clear()
	v.RenderTextArea()
	return nil
}

// Calendar keybinding handlers

func (gui *Gui) calendarGridLeft(g *gocui.Gui, v *gocui.View) error {
	gui.calendarMoveDay(-1)
	return nil
}

func (gui *Gui) calendarGridRight(g *gocui.Gui, v *gocui.View) error {
	gui.calendarMoveDay(1)
	return nil
}

func (gui *Gui) calendarGridUp(g *gocui.Gui, v *gocui.View) error {
	gui.calendarMoveDay(-7)
	return nil
}

func (gui *Gui) calendarGridDown(g *gocui.Gui, v *gocui.View) error {
	gui.calendarMoveDay(7)
	return nil
}

func (gui *Gui) calendarGridEnter(g *gocui.Gui, v *gocui.View) error {
	gui.calendarLoadInPreview()
	return nil
}

func (gui *Gui) calendarEsc(g *gocui.Gui, v *gocui.View) error {
	gui.closeCalendar()
	return nil
}

func (gui *Gui) calendarTab(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Calendar
	s.Focus = (s.Focus + 1) % 3
	return nil
}

func (gui *Gui) calendarBacktab(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Calendar
	s.Focus = (s.Focus + 2) % 3
	return nil
}

func (gui *Gui) calendarFocusInput(g *gocui.Gui, v *gocui.View) error {
	gui.state.Calendar.Focus = calFocusInput
	return nil
}

func (gui *Gui) calendarNoteDown(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Calendar
	if s.NoteIndex < len(s.Notes)-1 {
		s.NoteIndex++
	}
	return nil
}

func (gui *Gui) calendarNoteUp(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Calendar
	if s.NoteIndex > 0 {
		s.NoteIndex--
	}
	return nil
}

func (gui *Gui) calendarNoteEnter(g *gocui.Gui, v *gocui.View) error {
	s := gui.state.Calendar
	if len(s.Notes) == 0 {
		return nil
	}
	gui.calendarLoadNoteInPreview(s.NoteIndex)
	return nil
}

// calendarInputEnter parses the input and navigates to the date.
func (gui *Gui) calendarInputEnter(g *gocui.Gui, v *gocui.View) error {
	raw := strings.TrimSpace(v.TextArea.GetContent())
	if raw == "" {
		gui.state.Calendar.Focus = calFocusGrid
		return nil
	}
	t, err := anytime.Parse(raw, time.Now())
	if err == nil {
		gui.calendarSetDate(t)
	}
	v.TextArea.Clear()
	v.Clear()
	gui.state.Calendar.Focus = calFocusGrid
	return nil
}

// calendarInputEsc cancels input and returns to grid.
func (gui *Gui) calendarInputEsc(g *gocui.Gui, v *gocui.View) error {
	v.TextArea.Clear()
	v.Clear()
	gui.state.Calendar.Focus = calFocusGrid
	return nil
}

// calendarLoadInPreview loads all notes for the selected date into the preview.
func (gui *Gui) calendarLoadInPreview() {
	s := gui.state.Calendar
	if len(s.Notes) == 0 {
		gui.closeCalendar()
		return
	}

	notes, err := gui.ruinCmd.Search.Search("created:"+gui.calendarSelectedDate(), commands.SearchOptions{
		Sort:           "created",
		Limit:          100,
		IncludeContent: true,
		StripTitle:     true,
	})
	if err != nil || len(notes) == 0 {
		gui.closeCalendar()
		return
	}

	date := gui.calendarSelectedDate()
	gui.preview.pushNavHistory()
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.ScrollOffset = 0
	gui.state.Preview.Mode = PreviewModeCardList
	gui.closeCalendar()
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Calendar: " + date + " "
	}
	gui.setContext(PreviewContext)
	gui.renderPreview()
}

// calendarLoadNoteInPreview loads a single note into the preview.
func (gui *Gui) calendarLoadNoteInPreview(index int) {
	s := gui.state.Calendar
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
	gui.preview.pushNavHistory()
	gui.state.Preview.Cards = []models.Note{*full}
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.ScrollOffset = 0
	gui.state.Preview.Mode = PreviewModeCardList
	gui.closeCalendar()
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " " + title + " "
	}
	gui.setContext(PreviewContext)
	gui.renderPreview()
}
