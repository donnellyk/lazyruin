package helpers

// Helpers aggregates all helper instances for easy access from controllers.
type Helpers struct {
	Refresh      *RefreshHelper
	Notes        *NotesHelper
	Editor       *EditorHelper
	Confirmation *ConfirmationHelper
	Search       *SearchHelper
	Clipboard    *ClipboardHelper
}

// NewHelpers creates a new Helpers aggregator.
func NewHelpers(common *HelperCommon) *Helpers {
	return &Helpers{
		Refresh:      NewRefreshHelper(common),
		Notes:        NewNotesHelper(common),
		Editor:       NewEditorHelper(common),
		Confirmation: NewConfirmationHelper(common),
		Search:       NewSearchHelper(common),
		Clipboard:    NewClipboardHelper(common),
	}
}
