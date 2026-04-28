package context

import "testing"

func TestNotesHomeContext_NavigationSkipsHeadersAndBlanks(t *testing.T) {
	c := NewNotesHomeContext()
	c.Rows = []NotesHomeRow{
		{Title: "Inbox", ItemID: "hardcoded:inbox"},
		{Blank: true},
		{Title: "Today", ItemID: "hardcoded:today"},
		{Title: "Next 7 Days", ItemID: "hardcoded:next7"},
		{Blank: true},
		{IsHeader: true, Title: "Pinned"},
		{Title: "Parent A", ItemID: "parent:u1"},
		{Title: "Parent B", ItemID: "parent:u2"},
	}
	c.SelectedIdx = 0

	// Forward
	c.SelectedIdx = c.NextSelectable(c.SelectedIdx)
	if c.SelectedIdx != 2 {
		t.Errorf("after first next: got idx %d, want 2 (Today)", c.SelectedIdx)
	}
	c.SelectedIdx = c.NextSelectable(c.SelectedIdx)
	if c.SelectedIdx != 3 {
		t.Errorf("after second next: got idx %d, want 3 (Next 7 Days)", c.SelectedIdx)
	}
	c.SelectedIdx = c.NextSelectable(c.SelectedIdx)
	if c.SelectedIdx != 6 {
		t.Errorf("after third next: got idx %d, want 6 (Parent A — header skipped)", c.SelectedIdx)
	}

	// Backward
	c.SelectedIdx = c.PrevSelectable(c.SelectedIdx)
	if c.SelectedIdx != 3 {
		t.Errorf("after prev: got idx %d, want 3", c.SelectedIdx)
	}

	// At end / start, return self.
	last := c.NextSelectable(7)
	if last != 7 {
		t.Errorf("next from last: got %d, want 7 (no movement)", last)
	}
	first := c.PrevSelectable(0)
	if first != 0 {
		t.Errorf("prev from first: got %d, want 0 (no movement)", first)
	}
}

func TestNotesHomeContext_SelectedReturnsNilOnHeader(t *testing.T) {
	c := NewNotesHomeContext()
	c.Rows = []NotesHomeRow{
		{IsHeader: true, Title: "Pinned"},
		{Title: "Item", ItemID: "x"},
	}
	c.SelectedIdx = 0
	if c.Selected() != nil {
		t.Error("expected nil when SelectedIdx points at a header")
	}
	c.SelectedIdx = 1
	if c.Selected() == nil {
		t.Error("expected non-nil when SelectedIdx points at an item")
	}
}

func TestNotesHomeContext_SetRowsPreservingSelection(t *testing.T) {
	c := NewNotesHomeContext()
	c.Rows = []NotesHomeRow{
		{Title: "Inbox", ItemID: "hardcoded:inbox"},
		{Title: "Today", ItemID: "hardcoded:today"},
	}
	c.SelectedIdx = 1

	// Replace with new rows containing the previous selection's ID.
	newRows := []NotesHomeRow{
		{IsHeader: true, Title: "Group"},
		{Title: "Other", ItemID: "other"},
		{Title: "Today", ItemID: "hardcoded:today"},
	}
	c.SetRowsPreservingSelection(newRows, "hardcoded:today")
	if c.SelectedIdx != 2 {
		t.Errorf("expected cursor at restored item idx 2, got %d", c.SelectedIdx)
	}

	// Replace with rows that don't contain the preserved ID — should land
	// on first selectable.
	c.SetRowsPreservingSelection(newRows, "missing")
	if c.SelectedIdx != 1 {
		t.Errorf("expected first selectable idx 1, got %d", c.SelectedIdx)
	}
}

func TestNotesHomeContext_FirstSelectableEmpty(t *testing.T) {
	c := NewNotesHomeContext()
	c.Rows = []NotesHomeRow{
		{IsHeader: true, Title: "Header only"},
		{Blank: true},
	}
	if got := c.FirstSelectableIdx(); got != -1 {
		t.Errorf("FirstSelectableIdx with no items = %d, want -1", got)
	}
	if c.Selected() != nil {
		t.Error("expected Selected() nil on header-only list")
	}
}

func TestNotesHomeContext_SelectedItemID(t *testing.T) {
	c := NewNotesHomeContext()
	c.Rows = []NotesHomeRow{
		{Title: "X", ItemID: "x"},
	}
	if c.SelectedItemID() != "x" {
		t.Errorf("SelectedItemID = %q, want x", c.SelectedItemID())
	}
}
