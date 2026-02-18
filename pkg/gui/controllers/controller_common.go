package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/types"
)

// IGuiCommon is the interface controllers use to interact with the GUI.
// This avoids importing the gui package directly.
type IGuiCommon interface {
	CurrentContext() types.Context
	CurrentContextKey() types.ContextKey
	PushContext(ctx types.Context, opts types.OnFocusOpts)
	PushContextByKey(key types.ContextKey)
	PopContext()
	ReplaceContext(ctx types.Context)
	PopupActive() bool
	SearchQueryActive() bool
	ContextByKey(key types.ContextKey) types.Context
	GetView(name string) *gocui.View
	Render()
}

// IHelpers provides access to helper instances.
// This will be expanded as helpers are implemented.
type IHelpers interface{}

// ControllerCommon provides shared dependencies for all controllers.
type ControllerCommon struct {
	guiCommon IGuiCommon
	ruinCmd   *commands.RuinCommand
	helpers   IHelpers
}

// NewControllerCommon creates a new ControllerCommon.
func NewControllerCommon(guiCommon IGuiCommon, ruinCmd *commands.RuinCommand, helpers IHelpers) *ControllerCommon {
	return &ControllerCommon{
		guiCommon: guiCommon,
		ruinCmd:   ruinCmd,
		helpers:   helpers,
	}
}

// GuiCommon returns the GUI common interface.
func (self *ControllerCommon) GuiCommon() IGuiCommon {
	return self.guiCommon
}

// RuinCmd returns the ruin command wrapper.
func (self *ControllerCommon) RuinCmd() *commands.RuinCommand {
	return self.ruinCmd
}

// Helpers returns the helpers aggregator.
func (self *ControllerCommon) Helpers() IHelpers {
	return self.helpers
}
