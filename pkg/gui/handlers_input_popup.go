package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
	"strings"

	"github.com/jesseduffield/gocui"
)

// openInputPopup opens the generic input popup with the given config.
func (gui *Gui) openInputPopup(config *types.InputPopupConfig) {
	gui.state.InputPopupCompletion = types.NewCompletionState()
	gui.state.InputPopupSeedDone = false
	gui.state.InputPopupConfig = config
	gui.pushContextByKey("inputPopup")
}

// closeInputPopup closes the input popup and restores focus.
func (gui *Gui) closeInputPopup() {
	gui.state.InputPopupCompletion = types.NewCompletionState()
	gui.state.InputPopupConfig = nil
	gui.g.Cursor = false
	gui.g.DeleteView(InputPopupView)
	gui.g.DeleteView(InputPopupSuggestView)
	gui.popContext()
}

// inputPopupEnter handles Enter in the input popup.
func (gui *Gui) inputPopupEnter(g *gocui.Gui, v *gocui.View) error {
	state := gui.state.InputPopupCompletion
	config := gui.state.InputPopupConfig

	raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	var item *types.CompletionItem
	if state.Active && len(state.Items) > 0 {
		selected := state.Items[state.SelectedIndex]
		item = &selected
	}

	gui.closeInputPopup()
	if (raw == "" && item == nil) || config == nil || config.OnAccept == nil {
		return nil
	}
	return config.OnAccept(raw, item)
}

// inputPopupTab accepts the current completion in the input popup.
func (gui *Gui) inputPopupTab(g *gocui.Gui, v *gocui.View) error {
	if gui.state.InputPopupCompletion.Active && len(gui.state.InputPopupCompletion.Items) > 0 {
		return gui.inputPopupEnter(g, v)
	}
	return nil
}

// inputPopupEsc cancels the input popup (first press dismisses suggestions).
func (gui *Gui) inputPopupEsc(g *gocui.Gui, v *gocui.View) error {
	if gui.state.InputPopupCompletion.Active {
		gui.state.InputPopupCompletion.Dismiss()
		return nil
	}
	gui.closeInputPopup()
	return nil
}
