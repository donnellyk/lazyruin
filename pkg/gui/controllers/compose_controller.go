package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/helpers"
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

func (self *ComposeController) lineOps() *helpers.PreviewLineOpsHelper {
	return self.c.Helpers().PreviewLineOps()
}

func (self *ComposeController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	bindings := self.NavBindings()
	bindings = append(bindings,
		&types.Binding{
			ID: "compose.toggle_todo", Key: 'x',
			Handler: self.lineOps().ToggleTodo, Description: "Toggle Todo", Category: "Preview",
		},
		&types.Binding{
			ID: "compose.append_done", Key: 'D',
			Handler: self.lineOps().AppendDone, Description: "Toggle #done", Category: "Preview",
		},
		&types.Binding{
			ID: "compose.toggle_inline_tag", Key: gocui.KeyCtrlT,
			Handler: self.lineOps().ToggleInlineTag, Description: "Toggle Inline Tag", Category: "Preview",
		},
		&types.Binding{
			ID: "compose.toggle_inline_date", Key: gocui.KeyCtrlD,
			Handler: self.lineOps().ToggleInlineDate, Description: "Toggle Inline Date", Category: "Preview",
		},
	)
	return bindings
}

func (self *ComposeController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
