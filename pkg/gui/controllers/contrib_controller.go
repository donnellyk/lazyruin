package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

// ContribController handles keybindings for the contribution chart dialog popup.
type ContribController struct {
	baseController
	GridNavTrait
	c          *ControllerCommon
	getContext func() *context.ContribContext
}

var _ types.IController = &ContribController{}

// ContribControllerOpts holds the dependencies for ContribController.
type ContribControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.ContribContext
}

// NewContribController creates a ContribController.
func NewContribController(opts ContribControllerOpts) *ContribController {
	ctrl := &ContribController{
		c:          opts.Common,
		getContext: opts.GetContext,
	}
	ctrl.GridNavTrait = GridNavTrait{
		c:       opts.Common,
		hDelta:  7,
		vDelta:  1,
		moveDay: opts.Common.Helpers().Contrib().MoveDay,
	}
	return ctrl
}

// Context returns the context this controller is attached to.
func (self *ContribController) Context() types.Context {
	return self.getContext()
}

func (self *ContribController) close() error {
	self.c.Helpers().Contrib().Close()
	return nil
}

// GetMouseKeybindings returns mouse bindings for the contribution notes list.
func (self *ContribController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return WheelScrollBindings("contribNotes", func() IGuiCommon { return self.c.GuiCommon() })
}

// GetKeybindings returns keybindings for the contribution chart dialog.
func (self *ContribController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	gv := "contribGrid"
	nv := "contribNotes"
	return []*types.Binding{
		// Grid navigation
		{ViewName: gv, Key: 'h', Handler: self.gridLeft},
		{ViewName: gv, Key: 'l', Handler: self.gridRight},
		{ViewName: gv, Key: 'k', Handler: self.gridUp},
		{ViewName: gv, Key: 'j', Handler: self.gridDown},
		{ViewName: gv, Key: gocui.KeyArrowLeft, Handler: self.gridLeft},
		{ViewName: gv, Key: gocui.KeyArrowRight, Handler: self.gridRight},
		{ViewName: gv, Key: gocui.KeyArrowUp, Handler: self.gridUp},
		{ViewName: gv, Key: gocui.KeyArrowDown, Handler: self.gridDown},
		{ViewName: gv, Key: gocui.KeyEnter, Handler: self.c.Helpers().Contrib().GridEnter},
		{ViewName: gv, Key: gocui.KeyEsc, Handler: self.close},
		{ViewName: gv, Key: gocui.KeyTab, Handler: self.c.Helpers().Contrib().Tab},
		// Note list navigation
		{ViewName: nv, Key: 'j', Handler: self.c.Helpers().Contrib().NoteDown},
		{ViewName: nv, Key: 'k', Handler: self.c.Helpers().Contrib().NoteUp},
		{ViewName: nv, Key: gocui.KeyArrowDown, Handler: self.c.Helpers().Contrib().NoteDown},
		{ViewName: nv, Key: gocui.KeyArrowUp, Handler: self.c.Helpers().Contrib().NoteUp},
		{ViewName: nv, Key: gocui.KeyEnter, Handler: self.c.Helpers().Contrib().NoteEnter},
		{ViewName: nv, Key: gocui.KeyEsc, Handler: self.close},
		{ViewName: nv, Key: gocui.KeyTab, Handler: self.c.Helpers().Contrib().Tab},
	}
}
