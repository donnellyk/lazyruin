package types

// IList defines the interface for a list data source with stable ID-based selection.
type IList interface {
	IListCursor
	Len() int
	// GetSelectedItemId returns a stable ID for the selected item (e.g., UUID).
	// Used by RefreshHelper to preserve selection across data refreshes.
	GetSelectedItemId() string
	// FindIndexById locates an item by stable ID, returning -1 if not found.
	FindIndexById(id string) int
}

// IListCursor manages the selection state for a list.
type IListCursor interface {
	GetSelectedLineIdx() int
	SetSelectedLineIdx(int)
	MoveSelectedLine(delta int)
	ClampSelection()
}
