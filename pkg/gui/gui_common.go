package gui

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// Adapter methods that implement types.IGuiCommon (and helpers.IGuiCommon
// which embeds it plus Contexts()).

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

func (gui *Gui) DeleteView(name string) {
	if gui.g != nil {
		gui.g.DeleteView(name)
	}
}

func (gui *Gui) CurrentContext() types.Context {
	return gui.contextMgr.ContextByKey(gui.contextMgr.Current())
}

func (gui *Gui) CurrentContextKey() types.ContextKey {
	return gui.contextMgr.Current()
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

func (gui *Gui) SearchQueryActive() bool { return gui.contexts.Search.Query != "" }

func (gui *Gui) ContextByKey(key types.ContextKey) types.Context {
	return gui.contextMgr.ContextByKey(key)
}

func (gui *Gui) PushContextByKey(key types.ContextKey)    { gui.pushContextByKey(key) }
func (gui *Gui) ReplaceContextByKey(key types.ContextKey) { gui.replaceContextByKey(key) }

// --- helpers.IGuiCommon adapter methods ---

func (gui *Gui) Contexts() *context.ContextTree { return gui.contexts }

// Refresh — convenience methods on *Gui for startup/background refresh.
// Not part of IGuiCommon; helpers call through Helpers() directly.
func (gui *Gui) RefreshNotes(preserve bool)   { gui.helpers.Notes().FetchNotesForCurrentTab(preserve) }
func (gui *Gui) RefreshTags(preserve bool)    { gui.helpers.Tags().RefreshTags(preserve) }
func (gui *Gui) RefreshQueries(preserve bool) { gui.helpers.Queries().RefreshQueries(preserve) }
func (gui *Gui) RefreshParents(preserve bool) { gui.helpers.Queries().RefreshParents(preserve) }
func (gui *Gui) RefreshAll()                  { gui.helpers.Refresh().RefreshAll() }

// State access
func (gui *Gui) GetInputPopupCompletion() *types.CompletionState {
	return gui.contexts.InputPopup.Completion
}

func (gui *Gui) SetCursorEnabled(enabled bool) {
	if gui.g != nil {
		gui.g.Cursor = enabled
	}
}

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

// Context state
func (gui *Gui) PreviousContextKey() types.ContextKey {
	return gui.contextMgr.Previous()
}

// Compile-time assertion that *Gui satisfies types.IGuiCommon.
var _ types.IGuiCommon = &Gui{}
