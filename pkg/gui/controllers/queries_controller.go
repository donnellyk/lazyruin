package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// QueriesController handles all Queries panel keybindings and behavior.
// The panel has two tabs — Queries and Parents — with separate navigation cursors.
type QueriesController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.QueriesContext

	// Queries tab callbacks
	onRunQuery    func(query *models.Query) error
	onDeleteQuery func(query *models.Query) error

	// Parents tab callbacks
	onViewParent   func(parent *models.ParentBookmark) error
	onDeleteParent func(parent *models.ParentBookmark) error

	// Mouse callbacks
	onClickFn   func(g *gocui.Gui, v *gocui.View) error
	onWheelDown func(g *gocui.Gui, v *gocui.View) error
	onWheelUp   func(g *gocui.Gui, v *gocui.View) error
}

var _ types.IController = &QueriesController{}

// QueriesControllerOpts holds the callbacks injected during wiring.
type QueriesControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.QueriesContext
	// Queries tab callbacks
	OnRunQuery    func(query *models.Query) error
	OnDeleteQuery func(query *models.Query) error
	// Parents tab callbacks
	OnViewParent   func(parent *models.ParentBookmark) error
	OnDeleteParent func(parent *models.ParentBookmark) error
	// Mouse callbacks
	OnClick     func(g *gocui.Gui, v *gocui.View) error
	OnWheelDown func(g *gocui.Gui, v *gocui.View) error
	OnWheelUp   func(g *gocui.Gui, v *gocui.View) error
}

// NewQueriesController creates a new QueriesController.
func NewQueriesController(opts QueriesControllerOpts) *QueriesController {
	return &QueriesController{
		c:              opts.Common,
		getContext:     opts.GetContext,
		onRunQuery:     opts.OnRunQuery,
		onDeleteQuery:  opts.OnDeleteQuery,
		onViewParent:   opts.OnViewParent,
		onDeleteParent: opts.OnDeleteParent,
		onClickFn:      opts.OnClick,
		onWheelDown:    opts.OnWheelDown,
		onWheelUp:      opts.OnWheelUp,
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
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "queries",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onClickFn(nil, nil)
				},
			},
			{
				ViewName: "queries",
				Key:      gocui.MouseWheelDown,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onWheelDown(nil, nil)
				},
			},
			{
				ViewName: "queries",
				Key:      gocui.MouseWheelUp,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.onWheelUp(nil, nil)
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

// Action handlers — dispatch based on current tab.

func (self *QueriesController) activeItemSelected() *types.DisabledReason {
	ctx := self.getContext()
	if ctx.ActiveItemCount() == 0 {
		return &types.DisabledReason{Text: "No items"}
	}
	return nil
}

func (self *QueriesController) runActiveItem() error {
	ctx := self.getContext()
	if ctx.CurrentTab == context.QueriesTabParents {
		p := ctx.SelectedParent()
		if p == nil {
			return nil
		}
		return self.onViewParent(p)
	}
	q := ctx.SelectedQuery()
	if q == nil {
		return nil
	}
	return self.onRunQuery(q)
}

func (self *QueriesController) deleteActiveItem() error {
	ctx := self.getContext()
	if ctx.CurrentTab == context.QueriesTabParents {
		p := ctx.SelectedParent()
		if p == nil {
			return nil
		}
		return self.onDeleteParent(p)
	}
	q := ctx.SelectedQuery()
	if q == nil {
		return nil
	}
	return self.onDeleteQuery(q)
}
