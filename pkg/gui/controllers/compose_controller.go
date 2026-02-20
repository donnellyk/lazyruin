package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// ComposeController handles keybindings for the compose preview mode.
// Nav-only — no mutations on synthetic composed notes.
type ComposeController struct {
	baseController
	PreviewNavTrait
	c          *ControllerCommon
	getContext func() *context.ComposeContext
}

var _ types.IController = &ComposeController{}

func NewComposeController(c *ControllerCommon, getContext func() *context.ComposeContext) *ComposeController {
	return &ComposeController{
		PreviewNavTrait: PreviewNavTrait{c: c},
		c:               c,
		getContext:      getContext,
	}
}

func (self *ComposeController) Context() types.Context { return self.getContext() }

func (self *ComposeController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return self.NavBindings()
}

func (self *ComposeController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
