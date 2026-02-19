package helpers

import "kvnd/lazyruin/pkg/gui/types"

// InputPopupHelper encapsulates the input popup logic.
type InputPopupHelper struct {
	c *HelperCommon
}

func NewInputPopupHelper(c *HelperCommon) *InputPopupHelper {
	return &InputPopupHelper{c: c}
}

// OpenInputPopup opens the input popup with the given config.
func (self *InputPopupHelper) OpenInputPopup(config *types.InputPopupConfig) {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().InputPopup
	ctx.Completion = types.NewCompletionState()
	ctx.SeedDone = false
	ctx.Config = config
	gui.PushContextByKey("inputPopup")
}

// CloseInputPopup closes the input popup and restores focus.
func (self *InputPopupHelper) CloseInputPopup() {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().InputPopup
	ctx.Completion = types.NewCompletionState()
	ctx.Config = nil
	gui.SetCursorEnabled(false)
	gui.DeleteView("inputPopup")
	gui.DeleteView("inputPopupSuggest")
	gui.PopContext()
}

// HandleEnter handles Enter in the input popup.
func (self *InputPopupHelper) HandleEnter(raw string, item *types.CompletionItem) error {
	ctx := self.c.GuiCommon().Contexts().InputPopup
	config := ctx.Config
	self.CloseInputPopup()
	if (raw == "" && item == nil) || config == nil || config.OnAccept == nil {
		return nil
	}
	return config.OnAccept(raw, item)
}

// HandleEsc handles Esc in the input popup.
func (self *InputPopupHelper) HandleEsc() error {
	ctx := self.c.GuiCommon().Contexts().InputPopup
	if ctx.Completion.Active {
		ctx.Completion.Dismiss()
		return nil
	}
	self.CloseInputPopup()
	return nil
}
