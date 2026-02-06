package gui

import (
	"time"

	"kvnd/lazyruin/pkg/commands"

	"github.com/jesseduffield/gocui"
)

// Gui manages the terminal user interface.
type Gui struct {
	g       *gocui.Gui
	views   *Views
	state   *GuiState
	ruinCmd *commands.RuinCommand
	stopBg  chan struct{}
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
	noteIdx := gui.state.Notes.SelectedIndex
	tagIdx := gui.state.Tags.SelectedIndex
	queryIdx := gui.state.Queries.SelectedIndex
	cardIdx := gui.state.Preview.SelectedCardIndex

	gui.loadNotesForCurrentTabPreserve()
	if noteIdx < len(gui.state.Notes.Items) {
		gui.state.Notes.SelectedIndex = noteIdx
	}

	if tags, err := gui.ruinCmd.Tags.List(); err == nil {
		gui.state.Tags.Items = tags
		if tagIdx < len(tags) {
			gui.state.Tags.SelectedIndex = tagIdx
		}
	}

	if queries, err := gui.ruinCmd.Queries.List(); err == nil {
		gui.state.Queries.Items = queries
		if queryIdx < len(queries) {
			gui.state.Queries.SelectedIndex = queryIdx
		}
	}

	if gui.state.Preview.SelectedCardIndex != cardIdx && cardIdx < len(gui.state.Preview.Cards) {
		gui.state.Preview.SelectedCardIndex = cardIdx
	}

	gui.renderNotes()
	gui.renderTags()
	gui.renderQueries()
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
	gui.refreshNotes()
	gui.refreshTags()
	gui.refreshQueries()
}

func (gui *Gui) refreshNotes() {
	gui.loadNotesForCurrentTab()
}

func (gui *Gui) refreshNotesPreserve() {
	idx := gui.state.Notes.SelectedIndex
	gui.loadNotesForCurrentTabPreserve()
	if idx < len(gui.state.Notes.Items) {
		gui.state.Notes.SelectedIndex = idx
	}
	gui.renderNotes()
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

func (gui *Gui) refreshTagsPreserve() {
	idx := gui.state.Tags.SelectedIndex
	tags, err := gui.ruinCmd.Tags.List()
	if err != nil {
		return
	}
	gui.state.Tags.Items = tags
	if idx < len(tags) {
		gui.state.Tags.SelectedIndex = idx
	}
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

func (gui *Gui) refreshQueriesPreserve() {
	idx := gui.state.Queries.SelectedIndex
	queries, err := gui.ruinCmd.Queries.List()
	if err != nil {
		return
	}
	gui.state.Queries.Items = queries
	if idx < len(queries) {
		gui.state.Queries.SelectedIndex = idx
	}
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

	// Refresh data (preserving selections) and update preview based on new context
	switch ctx {
	case NotesContext:
		gui.refreshNotesPreserve()
		gui.updatePreviewForNotes()
	case QueriesContext:
		gui.refreshQueriesPreserve()
		gui.updatePreviewForQueries()
	case TagsContext:
		gui.refreshTagsPreserve()
		gui.updatePreviewForTags()
	case PreviewContext:
		gui.renderPreview()
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
	case SearchFilterContext:
		return SearchFilterView
	}
	return NotesView
}
