package helpers

import (
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// IGuiCommon is the interface helpers use to interact with the GUI.
// Avoids importing the gui package directly.
type IGuiCommon interface {
	// Rendering
	Render()
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

	// Context access
	Contexts() *context.ContextTree

	// Navigation
	PushContextByKey(key types.ContextKey)
	PopContext()
	ReplaceContextByKey(key types.ContextKey)

	// Dialogs
	ShowConfirm(title, message string, onConfirm func() error)
	ShowInput(title, message string, onConfirm func(string) error)
	ShowError(err error)
	OpenInputPopup(config *types.InputPopupConfig)

	// Refresh
	RefreshNotes(preserve bool)
	RefreshTags(preserve bool)
	RefreshQueries(preserve bool)
	RefreshParents(preserve bool)
	RefreshAll()

	// State access
	GetInputPopupCompletion() *types.CompletionState

	// Search
	BuildSearchOptions() commands.SearchOptions
	SetSearchQuery(query string)
	GetSearchQuery() string
	SetSearchCompletion(state *types.CompletionState)
	PopupActive() bool
	SetCursorEnabled(enabled bool)
	AmbientDateCandidates() func(string) []types.CompletionItem

	// Preview
	PreviewPushNavHistory()
	PreviewReloadContent()
	PreviewUpdatePreviewForNotes()
	PreviewUpdatePreviewCardList(title string, fetch func() ([]models.Note, error))
	PreviewCurrentCard() *models.Note
	SetPreviewCards(cards []models.Note, selectedIdx int, title string)
	SetPreviewPickResults(results []models.PickResult, selectedIdx int, cursorLine int, scrollOffset int, title string)

	// Completion candidates
	TagCandidates(filter string) []types.CompletionItem
	CurrentCardTagCandidates(filter string) []types.CompletionItem
	ParentCandidatesFor(state *types.CompletionState) func(string) []types.CompletionItem

	// Editor
	Suspend() error
	Resume() error
}

// HelperCommon provides shared dependencies for all helpers.
type HelperCommon struct {
	ruinCmd   *commands.RuinCommand
	guiCommon IGuiCommon
	helpers   *Helpers
}

// NewHelperCommon creates a new HelperCommon.
func NewHelperCommon(ruinCmd *commands.RuinCommand, guiCommon IGuiCommon) *HelperCommon {
	return &HelperCommon{
		ruinCmd:   ruinCmd,
		guiCommon: guiCommon,
	}
}

// SetHelpers sets the helpers reference (called after Helpers is constructed).
func (self *HelperCommon) SetHelpers(h *Helpers) {
	self.helpers = h
}

// RuinCmd returns the ruin command wrapper.
func (self *HelperCommon) RuinCmd() *commands.RuinCommand {
	return self.ruinCmd
}

// GuiCommon returns the GUI common interface.
func (self *HelperCommon) GuiCommon() IGuiCommon {
	return self.guiCommon
}

// Helpers returns the helpers aggregator (for cross-helper calls).
func (self *HelperCommon) Helpers() *Helpers {
	return self.helpers
}
