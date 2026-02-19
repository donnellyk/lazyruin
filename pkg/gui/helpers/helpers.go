package helpers

// Helpers aggregates all helper instances for easy access from controllers.
type Helpers struct {
	refresh          *RefreshHelper
	notes            *NotesHelper
	noteActions      *NoteActionsHelper
	tags             *TagsHelper
	queries          *QueriesHelper
	editor           *EditorHelper
	confirmation     *ConfirmationHelper
	search           *SearchHelper
	clipboard        *ClipboardHelper
	preview          *PreviewHelper
	previewNav       *PreviewNavHelper
	previewLinks     *PreviewLinksHelper
	previewMutations *PreviewMutationsHelper
	previewLineOps   *PreviewLineOpsHelper
	previewInfo      *PreviewInfoHelper
	capture          *CaptureHelper
	pick             *PickHelper
	inputPopup       *InputPopupHelper
	snippet          *SnippetHelper
	calendar         *CalendarHelper
	contrib          *ContribHelper
	completion       *CompletionHelper
}

// NewHelpers creates a new Helpers aggregator.
func NewHelpers(common *HelperCommon) *Helpers {
	h := &Helpers{
		refresh:          NewRefreshHelper(common),
		notes:            NewNotesHelper(common),
		noteActions:      NewNoteActionsHelper(common),
		tags:             NewTagsHelper(common),
		queries:          NewQueriesHelper(common),
		editor:           NewEditorHelper(common),
		confirmation:     NewConfirmationHelper(common),
		search:           NewSearchHelper(common),
		clipboard:        NewClipboardHelper(common),
		preview:          NewPreviewHelper(common),
		previewNav:       NewPreviewNavHelper(common),
		previewLinks:     NewPreviewLinksHelper(common),
		previewMutations: NewPreviewMutationsHelper(common),
		previewLineOps:   NewPreviewLineOpsHelper(common),
		previewInfo:      NewPreviewInfoHelper(common),
		capture:          NewCaptureHelper(common),
		pick:             NewPickHelper(common),
		inputPopup:       NewInputPopupHelper(common),
		snippet:          NewSnippetHelper(common),
		calendar:         NewCalendarHelper(common),
		contrib:          NewContribHelper(common),
		completion:       NewCompletionHelper(common),
	}
	common.SetHelpers(h)
	return h
}

// Accessor methods — satisfy controllers.IHelpers interface.

func (h *Helpers) Refresh() *RefreshHelper                   { return h.refresh }
func (h *Helpers) Notes() *NotesHelper                       { return h.notes }
func (h *Helpers) NoteActions() *NoteActionsHelper           { return h.noteActions }
func (h *Helpers) Tags() *TagsHelper                         { return h.tags }
func (h *Helpers) Queries() *QueriesHelper                   { return h.queries }
func (h *Helpers) Editor() *EditorHelper                     { return h.editor }
func (h *Helpers) Confirmation() *ConfirmationHelper         { return h.confirmation }
func (h *Helpers) Search() *SearchHelper                     { return h.search }
func (h *Helpers) Clipboard() *ClipboardHelper               { return h.clipboard }
func (h *Helpers) Preview() *PreviewHelper                   { return h.preview }
func (h *Helpers) PreviewNav() *PreviewNavHelper             { return h.previewNav }
func (h *Helpers) PreviewLinks() *PreviewLinksHelper         { return h.previewLinks }
func (h *Helpers) PreviewMutations() *PreviewMutationsHelper { return h.previewMutations }
func (h *Helpers) PreviewLineOps() *PreviewLineOpsHelper     { return h.previewLineOps }
func (h *Helpers) PreviewInfo() *PreviewInfoHelper           { return h.previewInfo }
func (h *Helpers) Capture() *CaptureHelper                   { return h.capture }
func (h *Helpers) Pick() *PickHelper                         { return h.pick }
func (h *Helpers) InputPopup() *InputPopupHelper             { return h.inputPopup }
func (h *Helpers) Snippet() *SnippetHelper                   { return h.snippet }
func (h *Helpers) Calendar() *CalendarHelper                 { return h.calendar }
func (h *Helpers) Contrib() *ContribHelper                   { return h.contrib }
func (h *Helpers) Completion() *CompletionHelper             { return h.completion }
