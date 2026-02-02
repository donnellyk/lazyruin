package gui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/commands"
)

// Gui manages the terminal user interface.
type Gui struct {
	g       *gocui.Gui
	ruinCmd *commands.RuinCommand
}

// NewGui creates a new Gui instance.
func NewGui(ruinCmd *commands.RuinCommand) *Gui {
	return &Gui{
		ruinCmd: ruinCmd,
	}
}

// Run starts the GUI event loop.
func (gui *Gui) Run() error {
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		return err
	}
	defer g.Close()

	gui.g = g
	g.Mouse = true
	g.Cursor = false
	g.SetLayout(gui.layout)

	if err := gui.setupKeybindings(); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}

// layout is called on every render to set up views.
func (gui *Gui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// For now, just create a simple main view
	if v, err := g.SetView("main", 0, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " LazyRuin "
		v.Wrap = true
		fmt.Fprintf(v, "Welcome to LazyRuin!\n\n")
		fmt.Fprintf(v, "Vault: %s\n\n", gui.ruinCmd.VaultPath())
		fmt.Fprintf(v, "Press 'q' to quit.\n")
	}

	return nil
}

// setupKeybindings configures keyboard shortcuts.
func (gui *Gui) setupKeybindings() error {
	// Quit
	if err := gui.g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
