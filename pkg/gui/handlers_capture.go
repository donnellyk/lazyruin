package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) openCapture(g *gocui.Gui, v *gocui.View) error {
	gui.state.CaptureMode = true
	gui.state.CaptureParent = nil
	gui.state.CaptureCompletion = NewCompletionState()
	gui.setContext(CaptureContext)
	return nil
}

func (gui *Gui) submitCapture(g *gocui.Gui, v *gocui.View) error {
	content := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	if content == "" {
		if gui.QuickCapture {
			return gocui.ErrQuit
		}
		return gui.closeCapture(g)
	}

	args := []string{"log", content}
	if gui.state.CaptureParent != nil {
		args = append(args, "--parent", gui.state.CaptureParent.UUID)
	}
	_, err := gui.ruinCmd.Execute(args...)
	if err != nil {
		if gui.QuickCapture {
			return gocui.ErrQuit
		}
		return gui.closeCapture(g)
	}

	if gui.QuickCapture {
		return gocui.ErrQuit
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
	if gui.QuickCapture {
		return gocui.ErrQuit
	}
	return gui.closeCapture(g)
}

func (gui *Gui) captureTab(g *gocui.Gui, v *gocui.View) error {
	state := gui.state.CaptureCompletion
	if state.Active {
		if isParentCompletion(v, state) {
			gui.acceptParentCompletion(v, state)
		} else {
			gui.acceptCompletion(v, state, gui.captureTriggers())
		}
	}
	return nil
}

func (gui *Gui) closeCapture(g *gocui.Gui) error {
	gui.state.CaptureMode = false
	gui.state.CaptureParent = nil
	gui.state.CaptureCompletion = NewCompletionState()
	g.Cursor = false
	gui.setContext(gui.state.PreviousContext)
	return nil
}
