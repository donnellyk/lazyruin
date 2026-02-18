package gui

import (
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"

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
	preview        *PreviewController
	stopBg         chan struct{}
	QuickCapture   bool // when true, open capture on start and quit on save
	darkBackground bool
}

// NewGui creates a new Gui instance.
func NewGui(cfg *config.Config, ruinCmd *commands.RuinCommand) *Gui {
	gui := &Gui{
		config:  cfg,
		ruinCmd: ruinCmd,
		views:   &Views{},
		state:   NewGuiState(),
	}
	gui.preview = NewPreviewController(gui)
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

// backgroundRefreshData reloads data without resetting focus, selection, or preview mode.
func (gui *Gui) backgroundRefreshData() {
	// Preserve selections
	cardIdx := gui.state.Preview.SelectedCardIndex

	gui.refreshNotes(true)
	gui.refreshTags(true)
	gui.refreshQueries(true)
	gui.refreshParents(true)

	if gui.state.Preview.SelectedCardIndex != cardIdx && cardIdx < len(gui.state.Preview.Cards) {
		gui.state.Preview.SelectedCardIndex = cardIdx
	}

	gui.renderPreview()
	gui.updateStatusBar()
}

// renderAll re-renders all views with current state (e.g. after resize)
func (gui *Gui) renderAll() {
	gui.renderNotes()
	gui.renderQueries()
	gui.renderTags()
	gui.renderPreview()
	gui.updateStatusBar()
}

func (gui *Gui) refreshAll() {
	gui.refreshNotes(false)
	gui.refreshTags(false)
	gui.refreshQueries(false)
	gui.refreshParents(false)
}

func (gui *Gui) refreshNotes(preserve bool) {
	gui.fetchNotesForCurrentTab(preserve)
}

func (gui *Gui) refreshTags(preserve bool) {
	idx := gui.state.Tags.SelectedIndex
	tags, err := gui.ruinCmd.Tags.List()
	if err != nil {
		return
	}
	gui.state.Tags.Items = tags
	if preserve && idx < len(tags) {
		gui.state.Tags.SelectedIndex = idx
	} else {
		gui.state.Tags.SelectedIndex = 0
	}
	gui.renderTags()
}

func (gui *Gui) refreshQueries(preserve bool) {
	idx := gui.state.Queries.SelectedIndex
	queries, err := gui.ruinCmd.Queries.List()
	if err != nil {
		return
	}
	gui.state.Queries.Items = queries
	if preserve && idx < len(queries) {
		gui.state.Queries.SelectedIndex = idx
	} else {
		gui.state.Queries.SelectedIndex = 0
	}
	gui.renderQueries()
}

// activateContext sets focus and refreshes data for the given context.
func (gui *Gui) activateContext(ctx ContextKey) {
	viewName := gui.contextToView(ctx)
	gui.g.SetCurrentView(viewName)

	// Re-render lists to update highlight visibility
	gui.renderNotes()
	gui.renderQueries()
	gui.renderTags()

	// Refresh data (preserving selections) and update preview based on new context
	switch ctx {
	case NotesContext:
		gui.refreshNotes(true)
		gui.preview.updatePreviewForNotes()
	case QueriesContext:
		if gui.state.Queries.CurrentTab == QueriesTabParents {
			gui.refreshParents(true)
			gui.updatePreviewForParents()
		} else {
			gui.refreshQueries(true)
			gui.updatePreviewForQueries()
		}
	case TagsContext:
		gui.refreshTags(true)
		gui.updatePreviewForTags()
	case PreviewContext:
		gui.renderPreview()
	}

	gui.updateStatusBar()
}

// pushContext pushes a new context onto the stack and activates it.
func (gui *Gui) pushContext(ctx ContextKey) {
	gui.state.ContextStack = append(gui.state.ContextStack, ctx)
	gui.activateContext(ctx)
}

// popContext pops the top context and activates the one below it.
func (gui *Gui) popContext() {
	if len(gui.state.ContextStack) > 1 {
		gui.state.ContextStack = gui.state.ContextStack[:len(gui.state.ContextStack)-1]
	}
	gui.activateContext(gui.state.currentContext())
}

// replaceContext replaces the top of the stack (e.g., search→preview).
func (gui *Gui) replaceContext(ctx ContextKey) {
	if len(gui.state.ContextStack) > 0 {
		gui.state.ContextStack[len(gui.state.ContextStack)-1] = ctx
	} else {
		gui.state.ContextStack = []ContextKey{ctx}
	}
	gui.activateContext(ctx)
}

// setContext is a convenience that pushes a new context (legacy compatibility).
func (gui *Gui) setContext(ctx ContextKey) {
	gui.pushContext(ctx)
}

// openOverlay opens a modal overlay. Returns false if one is already active.
func (gui *Gui) openOverlay(overlay OverlayType) bool {
	if gui.state.ActiveOverlay != OverlayNone {
		return false
	}
	gui.state.ActiveOverlay = overlay
	return true
}

// closeOverlay closes the current modal overlay.
func (gui *Gui) closeOverlay() {
	gui.state.ActiveOverlay = OverlayNone
}

// overlayActive returns true when any overlay or dialog is open.
func (gui *Gui) overlayActive() bool {
	return gui.state.ActiveOverlay != OverlayNone ||
		(gui.state.Dialog != nil && gui.state.Dialog.Active)
}

func (gui *Gui) contextToView(ctx ContextKey) string {
	switch ctx {
	case NotesContext:
		return NotesView
	case QueriesContext:
		return QueriesView
	case TagsContext:
		return TagsView
	case PreviewContext:
		return PreviewView
	case SearchContext:
		return SearchView
	case SearchFilterContext:
		return SearchFilterView
	case CaptureContext:
		return CaptureView
	case PickContext:
		return PickView
	case PaletteContext:
		return PaletteView
	}
	return NotesView
}
