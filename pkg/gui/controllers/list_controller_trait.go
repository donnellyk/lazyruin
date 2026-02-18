package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// ListControllerTrait provides generic list navigation behavior for controllers.
// T is the type of the list items (e.g., *models.Note, *models.Tag).
// It is embedded into concrete list controllers.
type ListControllerTrait[T any] struct {
	c          *ControllerCommon
	getContext func() types.IListContext
	getItems   func() []T
	trait      func() *context.ListContextTrait
}

// NewListControllerTrait creates a new ListControllerTrait.
func NewListControllerTrait[T any](
	c *ControllerCommon,
	getContext func() types.IListContext,
	getItems func() []T,
	trait func() *context.ListContextTrait,
) *ListControllerTrait[T] {
	return &ListControllerTrait[T]{
		c:          c,
		getContext: getContext,
		getItems:   getItems,
		trait:      trait,
	}
}

// withItem wraps a handler that requires a selected item.
// If no item is selected, the handler is skipped.
func (self *ListControllerTrait[T]) withItem(fn func(item T) error) func() error {
	return func() error {
		items := self.getItems()
		idx := self.trait().GetSelectedLineIdx()
		if idx < 0 || idx >= len(items) {
			return nil
		}
		return fn(items[idx])
	}
}

// singleItemSelected returns a DisabledReason producer that checks
// whether the list has at least one item selected.
func (self *ListControllerTrait[T]) singleItemSelected() func() *types.DisabledReason {
	return func() *types.DisabledReason {
		items := self.getItems()
		if len(items) == 0 {
			return &types.DisabledReason{Text: "No items"}
		}
		idx := self.trait().GetSelectedLineIdx()
		if idx < 0 || idx >= len(items) {
			return &types.DisabledReason{Text: "No item selected"}
		}
		return nil
	}
}

// require combines multiple disabled-reason producers. The first non-nil reason wins.
func (self *ListControllerTrait[T]) require(fns ...func() *types.DisabledReason) func() *types.DisabledReason {
	return func() *types.DisabledReason {
		for _, fn := range fns {
			if reason := fn(); reason != nil {
				return reason
			}
		}
		return nil
	}
}

// Navigation handlers — ported from existing listPanel listDown/listUp/listTop/listBottom.

func (self *ListControllerTrait[T]) nextItem() error {
	t := self.trait()
	items := self.getItems()
	idx := t.GetSelectedLineIdx()
	if idx+1 < len(items) {
		t.MoveSelectedLine(1)
		t.HandleLineChange()
	}
	return nil
}

func (self *ListControllerTrait[T]) prevItem() error {
	t := self.trait()
	idx := t.GetSelectedLineIdx()
	if idx > 0 {
		t.MoveSelectedLine(-1)
		t.HandleLineChange()
	}
	return nil
}

func (self *ListControllerTrait[T]) goTop() error {
	t := self.trait()
	t.SetSelectedLineIdx(0)
	t.HandleLineChange()
	return nil
}

func (self *ListControllerTrait[T]) goBottom() error {
	t := self.trait()
	items := self.getItems()
	if len(items) > 0 {
		t.SetSelectedLineIdx(len(items) - 1)
		t.HandleLineChange()
	}
	return nil
}

// NavBindings returns the standard j/k/g/G/arrow navigation bindings.
// These have no Description so they're excluded from the palette.
func (self *ListControllerTrait[T]) NavBindings() []*types.Binding {
	return []*types.Binding{
		{Key: 'j', Handler: self.nextItem},
		{Key: 'k', Handler: self.prevItem},
		{Key: 'g', Handler: self.goTop},
		{Key: 'G', Handler: self.goBottom},
		{Key: gocui.KeyArrowDown, Handler: self.nextItem},
		{Key: gocui.KeyArrowUp, Handler: self.prevItem},
	}
}
