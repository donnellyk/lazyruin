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
