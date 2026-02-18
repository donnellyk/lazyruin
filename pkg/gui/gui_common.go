package gui

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/types"
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

// Adapter methods that satisfy IGuiCommon.
// These live here so gui_common.go is the single source of truth for the interface.

func (gui *Gui) Render() { gui.renderAll() }

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

func (gui *Gui) PushContextByKey(key types.ContextKey) { gui.setContext(key) }

// contextByKey looks up a types.Context by its ContextKey.
func (gui *Gui) contextByKey(key types.ContextKey) types.Context {
	for _, ctx := range gui.contexts.All() {
		if ctx.GetKey() == key {
			return ctx
		}
	}
	return nil
}

// Compile-time assertion that Gui satisfies IGuiCommon.
var _ IGuiCommon = &Gui{}
