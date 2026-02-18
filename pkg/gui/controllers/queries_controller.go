package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	helpers "kvnd/lazyruin/pkg/gui/helpers"
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

func (self *QueriesController) h() *helpers.Helpers {
	return self.c.Helpers().(*helpers.Helpers)
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
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "queries",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("queries")
					if v == nil {
						return nil
					}
					ctx := self.getContext()
					idx := helpers.ListClickIndex(v, 2)
					if ctx.CurrentTab == context.QueriesTabParents {
						if idx >= 0 && idx < len(ctx.Parents) {
							ctx.ParentsTrait().SetSelectedLineIdx(idx)
						}
					} else {
						if idx >= 0 && idx < len(ctx.Queries) {
							ctx.QueriesTrait().SetSelectedLineIdx(idx)
						}
					}
					self.c.GuiCommon().PushContext(ctx, types.OnFocusOpts{})
					return nil
				},
			},
			{
				ViewName: "queries",
				Key:      gocui.MouseWheelDown,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("queries")
					if v != nil {
						helpers.ScrollViewport(v, 3)
					}
					return nil
				},
			},
			{
				ViewName: "queries",
				Key:      gocui.MouseWheelUp,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("queries")
					if v != nil {
						helpers.ScrollViewport(v, -3)
					}
					return nil
				},
			},
		}
	}
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
	return self.h().Queries().RunQuery()
}

func (self *QueriesController) deleteActiveItem() error {
	return self.h().Queries().DeleteQuery()
}
