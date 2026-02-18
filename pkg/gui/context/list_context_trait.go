package context

import "kvnd/lazyruin/pkg/gui/types"

// ListContextTrait provides shared list behavior for list contexts.
// It composes a ListCursor with render/preview callbacks.
// Ported from the existing listPanel logic.
type ListContextTrait struct {
	cursor    *ListCursor
	renderFn  func()
	previewFn func()
}

// NewListContextTrait creates a new trait bound to a cursor.
func NewListContextTrait(cursor *ListCursor, renderFn func(), previewFn func()) *ListContextTrait {
	return &ListContextTrait{
		cursor:    cursor,
		renderFn:  renderFn,
		previewFn: previewFn,
	}
}

// GetCursor returns the underlying list cursor.
func (self *ListContextTrait) GetCursor() *ListCursor {
	return self.cursor
}

func (self *ListContextTrait) GetSelectedLineIdx() int   { return self.cursor.GetSelectedLineIdx() }
func (self *ListContextTrait) SetSelectedLineIdx(idx int) { self.cursor.SetSelectedLineIdx(idx) }
func (self *ListContextTrait) MoveSelectedLine(delta int) { self.cursor.MoveSelectedLine(delta) }
func (self *ListContextTrait) ClampSelection()            { self.cursor.ClampSelection() }

// HandleLineChange re-renders the list and updates the preview.
// This is the equivalent of the old listPanel's render+updatePreview pattern.
func (self *ListContextTrait) HandleLineChange() {
	if self.renderFn != nil {
		self.renderFn()
	}
	if self.previewFn != nil {
		self.previewFn()
	}
}

// Verify ListContextTrait satisfies IListCursor at compile time.
var _ types.IListCursor = &ListContextTrait{}
