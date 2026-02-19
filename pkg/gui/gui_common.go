package gui

import (
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// Adapter methods that implement types.IGuiCommon (and helpers.IGuiCommon
// which embeds it plus Contexts()).

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
		gui.pushContext(ctx)
	}
}

func (gui *Gui) PopContext() { gui.popContext() }

func (gui *Gui) ReplaceContext(ctx types.Context) {
	if ctx != nil {
		gui.replaceContext(ctx)
	}
}

func (gui *Gui) PopupActive() bool { return gui.overlayActive() }

func (gui *Gui) SearchQueryActive() bool { return gui.state.SearchQuery != "" }

func (gui *Gui) ContextByKey(key types.ContextKey) types.Context {
	return gui.contextByKey(key)
}

func (gui *Gui) PushContextByKey(key types.ContextKey)    { gui.pushContextByKey(key) }
func (gui *Gui) ReplaceContextByKey(key types.ContextKey) { gui.replaceContextByKey(key) }

// contextByKey looks up a types.Context by its types.ContextKey.
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

// Preview — delegates to PreviewHelper.
func (gui *Gui) PreviewPushNavHistory()        { gui.helpers.Preview().PushNavHistory() }
func (gui *Gui) PreviewReloadContent()         { gui.helpers.Preview().ReloadContent() }
func (gui *Gui) PreviewUpdatePreviewForNotes() { gui.helpers.Preview().UpdatePreviewForNotes() }
func (gui *Gui) PreviewUpdatePreviewCardList(title string, fetch func() ([]models.Note, error)) {
	gui.helpers.Preview().UpdatePreviewCardList(title, fetch)
}
func (gui *Gui) PreviewCurrentCard() *models.Note {
	return gui.helpers.Preview().CurrentPreviewCard()
}
func (gui *Gui) SetPreviewCards(cards []models.Note, selectedIdx int, title string) {
	gui.helpers.Preview().ShowCardList(title, cards)
}
func (gui *Gui) SetPreviewPickResults(results []models.PickResult, selectedIdx int, cursorLine int, scrollOffset int, title string) {
	gui.helpers.Preview().ShowPickResults(title, results)
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

// View access (already satisfies controllers.IGuiCommon; now also helpers.IGuiCommon)
// GetView is defined above.

// Dialogs (menu)
func (gui *Gui) ShowMenuDialog(title string, items []types.MenuItem) {
	gui.state.Dialog = &DialogState{
		Active:        true,
		Type:          "menu",
		Title:         title,
		MenuItems:     items,
		MenuSelection: 0,
	}
}

// Preview rendering
func (gui *Gui) BuildCardContent(note models.Note, width int) []string {
	return gui.buildCardContent(note, width)
}

// Context state
// CurrentContextKey is defined above (line 59).
func (gui *Gui) PreviousContextKey() types.ContextKey {
	return gui.state.previousContext()
}

// Date candidates
func (gui *Gui) AtDateCandidates(filter string) []types.CompletionItem {
	return atDateCandidates(filter)
}

// Compile-time assertion that *Gui satisfies types.IGuiCommon.
var _ types.IGuiCommon = &Gui{}
