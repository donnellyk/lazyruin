package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// PaletteController handles keybindings for the command palette popup.
type PaletteController struct {
	baseController
	getContext  func() *context.PaletteContext
	onEnter     func() error
	onEsc       func() error
	onListClick func() error
}

var _ types.IController = &PaletteController{}

// PaletteControllerOpts holds the callbacks injected during wiring.
type PaletteControllerOpts struct {
	GetContext  func() *context.PaletteContext
	OnEnter     func() error
	OnEsc       func() error
	OnListClick func() error
}

// NewPaletteController creates a PaletteController.
func NewPaletteController(opts PaletteControllerOpts) *PaletteController {
	return &PaletteController{
		getContext:  opts.GetContext,
		onEnter:     opts.OnEnter,
		onEsc:       opts.OnEsc,
		onListClick: opts.OnListClick,
	}
}

// Context returns the context this controller is attached to.
func (self *PaletteController) Context() types.Context {
	return self.getContext()
}

// GetKeybindings returns keybindings for the palette popup.
func (self *PaletteController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{ViewName: "palette", Key: gocui.KeyEnter, Handler: self.onEnter},
		{ViewName: "palette", Key: gocui.KeyEsc, Handler: self.onEsc},
	}
}

// GetMouseKeybindings returns mouse bindings for the palette list.
func (self *PaletteController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: "paletteList",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.onListClick()
			},
		},
	}
}
