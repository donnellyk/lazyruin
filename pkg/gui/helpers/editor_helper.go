package helpers

// EditorHelper manages editor suspension and resumption.
type EditorHelper struct {
	c *HelperCommon
}

// NewEditorHelper creates a new EditorHelper.
func NewEditorHelper(c *HelperCommon) *EditorHelper {
	return &EditorHelper{c: c}
}
