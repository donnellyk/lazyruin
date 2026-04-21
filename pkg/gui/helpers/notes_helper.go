package helpers

import (
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/models"
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

	err := RefreshList(
		func() ([]models.Note, error) {
			return self.loadNotesForTab(notesCtx.CurrentTab)
		},
		func(notes []models.Note) {
			notesCtx.Items = notes
			self.c.Helpers().TitleCache().PutNotes(notes)
			self.c.Helpers().TitleCache().ResolveUnknownParents(notes)
		},
		notesCtx.GetList(),
		preserve,
	)
	if err != nil {
		gui.ShowError(err)
	}
	gui.RenderNotes()
	gui.UpdateNotesTab()
}

// loadNotesForTab fetches notes for the given tab.
func (self *NotesHelper) loadNotesForTab(tab context.NotesTab) ([]models.Note, error) {
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	opts.Sort = "created:desc"
	opts.IncludeContent = true
	opts.StripTitle = true
	opts.StripGlobalTags = true

	switch tab {
	case context.NotesTabAll:
		opts.Limit = 50
		opts.Everything = true
		return self.c.RuinCmd().Search.Search("", opts)
	case context.NotesTabToday:
		return self.c.RuinCmd().Search.Search("created:today", opts)
	case context.NotesTabRecent:
		opts.Limit = 20
		recentDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		return self.c.RuinCmd().Search.Search("after:"+recentDate, opts)
	case context.NotesTabLinks:
		opts.Limit = 50
		return self.c.RuinCmd().Search.Search("#link", opts)
	default:
		return nil, nil
	}
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
	CycleTab(context.NotesTabs, notesCtx.TabIndex(), func(tab context.NotesTab) {
		notesCtx.CurrentTab = tab
		notesCtx.SetSelectedLineIdx(0)
	}, self.LoadNotesForCurrentTab)
}

// SwitchNotesTabByIndex switches to a specific tab by index (for tab click).
func (self *NotesHelper) SwitchNotesTabByIndex(tabIndex int) error {
	gui := self.c.GuiCommon()
	notesCtx := gui.Contexts().Notes
	SwitchTab(context.NotesTabs, tabIndex, func(tab context.NotesTab) {
		notesCtx.CurrentTab = tab
		notesCtx.SetSelectedLineIdx(0)
	}, func() {
		self.LoadNotesForCurrentTab()
		gui.PushContextByKey("notes")
	})
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
	uuid := note.UUID
	self.c.Helpers().Confirmation().ConfirmDelete("Note", displayName,
		func() error { return self.c.RuinCmd().Note.Delete(uuid) },
		func() {
			self.c.Helpers().Navigator().NoteDeleted(uuid)
			self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
		},
	)
	return nil
}
