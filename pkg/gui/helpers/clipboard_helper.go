package helpers

// ClipboardHelper manages clipboard operations.
type ClipboardHelper struct {
	c *HelperCommon
}

// NewClipboardHelper creates a new ClipboardHelper.
func NewClipboardHelper(c *HelperCommon) *ClipboardHelper {
	return &ClipboardHelper{c: c}
}
