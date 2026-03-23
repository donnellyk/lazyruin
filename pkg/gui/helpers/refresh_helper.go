package helpers

import "kvnd/lazyruin/pkg/gui/types"

// RefreshHelper handles data refreshing with selection preservation.
// It uses stable IDs (GetSelectedItemId + FindIndexById) to preserve
// selection across data refreshes, not raw indices.
type RefreshHelper struct {
	c *HelperCommon
}

// NewRefreshHelper creates a new RefreshHelper.
func NewRefreshHelper(c *HelperCommon) *RefreshHelper {
	return &RefreshHelper{c: c}
}

// PreserveSelection refreshes a list context while preserving the
// selected item by stable ID. If the previously selected item is
// no longer present, selection falls back to index 0.
func (self *RefreshHelper) PreserveSelection(list types.IListContext) {
	prevID := list.GetSelectedItemId()
	l := list.GetList()
	l.ClampSelection()
	if prevID != "" {
		newIdx := l.FindIndexById(prevID)
		if newIdx >= 0 {
			l.SetSelectedLineIdx(newIdx)
		}
	}
}

// RefreshList fetches items, updates a list context, and optionally preserves
// selection by stable ID. It extracts the common refresh-and-preserve pattern
// used by NotesHelper, TagsHelper, and QueriesHelper.
//
// Parameters:
//   - load: fetches the new items from the backend
//   - setItems: sets the fetched items on the context
//   - list: provides selection state and ID-based lookup (typically from GetList())
//   - preserve: if true, the previously selected item is restored by ID
func RefreshList[T any](
	load func() ([]T, error),
	setItems func([]T),
	list types.IList,
	preserve bool,
) error {
	prevID := ""
	if preserve {
		prevID = list.GetSelectedItemId()
	}

	items, err := load()
	if err != nil {
		return err
	}

	setItems(items)

	if preserve && prevID != "" {
		if newIdx := list.FindIndexById(prevID); newIdx >= 0 {
			list.SetSelectedLineIdx(newIdx)
		}
	} else {
		list.SetSelectedLineIdx(0)
	}
	list.ClampSelection()

	return nil
}

// RefreshAll refreshes data for all panels.
func (self *RefreshHelper) RefreshAll() {
	h := self.c.Helpers()
	h.Notes().FetchNotesForCurrentTab(false)
	h.Tags().RefreshTags(false)
	h.Queries().RefreshQueries(false)
	h.Queries().RefreshParents(false)
}

// RenderAll re-renders all panels.
func (self *RefreshHelper) RenderAll() {
	gui := self.c.GuiCommon()
	gui.RenderNotes()
	gui.RenderQueries()
	gui.RenderTags()
	gui.RenderPreview()
	gui.UpdateStatusBar()
}
