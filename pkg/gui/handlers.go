package gui

import (
	"github.com/jesseduffield/gocui"
)

// Global handlers

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (gui *Gui) refresh(g *gocui.Gui, v *gocui.View) error {
	gui.RefreshAll()
	return nil
}
