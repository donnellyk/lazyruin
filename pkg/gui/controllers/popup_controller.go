package controllers

import "kvnd/lazyruin/pkg/gui/types"

// PopupController is a generic controller for popup contexts
// that need a fixed set of keybindings (e.g. Enter/Esc/Tab).
// Replaces SearchController, InputPopupController, CaptureController,
// and PickController.
type PopupController[C types.Context] struct {
	baseController
	getContext func() C
	bindings   []*types.Binding
}

var _ types.IController = &PopupController[types.Context]{}

// NewPopupController creates a PopupController.
func NewPopupController[C types.Context](
	getContext func() C,
	bindings []*types.Binding,
) *PopupController[C] {
	return &PopupController[C]{
		getContext: getContext,
		bindings:   bindings,
	}
}

// Context returns the context this controller is attached to.
func (self *PopupController[C]) Context() types.Context {
	return self.getContext()
}

// GetKeybindings returns the keybindings for this popup.
func (self *PopupController[C]) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return self.bindings
}
