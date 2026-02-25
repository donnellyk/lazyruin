package gui

import (
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/controllers"
	helperspkg "kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
	"github.com/muesli/termenv"
)

// Gui manages the terminal user interface.
type Gui struct {
	g              *gocui.Gui
	views          *Views
	state          *GuiState
	config         *config.Config
	ruinCmd        *commands.RuinCommand
	stopBg         chan struct{}
	QuickCapture   bool // when true, open capture on start and quit on save
	darkBackground bool

	// New controller/context architecture (Phase 2+)
	contexts          *context.ContextTree
	contextMgr        *ContextMgr
	notesController   *controllers.NotesController
	tagsController    *controllers.TagsController
	queriesController *controllers.QueriesController
	globalController  *controllers.GlobalController

	// Shared helper/controller dependencies
	helpers          *helperspkg.Helpers
	controllerCommon *controllers.ControllerCommon
}

// NewGui creates a new Gui instance.
func NewGui(cfg *config.Config, ruinCmd *commands.RuinCommand) *Gui {
	gui := &Gui{
		config:     cfg,
		ruinCmd:    ruinCmd,
		views:      &Views{},
		state:      NewGuiState(),
		contexts:   &context.ContextTree{},
		contextMgr: NewContextMgr(),
	}
	// Wire shared helper/controller dependencies.
	helperCommon := helperspkg.NewHelperCommon(ruinCmd, gui.config, gui)
	gui.helpers = helperspkg.NewHelpers(helperCommon)
	gui.controllerCommon = controllers.NewControllerCommon(gui, ruinCmd, gui.helpers)

	gui.setupNotesContext()
	gui.setupTagsContext()
	gui.setupQueriesContext()
	gui.setupPreviewContext()
	gui.setupSearchContext()
	gui.setupCaptureContext()
	gui.setupPickContext()
	gui.setupInputPopupContext()
	gui.setupGlobalContext()
	gui.setupPaletteContext()
	gui.setupSnippetEditorContext()
	gui.setupCalendarContext()
	gui.setupContribContext()
	gui.setupPickDialogContext()
	return gui
}

// Run starts the GUI event loop.
func (gui *Gui) Run() error {
	// Detect terminal background before gocui takes over the terminal.
	gui.darkBackground = termenv.HasDarkBackground()

	err := gui.runMainLoop()
	if err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (gui *Gui) runMainLoop() error {
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
	})
	if err != nil {
		return err
	}
	defer g.Close()

	gui.g = g
	gui.views = &Views{} // Reset views for fresh layout
	g.Mouse = true
	g.Cursor = false
	g.ShowListFooter = true
	g.SetManager(gocui.ManagerFunc(gui.layout))

	if err := gui.setupKeybindings(); err != nil {
		return err
	}

	gui.stopBg = make(chan struct{})
	go gui.backgroundRefresh()

	err = g.MainLoop()
	close(gui.stopBg)
	return err
}

// backgroundRefresh polls for external changes every 30 seconds.
func (gui *Gui) backgroundRefresh() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-gui.stopBg:
			return
		case <-ticker.C:
			gui.g.Update(func(g *gocui.Gui) error {
				gui.backgroundRefreshData()
				return nil
			})
		}
	}
}

// backgroundRefreshData reloads sidebar lists without touching the preview.
// Preview content lives in CardList.Cards (or PickResults, Compose, etc.)
// which is separate from the sidebar data, so re-rendering it would only
// disturb the scroll offset and nav-history state.
func (gui *Gui) backgroundRefreshData() {
	gui.RefreshNotes(true)
	gui.RefreshTags(true)
	gui.RefreshQueries(true)
	gui.RefreshParents(true)
	gui.UpdateStatusBar()
}

// activateContext sets focus and re-renders lists for the given context.
// Per-context refresh/preview logic is handled by HandleFocus hooks.
func (gui *Gui) activateContext(ctx types.ContextKey) {
	viewName := gui.contextToView(ctx)
	gui.g.SetCurrentView(viewName)

	// When switching to a preview context, re-register only that context's
	// keybindings on the shared "preview" view.
	if context.IsPreviewContextKey(ctx) {
		gui.reregisterPreviewBindings()
	}

	// Re-render lists to update highlight visibility
	gui.RenderNotes()
	gui.RenderQueries()
	gui.RenderTags()

	gui.UpdateStatusBar()
}

// pushContext pushes a new context onto the stack and activates it.
func (gui *Gui) pushContext(ctx types.Context) {
	if cur := gui.currentContextObject(); cur != nil {
		cur.HandleFocusLost(types.OnFocusLostOpts{})
	}
	gui.contextMgr.Push(ctx.GetKey())
	gui.activateContext(ctx.GetKey())
	ctx.HandleFocus(types.OnFocusOpts{})
}

// popContext pops the top context and activates the one below it.
func (gui *Gui) popContext() {
	if cur := gui.currentContextObject(); cur != nil {
		cur.HandleFocusLost(types.OnFocusLostOpts{})
	}
	gui.contextMgr.Pop()
	gui.activateContext(gui.contextMgr.Current())
	if next := gui.currentContextObject(); next != nil {
		next.HandleFocus(types.OnFocusOpts{})
	}
}

// replaceContext replaces the top of the stack (e.g., search->preview).
func (gui *Gui) replaceContext(ctx types.Context) {
	if cur := gui.currentContextObject(); cur != nil {
		cur.HandleFocusLost(types.OnFocusLostOpts{})
	}
	gui.contextMgr.Replace(ctx.GetKey())
	gui.activateContext(ctx.GetKey())
	ctx.HandleFocus(types.OnFocusOpts{})
}

// pushContextByKey looks up the context by key and pushes it.
// Falls back to a direct stack push for lightweight contexts not in the tree.
func (gui *Gui) pushContextByKey(key types.ContextKey) {
	ctx := gui.contextMgr.ContextByKey(key)
	if ctx != nil {
		gui.pushContext(ctx)
		return
	}
	// Lightweight context (e.g., "searchFilter"): push key directly.
	gui.contextMgr.Push(key)
	gui.activateContext(key)
}

// replaceContextByKey looks up the context by key and replaces the top of the stack.
// Falls back to a direct stack replace for lightweight contexts not in the tree.
func (gui *Gui) replaceContextByKey(key types.ContextKey) {
	ctx := gui.contextMgr.ContextByKey(key)
	if ctx != nil {
		gui.replaceContext(ctx)
		return
	}
	// Lightweight context: replace key directly.
	gui.contextMgr.Replace(key)
	gui.activateContext(key)
}

// currentContextObject looks up the types.Context for the top of stack.
func (gui *Gui) currentContextObject() types.Context {
	return gui.contextMgr.ContextByKey(gui.contextMgr.Current())
}

// popupActive returns true when the current context is a popup (not a main panel).
func (gui *Gui) popupActive() bool {
	ctx := gui.contextMgr.ContextByKey(gui.contextMgr.Current())
	if ctx == nil {
		return false
	}
	kind := ctx.GetKind()
	return kind != types.SIDE_CONTEXT && kind != types.MAIN_CONTEXT
}

// overlayActive returns true when any overlay or dialog is open.
func (gui *Gui) overlayActive() bool {
	return gui.popupActive() ||
		(gui.state.Dialog != nil && gui.state.Dialog.Active)
}

func (gui *Gui) contextToView(ctx types.ContextKey) string {
	return gui.contexts.ViewNameForKey(ctx)
}
