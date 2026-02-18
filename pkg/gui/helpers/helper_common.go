package helpers

import "kvnd/lazyruin/pkg/commands"

// IGuiCommon is the interface helpers use to interact with the GUI.
// Avoids importing the gui package directly.
type IGuiCommon interface {
	Render()
	Update(func() error)
}

// HelperCommon provides shared dependencies for all helpers.
type HelperCommon struct {
	ruinCmd   *commands.RuinCommand
	guiCommon IGuiCommon
}

// NewHelperCommon creates a new HelperCommon.
func NewHelperCommon(ruinCmd *commands.RuinCommand, guiCommon IGuiCommon) *HelperCommon {
	return &HelperCommon{
		ruinCmd:   ruinCmd,
		guiCommon: guiCommon,
	}
}

// RuinCmd returns the ruin command wrapper.
func (self *HelperCommon) RuinCmd() *commands.RuinCommand {
	return self.ruinCmd
}

// GuiCommon returns the GUI common interface.
func (self *HelperCommon) GuiCommon() IGuiCommon {
	return self.guiCommon
}
