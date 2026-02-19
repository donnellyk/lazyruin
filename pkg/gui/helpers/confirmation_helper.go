package helpers

import "kvnd/lazyruin/pkg/gui/types"

// ConfirmationHelper manages confirmation, input, and error dialogs.
type ConfirmationHelper struct {
	c *HelperCommon
}

// NewConfirmationHelper creates a new ConfirmationHelper.
func NewConfirmationHelper(c *HelperCommon) *ConfirmationHelper {
	return &ConfirmationHelper{c: c}
}

// ShowConfirm opens a yes/no confirmation dialog.
func (self *ConfirmationHelper) ShowConfirm(title, message string, onConfirm func() error) {
	self.c.GuiCommon().ShowConfirm(title, message, onConfirm)
}

// ShowInput opens a text input dialog.
func (self *ConfirmationHelper) ShowInput(title, message string, onConfirm func(string) error) {
	self.c.GuiCommon().ShowInput(title, message, onConfirm)
}

// ShowError displays an error in the status bar.
func (self *ConfirmationHelper) ShowError(err error) {
	self.c.GuiCommon().ShowError(err)
}

// OpenInputPopup opens the generic input popup with completion support.
func (self *ConfirmationHelper) OpenInputPopup(config *types.InputPopupConfig) {
	self.c.Helpers().InputPopup().OpenInputPopup(config)
}

// ConfirmDelete shows a confirmation dialog and executes deleteFn on confirm.
// displayName is truncated to 30 characters. On success, onSuccess is called.
func (self *ConfirmationHelper) ConfirmDelete(
	entityType string,
	displayName string,
	deleteFn func() error,
	onSuccess func(),
) {
	name := displayName
	if len(name) > 30 {
		name = name[:30] + "..."
	}
	gui := self.c.GuiCommon()
	gui.ShowConfirm("Delete "+entityType, "Delete \""+name+"\"?", func() error {
		if err := deleteFn(); err != nil {
			gui.ShowError(err)
			return nil
		}
		onSuccess()
		return nil
	})
}
