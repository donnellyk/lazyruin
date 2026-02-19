package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// QueriesController handles all Queries panel keybindings and behavior.
// The panel has two tabs — Queries and Parents — with separate navigation cursors.
type QueriesController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.QueriesContext
}

var _ types.IController = &QueriesController{}

// QueriesControllerOpts holds dependencies for construction.
type QueriesControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.QueriesContext
}

// NewQueriesController creates a new QueriesController.
func NewQueriesController(opts QueriesControllerOpts) *QueriesController {
	return &QueriesController{
		c:          opts.Common,
		getContext: opts.GetContext,
	}
}

// Context returns the context this controller is attached to.
func (self *QueriesController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns the keybinding producer for queries.
func (self *QueriesController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			// Actions
			{
				ID:                "queries.run",
				Key:               gocui.KeyEnter,
				Handler:           self.runActiveItem,
				GetDisabledReason: self.activeItemSelected,
				Description:       "Run Query / View Parent",
				Category:          "Queries",
				DisplayOnScreen:   true,
			},
			{
				ID:                "queries.delete",
				Key:               'd',
				Handler:           self.deleteActiveItem,
				GetDisabledReason: self.activeItemSelected,
				Description:       "Delete Query / Parent",
				Category:          "Queries",
			},
			// Navigation (no Description → excluded from palette)
			{Key: 'j', Handler: self.nextItem},
			{Key: 'k', Handler: self.prevItem},
			{Key: gocui.KeyArrowDown, Handler: self.nextItem},
			{Key: gocui.KeyArrowUp, Handler: self.prevItem},
		}
	}
}

// GetMouseKeybindingsFn returns mouse bindings for the queries panel.
func (self *QueriesController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
	return ListMouseBindings(ListMouseOpts{
		ViewName:    "queries",
		ClickMargin: 2,
		ItemCount:   func() int { return self.getContext().ActiveItemCount() },
		SetSelection: func(idx int) {
			self.getContext().ActiveTrait().SetSelectedLineIdx(idx)
		},
		GetContext: func() types.Context { return self.getContext() },
		GuiCommon:  func() IGuiCommon { return self.c.GuiCommon() },
	})
}

// Navigation — implemented directly (dual-cursor, no ListControllerTrait).

func (self *QueriesController) nextItem() error {
	ctx := self.getContext()
	t := ctx.ActiveTrait()
	count := ctx.ActiveItemCount()
	if t.GetSelectedLineIdx()+1 < count {
		t.MoveSelectedLine(1)
		t.HandleLineChange()
	}
	return nil
}

func (self *QueriesController) prevItem() error {
	ctx := self.getContext()
	t := ctx.ActiveTrait()
	if t.GetSelectedLineIdx() > 0 {
		t.MoveSelectedLine(-1)
		t.HandleLineChange()
	}
	return nil
}

// Action handlers — call helpers directly.

func (self *QueriesController) activeItemSelected() *types.DisabledReason {
	ctx := self.getContext()
	if ctx.ActiveItemCount() == 0 {
		return &types.DisabledReason{Text: "No items"}
	}
	return nil
}

func (self *QueriesController) runActiveItem() error {
	return self.c.H().Queries().RunQuery()
}

func (self *QueriesController) deleteActiveItem() error {
	return self.c.H().Queries().DeleteQuery()
}
