package gui

import (
	"fmt"

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

	g.SetCurrentView(NotesView)

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
	notes, err := gui.ruinCmd.Search.Today() // TODO: Fix this!
	if err != nil {
		// Show error in status - TODO?
		return
	}

	gui.state.Notes.Items = notes
	gui.state.Notes.SelectedIndex = 0
	// gui.renderNotes() // TODO:
}

func (gui *Gui) refreshTags() {
	tags, err := gui.ruinCmd.Tags.List()
	if err != nil {
		return
	}

	gui.state.Tags.Items = tags
	gui.state.Tags.SelectedIndex = 0
	// gui.renderTags() // TODO:
}

func (gui *Gui) refreshQueries() {
	queries, err := gui.ruinCmd.Queries.List()
	if err != nil {
		return
	}

	gui.state.Queries.Items = queries
	gui.state.Queries.SelectedIndex = 0
	// gui.renderQueries() // TODO:
}

func (gui *Gui) setContext(ctx ContextKey) {
	gui.state.PreviousContext = gui.state.CurrentContext
	gui.state.CurrentContext = ctx

	viewName := gui.contextToView(ctx)
	gui.g.SetCurrentView(viewName)

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

// layout is called on every render to set up views.
func (gui *Gui) oldLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// For now, just create a simple main view
	if v, err := g.SetView("main", 0, 0, maxX-1, maxY-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " LazyRuin "
		v.Wrap = true
		fmt.Fprintf(v, "Welcome to LazyRuin!\n\n")
		fmt.Fprintf(v, "Vault: %s\n\n", gui.ruinCmd.VaultPath())
		fmt.Fprintf(v, "Press 'q' to quit.\n")
	}

	return nil
}

// setupKeybindings configures keyboard shortcuts.
func (gui *Gui) setupKeybindings() error {
	// Quit
	if err := gui.g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
