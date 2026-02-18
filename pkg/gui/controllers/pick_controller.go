package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// PickController handles all pick popup keybindings.
type PickController struct {
	baseController
	getContext  func() *context.PickContext
	onEnter     func() error
	onEsc       func() error
	onTab       func() error
	onToggleAny func() error
}

var _ types.IController = &PickController{}

// PickControllerOpts holds the callbacks injected during wiring.
type PickControllerOpts struct {
	GetContext  func() *context.PickContext
	OnEnter     func() error
	OnEsc       func() error
	OnTab       func() error
	OnToggleAny func() error
}

// NewPickController creates a PickController.
func NewPickController(opts PickControllerOpts) *PickController {
	return &PickController{
		getContext:  opts.GetContext,
		onEnter:     opts.OnEnter,
		onEsc:       opts.OnEsc,
		onTab:       opts.OnTab,
		onToggleAny: opts.OnToggleAny,
	}
}

// Context returns the context this controller is attached to.
func (self *PickController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns the keybinding producer for the pick popup.
func (self *PickController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			{Key: gocui.KeyEnter, Handler: self.onEnter},
			{Key: gocui.KeyEsc, Handler: self.onEsc},
			{Key: gocui.KeyTab, Handler: self.onTab},
			{Key: gocui.KeyCtrlA, Handler: self.onToggleAny},
		}
	}
}
