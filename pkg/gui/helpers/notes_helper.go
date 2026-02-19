package helpers

import (
	"time"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"
)

// NotesHelper encapsulates note domain operations.
type NotesHelper struct {
	c *HelperCommon
}

// NewNotesHelper creates a new NotesHelper.
func NewNotesHelper(c *HelperCommon) *NotesHelper {
	return &NotesHelper{c: c}
}

// FetchNotesForCurrentTab loads notes for the current tab and renders the list.
// If preserve is true, the current selection is preserved by UUID; otherwise it resets to 0.
func (self *NotesHelper) FetchNotesForCurrentTab(preserve bool) {
	gui := self.c.GuiCommon()
	notesCtx := gui.Contexts().Notes
	prevID := ""
	if preserve {
		prevID = notesCtx.GetSelectedItemId()
	}

	var notes []models.Note
	var err error

	opts := gui.BuildSearchOptions()
	opts.Sort = "created:desc"
	opts.IncludeContent = true
	opts.StripTitle = true
	opts.StripGlobalTags = true

	switch notesCtx.CurrentTab {
	case context.NotesTabAll:
		opts.Limit = 50
		opts.Everything = true
		notes, err = self.c.RuinCmd().Search.Search("", opts)
	case context.NotesTabToday:
		notes, err = self.c.RuinCmd().Search.Search("created:today", opts)
	case context.NotesTabRecent:
		opts.Limit = 20
		recentDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		notes, err = self.c.RuinCmd().Search.Search("after:"+recentDate, opts)
	}

	if err == nil {
		notesCtx.Items = notes
		if preserve && prevID != "" {
			if newIdx := notesCtx.GetList().FindIndexById(prevID); newIdx >= 0 {
				notesCtx.SetSelectedLineIdx(newIdx)
			} else {
				notesCtx.SetSelectedLineIdx(0)
			}
		} else {
			notesCtx.SetSelectedLineIdx(0)
		}
		notesCtx.ClampSelection()
	}
	gui.RenderNotes()
	gui.UpdateNotesTab()
}

// LoadNotesForCurrentTab loads notes based on the current tab,
// resets selection, renders the list, and updates the preview.
func (self *NotesHelper) LoadNotesForCurrentTab() {
	self.FetchNotesForCurrentTab(false)
	self.c.Helpers().Preview().UpdatePreviewForNotes()
}

// CycleNotesTab cycles through All -> Today -> Recent tabs.
func (self *NotesHelper) CycleNotesTab() {
	notesCtx := self.c.GuiCommon().Contexts().Notes
	idx := (notesCtx.TabIndex() + 1) % len(context.NotesTabs)
	notesCtx.CurrentTab = context.NotesTabs[idx]
	notesCtx.SetSelectedLineIdx(0)
	self.LoadNotesForCurrentTab()
}

// SwitchNotesTabByIndex switches to a specific tab by index (for tab click).
func (self *NotesHelper) SwitchNotesTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(context.NotesTabs) {
		return nil
	}
	gui := self.c.GuiCommon()
	notesCtx := gui.Contexts().Notes
	notesCtx.CurrentTab = context.NotesTabs[tabIndex]
	notesCtx.SetSelectedLineIdx(0)
	self.LoadNotesForCurrentTab()
	gui.PushContextByKey("notes")
	return nil
}

// DeleteNote shows a confirmation dialog and deletes the note.
func (self *NotesHelper) DeleteNote(note *models.Note) error {
	if note == nil {
		return nil
	}
	displayName := note.Title
	if displayName == "" {
		displayName = note.Path
	}
	self.c.Helpers().Confirmation().ConfirmDelete("Note", displayName,
		func() error { return self.c.RuinCmd().Note.Delete(note.UUID) },
		func() { self.c.Helpers().Notes().FetchNotesForCurrentTab(false) },
	)
	return nil
}
