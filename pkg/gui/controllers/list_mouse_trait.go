package controllers

import (
	"kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// ListMouseOpts configures mouse behavior for a list panel.
type ListMouseOpts struct {
	ViewName     string
	ClickMargin  int
	ItemCount    func() int
	SetSelection func(idx int)
	GetContext   func() types.Context
	GuiCommon    func() IGuiCommon
}

// ListMouseBindings returns the standard mouse bindings for a list panel:
// click-to-select, wheel-up, and wheel-down.
func ListMouseBindings(opts ListMouseOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: opts.ViewName,
			Key:      gocui.MouseLeft,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				v := opts.GuiCommon().GetView(opts.ViewName)
				if v == nil {
					return nil
				}
				idx := helpers.ListClickIndex(v, opts.ClickMargin)
				if idx >= 0 && idx < opts.ItemCount() {
					opts.SetSelection(idx)
				}
				opts.GuiCommon().PushContext(opts.GetContext(), types.OnFocusOpts{})
				return nil
			},
		},
		{
			ViewName: opts.ViewName,
			Key:      gocui.MouseWheelDown,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				if v := opts.GuiCommon().GetView(opts.ViewName); v != nil {
					helpers.ScrollViewport(v, 3)
				}
				return nil
			},
		},
		{
			ViewName: opts.ViewName,
			Key:      gocui.MouseWheelUp,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				if v := opts.GuiCommon().GetView(opts.ViewName); v != nil {
					helpers.ScrollViewport(v, -3)
				}
				return nil
			},
		},
	}
}
