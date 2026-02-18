package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// ContribController handles keybindings for the contribution chart dialog popup.
type ContribController struct {
	baseController
	getContext  func() *context.ContribContext
	onGridLeft  func() error
	onGridRight func() error
	onGridUp    func() error
	onGridDown  func() error
	onGridEnter func() error
	onEsc       func() error // both views
	onTab       func() error // both views
	onNoteDown  func() error
	onNoteUp    func() error
	onNoteEnter func() error
}

var _ types.IController = &ContribController{}

// ContribControllerOpts holds the callbacks injected during wiring.
type ContribControllerOpts struct {
	GetContext  func() *context.ContribContext
	OnGridLeft  func() error
	OnGridRight func() error
	OnGridUp    func() error
	OnGridDown  func() error
	OnGridEnter func() error
	OnEsc       func() error
	OnTab       func() error
	OnNoteDown  func() error
	OnNoteUp    func() error
	OnNoteEnter func() error
}

// NewContribController creates a ContribController.
func NewContribController(opts ContribControllerOpts) *ContribController {
	return &ContribController{
		getContext:  opts.GetContext,
		onGridLeft:  opts.OnGridLeft,
		onGridRight: opts.OnGridRight,
		onGridUp:    opts.OnGridUp,
		onGridDown:  opts.OnGridDown,
		onGridEnter: opts.OnGridEnter,
		onEsc:       opts.OnEsc,
		onTab:       opts.OnTab,
		onNoteDown:  opts.OnNoteDown,
		onNoteUp:    opts.OnNoteUp,
		onNoteEnter: opts.OnNoteEnter,
	}
}

// Context returns the context this controller is attached to.
func (self *ContribController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns keybindings for the contribution chart dialog.
func (self *ContribController) GetKeybindingsFn() types.KeybindingsFn {
	gv := "contribGrid"
	nv := "contribNotes"
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			// Grid navigation
			{ViewName: gv, Key: 'h', Handler: self.onGridLeft},
			{ViewName: gv, Key: 'l', Handler: self.onGridRight},
			{ViewName: gv, Key: 'k', Handler: self.onGridUp},
			{ViewName: gv, Key: 'j', Handler: self.onGridDown},
			{ViewName: gv, Key: gocui.KeyArrowLeft, Handler: self.onGridLeft},
			{ViewName: gv, Key: gocui.KeyArrowRight, Handler: self.onGridRight},
			{ViewName: gv, Key: gocui.KeyArrowUp, Handler: self.onGridUp},
			{ViewName: gv, Key: gocui.KeyArrowDown, Handler: self.onGridDown},
			{ViewName: gv, Key: gocui.KeyEnter, Handler: self.onGridEnter},
			{ViewName: gv, Key: gocui.KeyEsc, Handler: self.onEsc},
			{ViewName: gv, Key: gocui.KeyTab, Handler: self.onTab},
			// Note list navigation
			{ViewName: nv, Key: 'j', Handler: self.onNoteDown},
			{ViewName: nv, Key: 'k', Handler: self.onNoteUp},
			{ViewName: nv, Key: gocui.KeyArrowDown, Handler: self.onNoteDown},
			{ViewName: nv, Key: gocui.KeyArrowUp, Handler: self.onNoteUp},
			{ViewName: nv, Key: gocui.KeyEnter, Handler: self.onNoteEnter},
			{ViewName: nv, Key: gocui.KeyEsc, Handler: self.onEsc},
			{ViewName: nv, Key: gocui.KeyTab, Handler: self.onTab},
		}
	}
}
