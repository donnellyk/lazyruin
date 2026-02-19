package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// SnippetEditorController handles keybindings for the snippet editor popup.
// The popup has two views: snippetName (top) and snippetExpansion (bottom).
type SnippetEditorController struct {
	baseController
	getContext       func() *context.SnippetEditorContext
	onEsc            func() error // same for both views
	onTab            func() error // same for both views
	onEnterName      func() error // Enter in name view = Tab
	onEnterExpansion func() error // Enter in expansion view = save
	onClickName      func() error
	onClickExpansion func() error
}

var _ types.IController = &SnippetEditorController{}

// SnippetEditorControllerOpts holds the callbacks injected during wiring.
type SnippetEditorControllerOpts struct {
	GetContext       func() *context.SnippetEditorContext
	OnEsc            func() error
	OnTab            func() error
	OnEnterName      func() error
	OnEnterExpansion func() error
	OnClickName      func() error
	OnClickExpansion func() error
}

// NewSnippetEditorController creates a SnippetEditorController.
func NewSnippetEditorController(opts SnippetEditorControllerOpts) *SnippetEditorController {
	return &SnippetEditorController{
		getContext:       opts.GetContext,
		onEsc:            opts.OnEsc,
		onTab:            opts.OnTab,
		onEnterName:      opts.OnEnterName,
		onEnterExpansion: opts.OnEnterExpansion,
		onClickName:      opts.OnClickName,
		onClickExpansion: opts.OnClickExpansion,
	}
}

// Context returns the context this controller is attached to.
func (self *SnippetEditorController) Context() types.Context {
	return self.getContext()
}

// GetKeybindings returns keybindings for the snippet editor.
func (self *SnippetEditorController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		// Esc and Tab are the same for both views (no ViewName = all views)
		{Key: gocui.KeyEsc, Handler: self.onEsc},
		{Key: gocui.KeyTab, Handler: self.onTab},
		// Enter is view-specific
		{ViewName: "snippetName", Key: gocui.KeyEnter, Handler: self.onEnterName},
		{ViewName: "snippetExpansion", Key: gocui.KeyEnter, Handler: self.onEnterExpansion},
	}
}

// GetMouseKeybindings returns mouse bindings for the snippet editor views.
func (self *SnippetEditorController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: "snippetName",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.onClickName()
			},
		},
		{
			ViewName: "snippetExpansion",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return self.onClickExpansion()
			},
		},
	}
}
