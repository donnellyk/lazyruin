package controllers

import "kvnd/lazyruin/pkg/gui/types"

// AttachController wires a controller's keybinding/focus/render producers
// into its context's aggregation points.
func AttachController(ctrl types.IController) {
	ctx := ctrl.Context()
	ctx.AddKeybindingsFn(ctrl.GetKeybindings)
	ctx.AddMouseKeybindingsFn(ctrl.GetMouseKeybindings)
	if f := ctrl.GetOnFocus(); f != nil {
		ctx.AddOnFocusFn(f)
	}
	if f := ctrl.GetOnFocusLost(); f != nil {
		ctx.AddOnFocusLostFn(f)
	}
	if f := ctrl.GetOnRenderToMain(); f != nil {
		ctx.AddOnRenderToMainFn(f)
	}
}
