package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"

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

// WheelScrollBindings returns mouse-wheel scroll bindings for a view. Use
// for lists that need wheel scrolling but aren't wired up via the fuller
// ListMouseBindings (e.g. popup lists like the command palette, inbox
// browser, calendar notes, contribution notes, pick dialog).
func WheelScrollBindings(viewName string, guiCommon func() IGuiCommon) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: viewName,
			Key:      gocui.MouseWheelDown,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				if v := guiCommon().GetView(viewName); v != nil {
					helpers.ScrollViewport(v, 3)
				}
				return nil
			},
		},
		{
			ViewName: viewName,
			Key:      gocui.MouseWheelUp,
			Handler: func(_ gocui.ViewMouseBindingOpts) error {
				if v := guiCommon().GetView(viewName); v != nil {
					helpers.ScrollViewport(v, -3)
				}
				return nil
			},
		},
	}
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
