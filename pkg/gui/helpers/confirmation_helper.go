package helpers

// ConfirmationHelper manages confirmation, menu, and prompt dialogs.
type ConfirmationHelper struct {
	c *HelperCommon
}

// NewConfirmationHelper creates a new ConfirmationHelper.
func NewConfirmationHelper(c *HelperCommon) *ConfirmationHelper {
	return &ConfirmationHelper{c: c}
}
