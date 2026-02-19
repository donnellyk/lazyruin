package helpers

// Helpers aggregates all helper instances for easy access from controllers.
type Helpers struct {
	refresh      *RefreshHelper
	notes        *NotesHelper
	noteActions  *NoteActionsHelper
	tags         *TagsHelper
	queries      *QueriesHelper
	editor       *EditorHelper
	confirmation *ConfirmationHelper
	search       *SearchHelper
	clipboard    *ClipboardHelper
	preview      *PreviewHelper
	capture      *CaptureHelper
	pick         *PickHelper
	inputPopup   *InputPopupHelper
}

// NewHelpers creates a new Helpers aggregator.
func NewHelpers(common *HelperCommon) *Helpers {
	h := &Helpers{
		refresh:      NewRefreshHelper(common),
		notes:        NewNotesHelper(common),
		noteActions:  NewNoteActionsHelper(common),
		tags:         NewTagsHelper(common),
		queries:      NewQueriesHelper(common),
		editor:       NewEditorHelper(common),
		confirmation: NewConfirmationHelper(common),
		search:       NewSearchHelper(common),
		clipboard:    NewClipboardHelper(common),
		preview:      NewPreviewHelper(common),
		capture:      NewCaptureHelper(common),
		pick:         NewPickHelper(common),
		inputPopup:   NewInputPopupHelper(common),
	}
	common.SetHelpers(h)
	return h
}

// Accessor methods — satisfy controllers.IHelpers interface.

func (h *Helpers) Refresh() *RefreshHelper           { return h.refresh }
func (h *Helpers) Notes() *NotesHelper               { return h.notes }
func (h *Helpers) NoteActions() *NoteActionsHelper   { return h.noteActions }
func (h *Helpers) Tags() *TagsHelper                 { return h.tags }
func (h *Helpers) Queries() *QueriesHelper           { return h.queries }
func (h *Helpers) Editor() *EditorHelper             { return h.editor }
func (h *Helpers) Confirmation() *ConfirmationHelper { return h.confirmation }
func (h *Helpers) Search() *SearchHelper             { return h.search }
func (h *Helpers) Clipboard() *ClipboardHelper       { return h.clipboard }
func (h *Helpers) Preview() *PreviewHelper           { return h.preview }
func (h *Helpers) Capture() *CaptureHelper           { return h.capture }
func (h *Helpers) Pick() *PickHelper                 { return h.pick }
func (h *Helpers) InputPopup() *InputPopupHelper     { return h.inputPopup }
