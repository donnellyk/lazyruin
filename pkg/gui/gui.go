package gui

import (
	"kvnd/lazyruin/pkg/commands"

	"github.com/jesseduffield/gocui"
)

// Gui manages the terminal user interface.
type Gui struct {
	g       *gocui.Gui
	views   *Views
	state   *GuiState
	ruinCmd *commands.RuinCommand
}

// NewGui creates a new Gui instance.
func NewGui(ruinCmd *commands.RuinCommand) *Gui {
	return &Gui{
		ruinCmd: ruinCmd,
		views:   &Views{},
		state:   NewGuiState(),
	}
}

// Run starts the GUI event loop.
func (gui *Gui) Run() error {
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
	})
	if err != nil {
		return err
	}
	defer g.Close()

	gui.g = g
	g.Mouse = true
	g.Cursor = false
	g.SetManager(gocui.ManagerFunc(gui.layout))

	if err := gui.setupKeybindings(); err != nil {
		return err
	}

	gui.refreshAll()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}

func (gui *Gui) refreshAll() {
	gui.refreshNotes()
	gui.refreshTags()
	gui.refreshQueries()
}

func (gui *Gui) refreshNotes() {
	gui.loadNotesForCurrentTab()
}

func (gui *Gui) refreshTags() {
	tags, err := gui.ruinCmd.Tags.List()
	if err != nil {
		return
	}

	gui.state.Tags.Items = tags
	gui.state.Tags.SelectedIndex = 0
	gui.renderTags()
}

func (gui *Gui) refreshQueries() {
	queries, err := gui.ruinCmd.Queries.List()
	if err != nil {
		return
	}

	gui.state.Queries.Items = queries
	gui.state.Queries.SelectedIndex = 0
	gui.renderQueries()
}

func (gui *Gui) setContext(ctx ContextKey) {
	gui.state.PreviousContext = gui.state.CurrentContext
	gui.state.CurrentContext = ctx

	viewName := gui.contextToView(ctx)
	gui.g.SetCurrentView(viewName)

	// Re-render lists to update highlight visibility
	gui.renderNotes()
	gui.renderQueries()
	gui.renderTags()

	// Update preview based on new context
	switch ctx {
	case NotesContext:
		gui.updatePreviewForNotes()
	case QueriesContext:
		gui.updatePreviewForQueries()
	case TagsContext:
		gui.updatePreviewForTags()
	}

	gui.updateStatusBar()
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
	}
	return NotesView
}
