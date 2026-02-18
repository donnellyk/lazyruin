package helpers

// NotesHelper encapsulates note domain operations.
type NotesHelper struct {
	c *HelperCommon
}

// NewNotesHelper creates a new NotesHelper.
func NewNotesHelper(c *HelperCommon) *NotesHelper {
	return &NotesHelper{c: c}
}
