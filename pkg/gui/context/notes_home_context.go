package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

// NotesOuterTab is the active outer tab in the Notes pane when
// sections_mode is enabled. The empty string signals "no outer-tab system
// in play" (sections_mode disabled).
type NotesOuterTab string

const (
	NotesOuterTabHome  NotesOuterTab = "home"
	NotesOuterTabNotes NotesOuterTab = "notes"
)

// NotesHomeActionKind enumerates the dispatch types for an item activation.
type NotesHomeActionKind int

const (
	NotesHomeActionInbox NotesHomeActionKind = iota
	NotesHomeActionToday
	NotesHomeActionNext7
	NotesHomeActionParent
	NotesHomeActionQuery
	NotesHomeActionEmbed
)

// NotesHomeAction tells the helper how to activate an item. Detail holds the
// payload: parent UUID, saved-query name, or full embed string.
type NotesHomeAction struct {
	Kind   NotesHomeActionKind
	Detail string
}

// NotesHomeRow is a single line in the Home tab — either a section header
// (non-selectable) or an activatable item.
type NotesHomeRow struct {
	IsHeader bool
	Blank    bool   // true for purely-blank spacer rows between groups
	Title    string // header label or item label
	ItemID   string // stable selection identifier; empty for headers/blanks
	Action   NotesHomeAction
}

// NotesHomeContext owns the Home tab's section list and cursor state. Lives
// alongside (and shares the "notes" view with) NotesContext when
// SectionsMode is enabled in config.
type NotesHomeContext struct {
	BaseContext

	Rows        []NotesHomeRow
	SelectedIdx int // always points at a selectable (item) row when Rows is non-empty
}

// NewNotesHomeContext creates a NotesHomeContext that shares the "notes"
// view with NotesContext. When this context is current, the Home tab is
// active; when NotesContext is current, the flat list is.
func NewNotesHomeContext() *NotesHomeContext {
	return &NotesHomeContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.SIDE_CONTEXT,
			Key:       "notesHome",
			ViewName:  "notes",
			Focusable: true,
			Title:     "Notes",
		}),
	}
}

// Selected returns the currently selected row, or nil if the list is empty
// or the selection points at a header (which should not happen in normal
// operation but is checked defensively).
func (self *NotesHomeContext) Selected() *NotesHomeRow {
	if len(self.Rows) == 0 {
		return nil
	}
	if self.SelectedIdx < 0 || self.SelectedIdx >= len(self.Rows) {
		return nil
	}
	row := self.Rows[self.SelectedIdx]
	if row.IsHeader || row.Blank {
		return nil
	}
	return &row
}

// FirstSelectableIdx returns the index of the first non-header, non-blank row,
// or -1 if no such row exists.
func (self *NotesHomeContext) FirstSelectableIdx() int {
	for i, r := range self.Rows {
		if !r.IsHeader && !r.Blank {
			return i
		}
	}
	return -1
}

// NextSelectable returns the next selectable index after `from`, or `from`
// itself if there is no next item.
func (self *NotesHomeContext) NextSelectable(from int) int {
	for i := from + 1; i < len(self.Rows); i++ {
		if !self.Rows[i].IsHeader && !self.Rows[i].Blank {
			return i
		}
	}
	return from
}

// PrevSelectable returns the previous selectable index before `from`, or
// `from` itself if there is no previous item.
func (self *NotesHomeContext) PrevSelectable(from int) int {
	for i := from - 1; i >= 0; i-- {
		if !self.Rows[i].IsHeader && !self.Rows[i].Blank {
			return i
		}
	}
	return from
}

// SetRowsPreservingSelection replaces Rows. If preserveID matches an existing
// item's ItemID, the cursor moves to that item; otherwise it lands on the
// first selectable row.
func (self *NotesHomeContext) SetRowsPreservingSelection(rows []NotesHomeRow, preserveID string) {
	self.Rows = rows
	if preserveID != "" {
		for i, r := range rows {
			if !r.IsHeader && !r.Blank && r.ItemID == preserveID {
				self.SelectedIdx = i
				return
			}
		}
	}
	self.SelectedIdx = self.FirstSelectableIdx()
	if self.SelectedIdx < 0 {
		self.SelectedIdx = 0
	}
}

// SelectedItemID returns the stable ID of the currently selected row, or
// empty string if no selectable row is highlighted.
func (self *NotesHomeContext) SelectedItemID() string {
	r := self.Selected()
	if r == nil {
		return ""
	}
	return r.ItemID
}

var _ types.Context = &NotesHomeContext{}
