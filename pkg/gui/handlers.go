package gui

import (
	"kvnd/lazyruin/pkg/commands"

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

// buildSearchOptions returns SearchOptions based on current preview toggle state
func (gui *Gui) BuildSearchOptions() commands.SearchOptions {
	return commands.SearchOptions{
		IncludeContent:  true,
		StripGlobalTags: !gui.contexts.Preview.ShowGlobalTags,
		StripTitle:      !gui.contexts.Preview.ShowTitle,
	}
}
