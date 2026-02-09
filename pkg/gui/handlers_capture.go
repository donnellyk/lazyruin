package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) openCapture(g *gocui.Gui, v *gocui.View) error {
	gui.state.CaptureMode = true
	gui.state.CaptureCompletion = NewCompletionState()
	gui.setContext(CaptureContext)
	return nil
}

func (gui *Gui) submitCapture(g *gocui.Gui, v *gocui.View) error {
	content := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	if content == "" {
		return gui.closeCapture(g)
	}

	_, err := gui.ruinCmd.Execute("log", content)
	if err != nil {
		return gui.closeCapture(g)
	}

	gui.closeCapture(g)
	gui.refreshNotes(false)
	gui.refreshTags(false)
	return nil
}

func (gui *Gui) cancelCapture(g *gocui.Gui, v *gocui.View) error {
	if gui.state.CaptureCompletion.Active {
		gui.state.CaptureCompletion.Active = false
		gui.state.CaptureCompletion.Items = nil
		gui.state.CaptureCompletion.SelectedIndex = 0
		return nil
	}
	return gui.closeCapture(g)
}

func (gui *Gui) captureTab(g *gocui.Gui, v *gocui.View) error {
	if gui.state.CaptureCompletion.Active {
		gui.acceptCompletion(v, gui.state.CaptureCompletion, gui.captureTriggers())
	}
	return nil
}

func (gui *Gui) closeCapture(g *gocui.Gui) error {
	gui.state.CaptureMode = false
	gui.state.CaptureCompletion = NewCompletionState()
	g.Cursor = false
	gui.setContext(gui.state.PreviousContext)
	return nil
}
