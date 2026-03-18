package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// PickResultsController handles keybindings for the pick-results preview mode
// and the pick dialog overlay. In dialog mode, Enter opens the selected result
// and Esc closes the dialog; in normal mode, full NavBindings apply.
type PickResultsController struct {
	baseController
	PreviewNavTrait
	c          *ControllerCommon
	getContext func() *context.PickResultsContext
	dialogMode bool
}

var _ types.IController = &PickResultsController{}

func NewPickResultsController(c *ControllerCommon, getContext func() *context.PickResultsContext) *PickResultsController {
	return &PickResultsController{
		PreviewNavTrait: PreviewNavTrait{c: c},
		c:               c,
		getContext:      getContext,
	}
}

// NewPickDialogController creates a PickResultsController in dialog mode.
func NewPickDialogController(c *ControllerCommon, getContext func() *context.PickResultsContext) *PickResultsController {
	return &PickResultsController{
		PreviewNavTrait: PreviewNavTrait{c: c},
		c:               c,
		getContext:      getContext,
		dialogMode:      true,
	}
}

func (self *PickResultsController) Context() types.Context { return self.getContext() }

func (self *PickResultsController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	if self.dialogMode {
		nav := self.c.Helpers().PreviewNav()
		bindings := []*types.Binding{
			{Key: 'j', Handler: nav.MoveDown},
			{Key: gocui.KeyArrowDown, Handler: nav.MoveDown},
			{Key: 'k', Handler: nav.MoveUp},
			{Key: gocui.KeyArrowUp, Handler: nav.MoveUp},
			{Key: 'J', Handler: nav.CardDown},
			{Key: 'K', Handler: nav.CardUp},
			{Key: gocui.KeyEnter, Handler: self.dialogEnter},
			{Key: gocui.KeyEsc, Handler: self.dialogClose},
		}
		bindings = append(bindings, self.LineOpsBindings("pickDialog")...)
		return bindings
	}
	bindings := self.NavBindings()
	bindings = append(bindings,
		&types.Binding{
			ID: "pickResults.filter", Key: gocui.KeyCtrlF,
			Handler: self.openFilter, Description: "Filter Results", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Filter",
		},
		&types.Binding{
			ID: "pickResults.clear_filter", Key: 'X',
			Handler:           self.clearFilter,
			GetDisabledReason: self.filterNotActive,
			Description:       "Clear Filter", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Clear",
		},
	)
	bindings = append(bindings, self.LineOpsBindings("pickResults")...)
	return bindings
}

func (self *PickResultsController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	if self.dialogMode {
		return nil
	}
	return self.NavMouseBindings()
}

func (self *PickResultsController) openFilter() error {
	return self.c.Helpers().CardListFilter().OpenFilterDialog()
}

func (self *PickResultsController) clearFilter() error {
	return self.c.Helpers().CardListFilter().ClearFilter()
}

func (self *PickResultsController) filterNotActive() *types.DisabledReason {
	if !self.c.Helpers().CardListFilter().FilterActive() {
		return &types.DisabledReason{Text: "No active filter"}
	}
	return nil
}

func (self *PickResultsController) dialogEnter() error {
	return self.c.Helpers().PreviewNav().OpenPickDialogResult()
}

func (self *PickResultsController) dialogClose() error {
	self.c.Helpers().Pick().ClosePickDialog()
	return nil
}
