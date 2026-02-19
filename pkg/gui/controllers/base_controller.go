package controllers

import (
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// baseController is the null object base for all controllers.
// All methods return nil by default. Concrete controllers embed
// this and override selectively.
type baseController struct{}

func (self *baseController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding { return nil }
func (self *baseController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return nil
}
func (self *baseController) GetOnRenderToMain() func()                   { return nil }
func (self *baseController) GetOnFocus() func(types.OnFocusOpts)         { return nil }
func (self *baseController) GetOnFocusLost() func(types.OnFocusLostOpts) { return nil }
