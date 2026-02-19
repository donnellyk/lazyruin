package context

import "kvnd/lazyruin/pkg/gui/types"

// ListAdapter implements types.IList by composing data-source functions
// and a ListContextTrait for cursor state. This replaces the per-context
// adapter structs (notesListAdapter, tagsListAdapter, etc.).
type ListAdapter struct {
	lenFn        func() int
	selectedIdFn func() string
	findByIdFn   func(string) int
	cursor       func() *ListContextTrait
}

// NewListAdapter creates a ListAdapter.
func NewListAdapter(
	lenFn func() int,
	selectedIdFn func() string,
	findByIdFn func(string) int,
	cursor func() *ListContextTrait,
) *ListAdapter {
	return &ListAdapter{
		lenFn:        lenFn,
		selectedIdFn: selectedIdFn,
		findByIdFn:   findByIdFn,
		cursor:       cursor,
	}
}

func (a *ListAdapter) Len() int                    { return a.lenFn() }
func (a *ListAdapter) GetSelectedItemId() string   { return a.selectedIdFn() }
func (a *ListAdapter) FindIndexById(id string) int { return a.findByIdFn(id) }
func (a *ListAdapter) GetSelectedLineIdx() int     { return a.cursor().GetSelectedLineIdx() }
func (a *ListAdapter) SetSelectedLineIdx(idx int)  { a.cursor().SetSelectedLineIdx(idx) }
func (a *ListAdapter) MoveSelectedLine(delta int)  { a.cursor().MoveSelectedLine(delta) }
func (a *ListAdapter) ClampSelection()             { a.cursor().ClampSelection() }

var _ types.IList = &ListAdapter{}
