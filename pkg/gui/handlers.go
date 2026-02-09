package gui

import (
	"os"
	"os/exec"
	"strings"

	"kvnd/lazyruin/pkg/commands"

	"github.com/jesseduffield/gocui"
)

// listMove adjusts *index by delta if the result is within [0, count).
// Returns true if the index was changed.
func listMove(index *int, count int, delta int) bool {
	next := *index + delta
	if next < 0 || next >= count {
		return false
	}
	*index = next
	return true
}

// Global handlers

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	order := []ContextKey{NotesContext, QueriesContext, TagsContext}

	// Include search filter in cycle if active
	if gui.state.SearchQuery != "" {
		order = []ContextKey{SearchFilterContext, NotesContext, QueriesContext, TagsContext}
	}

	for i, ctx := range order {
		if ctx == gui.state.CurrentContext {
			next := order[(i+1)%len(order)]
			gui.setContext(next)
			return nil
		}
	}

	gui.setContext(NotesContext)
	return nil
}

func (gui *Gui) prevPanel(g *gocui.Gui, v *gocui.View) error {
	order := []ContextKey{NotesContext, QueriesContext, TagsContext}

	// Include search filter in cycle if active
	if gui.state.SearchQuery != "" {
		order = []ContextKey{SearchFilterContext, NotesContext, QueriesContext, TagsContext}
	}

	for i, ctx := range order {
		if ctx == gui.state.CurrentContext {
			prev := order[(i-1+len(order))%len(order)]
			gui.setContext(prev)
			return nil
		}
	}

	gui.setContext(NotesContext)
	return nil
}

func (gui *Gui) focusNotes(g *gocui.Gui, v *gocui.View) error {
	if gui.state.CurrentContext == NotesContext {
		// Already focused - cycle through tabs
		gui.cycleNotesTab()
		return nil
	}
	gui.setContext(NotesContext)
	return nil
}

func (gui *Gui) focusQueries(g *gocui.Gui, v *gocui.View) error {
	if gui.state.CurrentContext == QueriesContext {
		gui.cycleQueriesTab()
		return nil
	}
	gui.setContext(QueriesContext)
	return nil
}

func (gui *Gui) focusTags(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(TagsContext)
	return nil
}

func (gui *Gui) focusPreview(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) focusSearchFilter(g *gocui.Gui, v *gocui.View) error {
	if gui.state.SearchQuery != "" {
		// Re-run search to restore results to Preview pane
		opts := gui.buildSearchOptions()
		notes, err := gui.ruinCmd.Search.Search(gui.state.SearchQuery, opts)
		if err == nil {
			gui.state.Preview.Mode = PreviewModeCardList
			gui.state.Preview.Cards = notes
			gui.state.Preview.SelectedCardIndex = 0
			if gui.views.Preview != nil {
				gui.views.Preview.Title = " Search: " + gui.state.SearchQuery + " "
			}
			gui.renderPreview()
		}
		gui.setContext(SearchFilterContext)
	}
	return nil
}

func (gui *Gui) openSearch(g *gocui.Gui, v *gocui.View) error {
	gui.state.SearchMode = true
	gui.setContext(SearchContext)
	return nil
}

func (gui *Gui) refresh(g *gocui.Gui, v *gocui.View) error {
	gui.refreshAll()
	return nil
}

// buildSearchOptions returns SearchOptions based on current preview toggle state
func (gui *Gui) buildSearchOptions() commands.SearchOptions {
	return commands.SearchOptions{
		IncludeContent:  true,
		StripGlobalTags: !gui.state.Preview.ShowGlobalTags,
		StripTitle:      !gui.state.Preview.ShowTitle,
	}
}

// Generic view scroll handlers (viewport only, no selection change)

func (gui *Gui) scrollViewDown(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	_, oy := v.Origin()
	v.SetOrigin(0, oy+1)
	return nil
}

func (gui *Gui) scrollViewUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	_, oy := v.Origin()
	if oy > 0 {
		v.SetOrigin(0, oy-1)
	}
	return nil
}

// Search handlers

func (gui *Gui) executeSearch(g *gocui.Gui, v *gocui.View) error {
	query := strings.TrimSpace(v.Buffer())
	if query == "" {
		return gui.cancelSearch(g, v)
	}

	opts := gui.buildSearchOptions()
	notes, err := gui.ruinCmd.Search.Search(query, opts)
	if err != nil {
		return nil
	}

	// Store search query for the search filter pane
	gui.state.SearchQuery = query
	gui.state.SearchMode = false

	// Display results in Preview pane (like tags)
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Search: " + query + " "
	}
	gui.renderPreview()

	gui.setContext(PreviewContext)

	return nil
}

func (gui *Gui) cancelSearch(g *gocui.Gui, v *gocui.View) error {
	gui.state.SearchMode = false
	gui.setContext(gui.state.PreviousContext)
	return nil
}

func (gui *Gui) clearSearch(g *gocui.Gui, v *gocui.View) error {
	gui.state.SearchQuery = ""
	gui.state.Notes.CurrentTab = NotesTabAll
	gui.loadNotesForCurrentTab()
	gui.setContext(NotesContext)
	return nil
}

// Help handler

func (gui *Gui) showHelpHandler(g *gocui.Gui, v *gocui.View) error {
	gui.showHelp()
	return nil
}

func (gui *Gui) openInEditor(path string) error {
	if err := gui.g.Suspend(); err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	parts := strings.Fields(editor)
	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	if err := gui.g.Resume(); err != nil {
		return err
	}

	gui.refreshAll()
	gui.renderAll()
	return nil
}

// Helper to check if a command is available
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
