package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) openCapture(g *gocui.Gui, v *gocui.View) error {
	if gui.popupActive() {
		return nil
	}
	gui.state.CaptureParent = nil
	gui.state.CaptureCompletion = types.NewCompletionState()
	gui.pushContextByKey("capture")
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
	gui.RefreshNotes(false)
	gui.RefreshTags(false)
	return nil
}

func (gui *Gui) cancelCapture(g *gocui.Gui, v *gocui.View) error {
	if gui.state.CaptureCompletion.Active {
		gui.state.CaptureCompletion.Dismiss()
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
		if isAbbreviationCompletion(v, state) {
			gui.acceptAbbreviationInCapture(v, state)
		} else if isParentCompletion(v, state) {
			gui.acceptParentCompletion(v, state)
		} else {
			gui.acceptCompletion(v, state, gui.captureTriggers())
		}
		gui.renderCaptureTextArea(v)
	}
	return nil
}

func (gui *Gui) closeCapture(g *gocui.Gui) error {
	gui.state.CaptureParent = nil
	gui.state.CaptureCompletion = types.NewCompletionState()
	g.Cursor = false
	gui.popContext()
	return nil
}
