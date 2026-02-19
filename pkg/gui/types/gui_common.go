package types

import (
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// IGuiCommon is the authoritative interface for GUI operations.
// Controllers, helpers, and handler code interact with the GUI through
// this interface instead of depending on the full Gui struct.
//
// The helpers package extends this with Contexts() *context.ContextTree
// (which can't live here due to the types↔context import cycle).
type IGuiCommon interface {
	// Rendering
	Update(func() error)
	RenderNotes()
	RenderTags()
	RenderQueries()
	RenderPreview()
	RenderAll()
	UpdateNotesTab()
	UpdateTagsTab()
	UpdateQueriesTab()
	UpdateStatusBar()

	// Navigation
	CurrentContext() Context
	CurrentContextKey() ContextKey
	PushContext(ctx Context, opts OnFocusOpts)
	PushContextByKey(key ContextKey)
	PopContext()
	ReplaceContext(ctx Context)
	ReplaceContextByKey(key ContextKey)
	ContextByKey(key ContextKey) Context
	PopupActive() bool
	SearchQueryActive() bool

	// Dialogs
	ShowConfirm(title, message string, onConfirm func() error)
	ShowInput(title, message string, onConfirm func(string) error)
	ShowError(err error)
	OpenInputPopup(config *InputPopupConfig)
	ShowMenuDialog(title string, items []MenuItem)

	// Refresh
	RefreshNotes(preserve bool)
	RefreshTags(preserve bool)
	RefreshQueries(preserve bool)
	RefreshParents(preserve bool)
	RefreshAll()

	// State access
	GetInputPopupCompletion() *CompletionState

	// Search
	BuildSearchOptions() commands.SearchOptions
	SetSearchQuery(query string)
	GetSearchQuery() string
	SetSearchCompletion(state *CompletionState)
	SetCursorEnabled(enabled bool)
	AmbientDateCandidates() func(string) []CompletionItem

	// Preview
	PreviewPushNavHistory()
	PreviewReloadContent()
	PreviewUpdatePreviewForNotes()
	PreviewUpdatePreviewCardList(title string, fetch func() ([]models.Note, error))
	PreviewCurrentCard() *models.Note
	SetPreviewCards(cards []models.Note, selectedIdx int, title string)
	SetPreviewPickResults(results []models.PickResult, selectedIdx int, cursorLine int, scrollOffset int, title string)

	// Completion candidates
	TagCandidates(filter string) []CompletionItem
	CurrentCardTagCandidates(filter string) []CompletionItem
	ParentCandidatesFor(state *CompletionState) func(string) []CompletionItem

	// Editor
	Suspend() error
	Resume() error

	// View access
	GetView(name string) *gocui.View
	DeleteView(name string)

	// Preview rendering
	BuildCardContent(note models.Note, width int) []string

	// Context state
	PreviousContextKey() ContextKey

	// Date candidates
	AtDateCandidates(filter string) []CompletionItem
}
