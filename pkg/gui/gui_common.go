package gui

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// IGuiCommon defines the interface that controllers and helpers use
// to interact with the GUI without depending on the full Gui struct.
type IGuiCommon interface {
	PushContext(ctx types.Context, opts types.OnFocusOpts)
	PopContext()
	ReplaceContext(ctx types.Context)
	CurrentContext() types.Context
	CurrentContextKey() types.ContextKey
	PopupActive() bool
	SearchQueryActive() bool
	ContextByKey(key types.ContextKey) types.Context
	PushContextByKey(key types.ContextKey)

	GetView(name string) *gocui.View
	Render()
	Update(func() error)
	Suspend() error
	Resume() error
}

// Adapter methods that satisfy both IGuiCommon interfaces
// (gui package IGuiCommon for controllers, and helpers.IGuiCommon for helpers).

func (gui *Gui) Render()    { gui.helpers.Refresh().RenderAll() }
func (gui *Gui) RenderAll() { gui.helpers.Refresh().RenderAll() }

func (gui *Gui) Update(fn func() error) {
	if gui.g == nil {
		return
	}
	gui.g.Update(func(_ *gocui.Gui) error { return fn() })
}

func (gui *Gui) Suspend() error { return gui.g.Suspend() }
func (gui *Gui) Resume() error  { return gui.g.Resume() }

func (gui *Gui) GetView(name string) *gocui.View {
	if gui.g == nil {
		return nil
	}
	v, _ := gui.g.View(name)
	return v
}

func (gui *Gui) CurrentContext() types.Context {
	return gui.contextByKey(gui.state.currentContext())
}

func (gui *Gui) CurrentContextKey() types.ContextKey {
	return gui.state.currentContext()
}

func (gui *Gui) PushContext(ctx types.Context, opts types.OnFocusOpts) {
	if ctx != nil {
		gui.setContext(ctx.GetKey())
	}
}

func (gui *Gui) PopContext() { gui.popContext() }

func (gui *Gui) ReplaceContext(ctx types.Context) {
	if ctx != nil {
		gui.replaceContext(ctx.GetKey())
	}
}

func (gui *Gui) PopupActive() bool { return gui.overlayActive() }

func (gui *Gui) SearchQueryActive() bool { return gui.state.SearchQuery != "" }

func (gui *Gui) ContextByKey(key types.ContextKey) types.Context {
	return gui.contextByKey(key)
}

func (gui *Gui) PushContextByKey(key types.ContextKey)    { gui.setContext(key) }
func (gui *Gui) ReplaceContextByKey(key types.ContextKey) { gui.replaceContext(key) }

// contextByKey looks up a types.Context by its ContextKey.
func (gui *Gui) contextByKey(key types.ContextKey) types.Context {
	for _, ctx := range gui.contexts.All() {
		if ctx.GetKey() == key {
			return ctx
		}
	}
	return nil
}

// --- helpers.IGuiCommon adapter methods ---

// Context access
func (gui *Gui) Contexts() *context.ContextTree { return gui.contexts }

// Rendering
func (gui *Gui) RenderNotes()      { gui.renderNotes() }
func (gui *Gui) RenderTags()       { gui.renderTags() }
func (gui *Gui) RenderQueries()    { gui.renderQueries() }
func (gui *Gui) RenderPreview()    { gui.renderPreview() }
func (gui *Gui) UpdateNotesTab()   { gui.updateNotesTab() }
func (gui *Gui) UpdateTagsTab()    { gui.updateTagsTab() }
func (gui *Gui) UpdateQueriesTab() { gui.updateQueriesTab() }
func (gui *Gui) UpdateStatusBar()  { gui.updateStatusBar() }

// Dialogs
func (gui *Gui) ShowConfirm(title, message string, onConfirm func() error) {
	gui.showConfirm(title, message, onConfirm)
}
func (gui *Gui) ShowInput(title, message string, onConfirm func(string) error) {
	gui.showInput(title, message, onConfirm)
}
func (gui *Gui) ShowError(err error) { gui.showError(err) }
func (gui *Gui) OpenInputPopup(config *types.InputPopupConfig) {
	gui.openInputPopup(config)
}

// Refresh — delegates to domain helpers.
func (gui *Gui) RefreshNotes(preserve bool)   { gui.helpers.Notes().FetchNotesForCurrentTab(preserve) }
func (gui *Gui) RefreshTags(preserve bool)    { gui.helpers.Tags().RefreshTags(preserve) }
func (gui *Gui) RefreshQueries(preserve bool) { gui.helpers.Queries().RefreshQueries(preserve) }
func (gui *Gui) RefreshParents(preserve bool) { gui.helpers.Queries().RefreshParents(preserve) }
func (gui *Gui) RefreshAll()                  { gui.helpers.Refresh().RefreshAll() }

// State access
func (gui *Gui) GetInputPopupCompletion() *types.CompletionState {
	return gui.state.InputPopupCompletion
}

// Search
func (gui *Gui) BuildSearchOptions() commands.SearchOptions {
	return gui.buildSearchOptions()
}
func (gui *Gui) SetSearchQuery(query string) { gui.state.SearchQuery = query }
func (gui *Gui) GetSearchQuery() string      { return gui.state.SearchQuery }
func (gui *Gui) SetSearchCompletion(state *types.CompletionState) {
	gui.state.SearchCompletion = state
}
func (gui *Gui) SetCursorEnabled(enabled bool) {
	if gui.g != nil {
		gui.g.Cursor = enabled
	}
}
func (gui *Gui) AmbientDateCandidates() func(string) []types.CompletionItem {
	return ambientDateCandidates
}

// Preview
func (gui *Gui) PreviewPushNavHistory()        { gui.preview.pushNavHistory() }
func (gui *Gui) PreviewReloadContent()         { gui.preview.reloadContent() }
func (gui *Gui) PreviewUpdatePreviewForNotes() { gui.preview.updatePreviewForNotes() }
func (gui *Gui) PreviewUpdatePreviewCardList(title string, fetch func() ([]models.Note, error)) {
	gui.preview.updatePreviewCardList(title, fetch)
}
func (gui *Gui) PreviewCurrentCard() *models.Note {
	return gui.preview.currentPreviewCard()
}
func (gui *Gui) SetPreviewCards(cards []models.Note, selectedIdx int, title string) {
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = cards
	gui.state.Preview.SelectedCardIndex = selectedIdx
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = title
	}
	gui.renderPreview()
}
func (gui *Gui) SetPreviewPickResults(results []models.PickResult, selectedIdx int, cursorLine int, scrollOffset int, title string) {
	gui.state.Preview.Mode = PreviewModePickResults
	gui.state.Preview.PickResults = results
	gui.state.Preview.SelectedCardIndex = selectedIdx
	gui.state.Preview.CursorLine = cursorLine
	gui.state.Preview.ScrollOffset = scrollOffset
	if gui.views.Preview != nil {
		gui.views.Preview.Title = title
	}
	gui.renderPreview()
}

// Completion candidates
func (gui *Gui) TagCandidates(filter string) []types.CompletionItem {
	return gui.tagCandidates(filter)
}
func (gui *Gui) CurrentCardTagCandidates(filter string) []types.CompletionItem {
	return gui.currentCardTagCandidates(filter)
}
func (gui *Gui) ParentCandidatesFor(state *types.CompletionState) func(string) []types.CompletionItem {
	return gui.parentCandidatesFor(state)
}

// Compile-time assertion that Gui satisfies IGuiCommon.
var _ IGuiCommon = &Gui{}
