package helpers

import (
	"fmt"
	"strings"
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"

	anytime "github.com/ijt/go-anytime"
)

// CalendarHelper encapsulates domain logic for the calendar dialog.
type CalendarHelper struct {
	c *HelperCommon
}

func NewCalendarHelper(c *HelperCommon) *CalendarHelper {
	return &CalendarHelper{c: c}
}

func (self *CalendarHelper) state() *context.CalendarState {
	return self.c.GuiCommon().Contexts().Calendar.State
}

// Open initializes calendar state and pushes the context.
func (self *CalendarHelper) Open() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}

	now := time.Now()
	if gui.Contexts().Calendar.State == nil {
		gui.Contexts().Calendar.State = &context.CalendarState{
			Year:        now.Year(),
			Month:       int(now.Month()),
			SelectedDay: now.Day(),
		}
	}

	self.RefreshNotes()
	gui.PushContextByKey("calendarGrid")
	return nil
}

// Close deletes calendar views and pops the context.
func (self *CalendarHelper) Close() {
	gui := self.c.GuiCommon()
	gui.DeleteView("calendarGrid")
	gui.DeleteView("calendarInput")
	gui.DeleteView("calendarNotes")
	gui.SetCursorEnabled(false)
	gui.PopContext()
}

// SelectedDate returns the currently selected date as YYYY-MM-DD.
func (self *CalendarHelper) SelectedDate() string {
	s := self.state()
	return fmt.Sprintf("%04d-%02d-%02d", s.Year, s.Month, s.SelectedDay)
}

// SelectedTime returns the selected date as a time.Time.
func (self *CalendarHelper) SelectedTime() time.Time {
	s := self.state()
	return time.Date(s.Year, time.Month(s.Month), s.SelectedDay, 0, 0, 0, 0, time.Local)
}

// RefreshNotes fetches notes for the currently selected date.
func (self *CalendarHelper) RefreshNotes() {
	s := self.state()
	s.Notes = self.fetchNotesForDate(self.SelectedDate())
	s.NoteIndex = 0
}

// MoveDay moves the selected day by delta days, crossing month boundaries.
func (self *CalendarHelper) MoveDay(delta int) {
	t := self.SelectedTime().AddDate(0, 0, delta)
	self.SetDate(t)
}

// SetDate sets the calendar to the given date directly.
func (self *CalendarHelper) SetDate(t time.Time) {
	s := self.state()
	s.Year = t.Year()
	s.Month = int(t.Month())
	s.SelectedDay = t.Day()
	self.RefreshNotes()
}

// Tab cycles focus forward: grid -> notes -> input.
func (self *CalendarHelper) Tab() error {
	s := self.state()
	s.Focus = (s.Focus + 1) % 3
	return nil
}

// Backtab cycles focus backward: grid -> input -> notes.
func (self *CalendarHelper) Backtab() error {
	s := self.state()
	s.Focus = (s.Focus + 2) % 3
	return nil
}

// FocusInput switches focus to the input view.
func (self *CalendarHelper) FocusInput() error {
	self.state().Focus = 2 // calFocusInput
	return nil
}

// NoteDown moves the note list selection down.
func (self *CalendarHelper) NoteDown() error {
	s := self.state()
	if s.NoteIndex < len(s.Notes)-1 {
		s.NoteIndex++
	}
	return nil
}

// NoteUp moves the note list selection up.
func (self *CalendarHelper) NoteUp() error {
	s := self.state()
	if s.NoteIndex > 0 {
		s.NoteIndex--
	}
	return nil
}

// NoteEnter loads the selected note in preview.
func (self *CalendarHelper) NoteEnter() error {
	s := self.state()
	if len(s.Notes) == 0 {
		return nil
	}
	self.LoadNoteInPreview(s.NoteIndex)
	return nil
}

// GridEnter loads all notes for the selected date into the preview.
func (self *CalendarHelper) GridEnter() error {
	self.LoadInPreview()
	return nil
}

// GridClick handles mouse clicks on the calendar grid to select a date.
func (self *CalendarHelper) GridClick() error {
	gui := self.c.GuiCommon()
	s := self.state()
	s.Focus = 0 // calFocusGrid

	v := gui.GetView("calendarGrid")
	if v == nil {
		return nil
	}

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
		self.RefreshNotes()
	} else {
		t := time.Date(s.Year, time.Month(s.Month), day, 0, 0, 0, 0, time.Local)
		self.SetDate(t)
	}

	return nil
}

// InputEnter parses the input and navigates to the date.
func (self *CalendarHelper) InputEnter() error {
	gui := self.c.GuiCommon()
	v := gui.GetView("calendarInput")
	if v == nil {
		return nil
	}

	raw := strings.TrimSpace(v.TextArea.GetContent())
	if raw == "" {
		self.state().Focus = 0 // calFocusGrid
		return nil
	}
	t, err := anytime.Parse(raw, time.Now())
	if err == nil {
		self.SetDate(t)
	}
	v.TextArea.Clear()
	v.Clear()
	self.state().Focus = 0 // calFocusGrid
	return nil
}

// InputEsc cancels input and returns to grid.
func (self *CalendarHelper) InputEsc() error {
	gui := self.c.GuiCommon()
	v := gui.GetView("calendarInput")
	if v == nil {
		return nil
	}
	v.TextArea.Clear()
	v.Clear()
	self.state().Focus = 0 // calFocusGrid
	return nil
}

// InputClick focuses the input view when clicked.
func (self *CalendarHelper) InputClick() error {
	gui := self.c.GuiCommon()
	self.state().Focus = 2 // calFocusInput
	v := gui.GetView("calendarInput")
	if v == nil {
		return nil
	}
	// Clear placeholder on focus
	v.Clear()
	v.RenderTextArea()
	return nil
}

// LoadInPreview loads all notes for the selected date into the preview.
func (self *CalendarHelper) LoadInPreview() {
	s := self.state()
	if len(s.Notes) == 0 {
		self.Close()
		return
	}

	notes, err := self.c.RuinCmd().Search.Search("created:"+self.SelectedDate(), commands.SearchOptions{
		Sort:           "created",
		Limit:          100,
		IncludeContent: true,
		StripTitle:     true,
	})
	if err != nil || len(notes) == 0 {
		self.Close()
		return
	}

	date := self.SelectedDate()
	gui := self.c.GuiCommon()
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.Close()
	self.c.Helpers().Preview().ShowCardList(" Calendar: "+date+" ", notes)
	gui.PushContextByKey("preview")
}

// LoadNoteInPreview loads a single note into the preview.
func (self *CalendarHelper) LoadNoteInPreview(index int) {
	s := self.state()
	if index >= len(s.Notes) {
		return
	}
	note := s.Notes[index]

	full, err := self.c.RuinCmd().Search.Get(note.UUID, commands.SearchOptions{
		IncludeContent: true,
		StripTitle:     true,
	})
	if err != nil || full == nil {
		return
	}

	title := full.Title
	gui := self.c.GuiCommon()
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.Close()
	self.c.Helpers().Preview().ShowCardList(" "+title+" ", []models.Note{*full})
	gui.PushContextByKey("preview")
}

// fetchNotesForDate fetches notes created on the given date (YYYY-MM-DD format).
func (self *CalendarHelper) fetchNotesForDate(date string) []models.Note {
	notes, err := self.c.RuinCmd().Search.Search("created:"+date, commands.SearchOptions{
		Sort:  "created",
		Limit: 100,
	})
	if err != nil {
		return nil
	}
	return notes
}

// daysIn returns the number of days in the given month.
func daysIn(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()
}
