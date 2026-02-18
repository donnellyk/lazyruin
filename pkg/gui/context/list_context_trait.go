package context

// ListContextTrait provides common list state and behavior for list contexts.
// It manages selection state and delegates rendering/preview updates.
type ListContextTrait struct {
	selectedLineIdx int
	onLineChange    func()
}

// NewListContextTrait creates a new ListContextTrait.
func NewListContextTrait(onLineChange func()) *ListContextTrait {
	return &ListContextTrait{
		onLineChange: onLineChange,
	}
}

// GetSelectedLineIdx returns the current selection index.
func (self *ListContextTrait) GetSelectedLineIdx() int {
	return self.selectedLineIdx
}

// SetSelectedLineIdx sets the current selection index.
func (self *ListContextTrait) SetSelectedLineIdx(idx int) {
	self.selectedLineIdx = idx
}

// MoveSelectedLine moves the selection by delta.
func (self *ListContextTrait) MoveSelectedLine(delta int) {
	self.selectedLineIdx += delta
}

// ClampSelection clamps the selection to valid bounds.
// Requires the caller to pass the item count.
func (self *ListContextTrait) ClampSelection(itemCount int) {
	if self.selectedLineIdx < 0 {
		self.selectedLineIdx = 0
	}
	if itemCount > 0 && self.selectedLineIdx >= itemCount {
		self.selectedLineIdx = itemCount - 1
	}
}

// HandleLineChange triggers rendering and preview updates after selection changes.
func (self *ListContextTrait) HandleLineChange() {
	if self.onLineChange != nil {
		self.onLineChange()
	}
}
