package context

// ListCursor implements types.IListCursor for managing list selection state.
type ListCursor struct {
	selectedLineIdx int
	rangeStartIdx   int
	list            ICursorList
}

// ICursorList provides the length needed for clamping.
type ICursorList interface {
	Len() int
}

// NewListCursor creates a ListCursor bound to a list for clamping.
func NewListCursor(list ICursorList) *ListCursor {
	return &ListCursor{list: list}
}

func (self *ListCursor) GetSelectedLineIdx() int {
	return self.selectedLineIdx
}

func (self *ListCursor) SetSelectedLineIdx(idx int) {
	self.selectedLineIdx = idx
}

func (self *ListCursor) MoveSelectedLine(delta int) {
	next := self.selectedLineIdx + delta
	if next < 0 {
		next = 0
	}
	length := self.list.Len()
	if length > 0 && next >= length {
		next = length - 1
	}
	self.selectedLineIdx = next
}

func (self *ListCursor) ClampSelection() {
	length := self.list.Len()
	if length == 0 {
		self.selectedLineIdx = 0
		return
	}
	if self.selectedLineIdx >= length {
		self.selectedLineIdx = length - 1
	}
	if self.selectedLineIdx < 0 {
		self.selectedLineIdx = 0
	}
}
