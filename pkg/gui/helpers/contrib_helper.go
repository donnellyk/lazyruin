package helpers

import (
	"fmt"
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"
)

// ContribHelper encapsulates domain logic for the contribution chart dialog.
type ContribHelper struct {
	c *HelperCommon
}

func NewContribHelper(c *HelperCommon) *ContribHelper {
	return &ContribHelper{c: c}
}

func (self *ContribHelper) state() *context.ContribState {
	return self.c.GuiCommon().Contexts().Contrib.State
}

// Open initializes contrib state and pushes the context.
func (self *ContribHelper) Open() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}

	now := time.Now()
	if gui.Contexts().Contrib.State == nil {
		gui.Contexts().Contrib.State = &context.ContribState{
			SelectedDate: now.Format("2006-01-02"),
		}
	}

	self.LoadData()
	self.RefreshNotes()
	gui.PushContextByKey("contribGrid")
	return nil
}

// Close deletes contrib views and pops the context.
func (self *ContribHelper) Close() {
	gui := self.c.GuiCommon()
	gui.DeleteView("contribGrid")
	gui.DeleteView("contribNotes")
	gui.PopContext()
}

// LoadData loads note counts for the past year.
func (self *ContribHelper) LoadData() {
	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	query := fmt.Sprintf("between:%s,%s", start.Format("2006-01-02"), now.Format("2006-01-02"))

	notes, err := self.c.RuinCmd().Search.Search(query, commands.SearchOptions{
		Limit: 5000,
	})
	if err != nil {
		self.state().DayCounts = make(map[string]int)
		return
	}

	counts := make(map[string]int)
	for _, n := range notes {
		day := n.Created.Format("2006-01-02")
		counts[day]++
	}
	self.state().DayCounts = counts
}

// RefreshNotes fetches notes for the selected date.
func (self *ContribHelper) RefreshNotes() {
	s := self.state()
	s.Notes = self.fetchNotesForDate(s.SelectedDate)
	s.NoteIndex = 0
}

// MoveDay moves the selected date by delta days.
func (self *ContribHelper) MoveDay(delta int) {
	s := self.state()
	t, _ := time.ParseInLocation("2006-01-02", s.SelectedDate, time.Local)
	t = t.AddDate(0, 0, delta)
	s.SelectedDate = t.Format("2006-01-02")
	self.RefreshNotes()
}

// Tab toggles focus between grid and notes.
func (self *ContribHelper) Tab() error {
	s := self.state()
	if s.Focus == 0 {
		s.Focus = 1
	} else {
		s.Focus = 0
	}
	return nil
}

// NoteDown moves the note list selection down.
func (self *ContribHelper) NoteDown() error {
	s := self.state()
	if s.NoteIndex < len(s.Notes)-1 {
		s.NoteIndex++
	}
	return nil
}

// NoteUp moves the note list selection up.
func (self *ContribHelper) NoteUp() error {
	s := self.state()
	if s.NoteIndex > 0 {
		s.NoteIndex--
	}
	return nil
}

// NoteEnter loads the selected note in preview.
func (self *ContribHelper) NoteEnter() error {
	s := self.state()
	if len(s.Notes) == 0 {
		return nil
	}
	self.LoadNoteInPreview(s.NoteIndex)
	return nil
}

// GridEnter loads all notes for the selected date into the preview.
func (self *ContribHelper) GridEnter() error {
	self.LoadInPreview()
	return nil
}

// LoadInPreview loads all notes for the selected date into preview.
func (self *ContribHelper) LoadInPreview() {
	s := self.state()
	if len(s.Notes) == 0 {
		self.Close()
		return
	}

	notes, err := self.c.RuinCmd().Search.Search("created:"+s.SelectedDate, commands.SearchOptions{
		Sort:           "created",
		Limit:          100,
		IncludeContent: true,
		StripTitle:     true,
	})
	if err != nil || len(notes) == 0 {
		self.Close()
		return
	}

	date := s.SelectedDate
	gui := self.c.GuiCommon()
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.Close()
	self.c.Helpers().Preview().ShowCardList(" Contrib: "+date+" ", notes)
	gui.PushContextByKey("preview")
}

// LoadNoteInPreview loads a single note into preview.
func (self *ContribHelper) LoadNoteInPreview(index int) {
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
func (self *ContribHelper) fetchNotesForDate(date string) []models.Note {
	notes, err := self.c.RuinCmd().Search.Search("created:"+date, commands.SearchOptions{
		Sort:  "created",
		Limit: 100,
	})
	if err != nil {
		return nil
	}
	return notes
}
