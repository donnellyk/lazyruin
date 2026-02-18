package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// NotesTab represents the sub-tabs within the Notes panel.
type NotesTab string

const (
	NotesTabAll    NotesTab = "all"
	NotesTabToday  NotesTab = "today"
	NotesTabRecent NotesTab = "recent"
)

// NotesTabs maps tab indices to NotesTab values.
var NotesTabs = []NotesTab{NotesTabAll, NotesTabToday, NotesTabRecent}

// NotesContext owns all Notes panel state: items, cursor, and tab.
type NotesContext struct {
	BaseContext
	*ListContextTrait

	Items      []models.Note
	CurrentTab NotesTab
	list       *notesList
}

// notesList adapts NotesContext for the IList and IListCursor interfaces.
type notesList struct {
	ctx *NotesContext
}

func (l *notesList) Len() int {
	return len(l.ctx.Items)
}

func (l *notesList) GetSelectedItemId() string {
	item := l.ctx.Selected()
	if item == nil {
		return ""
	}
	return item.UUID
}

func (l *notesList) FindIndexById(id string) int {
	for i, note := range l.ctx.Items {
		if note.UUID == id {
			return i
		}
	}
	return -1
}

// NewNotesContext creates a NotesContext.
// renderFn and previewFn are called after selection changes.
func NewNotesContext(renderFn func(), previewFn func()) *NotesContext {
	ctx := &NotesContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.SIDE_CONTEXT,
			Key:       "notes",
			ViewName:  "notes",
			Focusable: true,
			Title:     "Notes",
		}),
		CurrentTab: NotesTabAll,
	}
	ctx.list = &notesList{ctx: ctx}
	cursor := NewListCursor(ctx.list)
	ctx.ListContextTrait = NewListContextTrait(cursor, renderFn, previewFn)
	return ctx
}

// Selected returns the currently selected note, or nil.
func (self *NotesContext) Selected() *models.Note {
	if len(self.Items) == 0 {
		return nil
	}
	idx := self.GetSelectedLineIdx()
	if idx >= len(self.Items) {
		idx = 0
	}
	return &self.Items[idx]
}

// TabIndex returns the current tab index.
func (self *NotesContext) TabIndex() int {
	for i, tab := range NotesTabs {
		if tab == self.CurrentTab {
			return i
		}
	}
	return 0
}

// GetList returns the IList adapter for this context.
func (self *NotesContext) GetList() types.IList {
	return &notesListAdapter{ctx: self}
}

// GetSelectedItemId returns the stable ID of the selected item.
func (self *NotesContext) GetSelectedItemId() string {
	return self.list.GetSelectedItemId()
}

// notesListAdapter wraps NotesContext to implement types.IList.
type notesListAdapter struct {
	ctx *NotesContext
}

func (a *notesListAdapter) Len() int                     { return a.ctx.list.Len() }
func (a *notesListAdapter) GetSelectedItemId() string    { return a.ctx.list.GetSelectedItemId() }
func (a *notesListAdapter) FindIndexById(id string) int  { return a.ctx.list.FindIndexById(id) }
func (a *notesListAdapter) GetSelectedLineIdx() int      { return a.ctx.GetSelectedLineIdx() }
func (a *notesListAdapter) SetSelectedLineIdx(idx int)   { a.ctx.SetSelectedLineIdx(idx) }
func (a *notesListAdapter) MoveSelectedLine(delta int)   { a.ctx.MoveSelectedLine(delta) }
func (a *notesListAdapter) ClampSelection()              { a.ctx.ClampSelection() }

// Verify interface compliance at compile time.
var _ types.IListContext = &NotesContext{}
