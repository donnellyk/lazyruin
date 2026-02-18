package controllers

import "kvnd/lazyruin/pkg/gui/types"

// baseController is the null object base for all controllers.
// All methods return nil by default. Concrete controllers embed
// this and override selectively.
type baseController struct{}

func (self *baseController) GetKeybindingsFn() types.KeybindingsFn           { return nil }
func (self *baseController) GetMouseKeybindingsFn() types.MouseKeybindingsFn { return nil }
func (self *baseController) GetOnRenderToMain() func()                       { return nil }
func (self *baseController) GetOnFocus() func(types.OnFocusOpts)             { return nil }
func (self *baseController) GetOnFocusLost() func(types.OnFocusLostOpts)     { return nil }
