package helpers

import (
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

// IGuiCommon extends types.IGuiCommon with Contexts(), which can't live
// in types/ due to the types↔context import cycle.
type IGuiCommon interface {
	types.IGuiCommon
	Contexts() *context.ContextTree
}

// HelperCommon provides shared dependencies for all helpers.
type HelperCommon struct {
	ruinCmd   *commands.RuinCommand
	guiCommon IGuiCommon
	helpers   *Helpers
}

// NewHelperCommon creates a new HelperCommon.
func NewHelperCommon(ruinCmd *commands.RuinCommand, guiCommon IGuiCommon) *HelperCommon {
	return &HelperCommon{
		ruinCmd:   ruinCmd,
		guiCommon: guiCommon,
	}
}

// SetHelpers sets the helpers reference (called after Helpers is constructed).
func (self *HelperCommon) SetHelpers(h *Helpers) {
	self.helpers = h
}

// RuinCmd returns the ruin command wrapper.
func (self *HelperCommon) RuinCmd() *commands.RuinCommand {
	return self.ruinCmd
}

// GuiCommon returns the GUI common interface.
func (self *HelperCommon) GuiCommon() IGuiCommon {
	return self.guiCommon
}

// Helpers returns the helpers aggregator (for cross-helper calls).
func (self *HelperCommon) Helpers() *Helpers {
	return self.helpers
}
