package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
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
	return self.BuildPreviewBindings("compose",
		&types.Binding{
			ID: "compose.open_editor", Key: 'E',
			Handler: self.nav().OpenInEditor, Description: "Open in Editor", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Editor",
		},
		&types.Binding{
			ID: "compose.new_child", Key: gocui.KeyCtrlN,
			Handler: self.newChildNote, Description: "New child note", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "New Child",
		},
	)
}

func (self *ComposeController) newChildNote() error {
	target := self.c.Helpers().PreviewLineOps().ResolveTarget()
	if target == nil {
		return nil
	}
	// Fetch the note title for the capture footer display.
	note, err := self.c.RuinCmd().Search.Get(target.UUID, commands.SearchOptions{})
	if err != nil {
		return nil
	}
	return self.c.Helpers().Capture().OpenCaptureWithParent(target.UUID, note.Title)
}

func (self *ComposeController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
