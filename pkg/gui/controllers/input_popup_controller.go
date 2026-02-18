package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// InputPopupController handles all generic input popup keybindings.
type InputPopupController struct {
	baseController
	getContext func() *context.InputPopupContext
	onEnter    func() error
	onEsc      func() error
	onTab      func() error
}

var _ types.IController = &InputPopupController{}

// InputPopupControllerOpts holds the callbacks injected during wiring.
type InputPopupControllerOpts struct {
	GetContext func() *context.InputPopupContext
	OnEnter    func() error
	OnEsc      func() error
	OnTab      func() error
}

// NewInputPopupController creates an InputPopupController.
func NewInputPopupController(opts InputPopupControllerOpts) *InputPopupController {
	return &InputPopupController{
		getContext: opts.GetContext,
		onEnter:    opts.OnEnter,
		onEsc:      opts.OnEsc,
		onTab:      opts.OnTab,
	}
}

// Context returns the context this controller is attached to.
func (self *InputPopupController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns the keybinding producer for the input popup.
func (self *InputPopupController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			{Key: gocui.KeyEnter, Handler: self.onEnter},
			{Key: gocui.KeyEsc, Handler: self.onEsc},
			{Key: gocui.KeyTab, Handler: self.onTab},
		}
	}
}
