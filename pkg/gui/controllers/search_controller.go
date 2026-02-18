package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// SearchController handles all search popup keybindings.
type SearchController struct {
	baseController
	getContext func() *context.SearchContext
	onEnter    func() error
	onEsc      func() error
	onTab      func() error
}

var _ types.IController = &SearchController{}

// SearchControllerOpts holds the callbacks injected during wiring.
type SearchControllerOpts struct {
	GetContext func() *context.SearchContext
	OnEnter    func() error
	OnEsc      func() error
	OnTab      func() error
}

// NewSearchController creates a SearchController.
func NewSearchController(opts SearchControllerOpts) *SearchController {
	return &SearchController{
		getContext: opts.GetContext,
		onEnter:    opts.OnEnter,
		onEsc:      opts.OnEsc,
		onTab:      opts.OnTab,
	}
}

// Context returns the context this controller is attached to.
func (self *SearchController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns the keybinding producer for the search popup.
func (self *SearchController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			{Key: gocui.KeyEnter, Handler: self.onEnter},
			{Key: gocui.KeyEsc, Handler: self.onEsc},
			{Key: gocui.KeyTab, Handler: self.onTab},
		}
	}
}
