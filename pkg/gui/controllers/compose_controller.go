package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// ComposeController handles keybindings for the compose preview mode.
// Includes navigation + line-level operations (todo, done, inline tag/date).
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
	bindings := self.NavBindings()
	bindings = append(bindings, self.LineOpsBindings("compose")...)
	return bindings
}

func (self *ComposeController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
