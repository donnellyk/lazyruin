package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// CaptureController handles all capture popup keybindings.
type CaptureController struct {
	baseController
	getContext func() *context.CaptureContext
	onSubmit   func() error
	onEsc      func() error
	onTab      func() error
}

var _ types.IController = &CaptureController{}

// CaptureControllerOpts holds the callbacks injected during wiring.
type CaptureControllerOpts struct {
	GetContext func() *context.CaptureContext
	OnSubmit   func() error
	OnEsc      func() error
	OnTab      func() error
}

// NewCaptureController creates a CaptureController.
func NewCaptureController(opts CaptureControllerOpts) *CaptureController {
	return &CaptureController{
		getContext: opts.GetContext,
		onSubmit:   opts.OnSubmit,
		onEsc:      opts.OnEsc,
		onTab:      opts.OnTab,
	}
}

// Context returns the context this controller is attached to.
func (self *CaptureController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns the keybinding producer for the capture popup.
func (self *CaptureController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			{Key: gocui.KeyCtrlS, Handler: self.onSubmit},
			{Key: gocui.KeyEsc, Handler: self.onEsc},
			{Key: gocui.KeyTab, Handler: self.onTab},
		}
	}
}
