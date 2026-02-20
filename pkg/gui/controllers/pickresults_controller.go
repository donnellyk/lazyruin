package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// PickResultsController handles keybindings for the pick-results preview mode.
// Nav-only — no mutations on pick results.
type PickResultsController struct {
	baseController
	PreviewNavTrait
	c          *ControllerCommon
	getContext func() *context.PickResultsContext
}

var _ types.IController = &PickResultsController{}

func NewPickResultsController(c *ControllerCommon, getContext func() *context.PickResultsContext) *PickResultsController {
	return &PickResultsController{
		PreviewNavTrait: PreviewNavTrait{c: c},
		c:               c,
		getContext:      getContext,
	}
}

func (self *PickResultsController) Context() types.Context { return self.getContext() }

func (self *PickResultsController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return self.NavBindings()
}

func (self *PickResultsController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
