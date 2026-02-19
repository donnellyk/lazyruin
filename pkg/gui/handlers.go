package gui

import (
	"os"
	"os/exec"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"

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
	return gui.globalController.NextPanel()
}

func (gui *Gui) prevPanel(g *gocui.Gui, v *gocui.View) error {
	return gui.globalController.PrevPanel()
}

func (gui *Gui) focusNotes(g *gocui.Gui, v *gocui.View) error {
	return gui.globalController.FocusNotes()
}

func (gui *Gui) focusQueries(g *gocui.Gui, v *gocui.View) error {
	return gui.globalController.FocusQueries()
}

func (gui *Gui) focusTags(g *gocui.Gui, v *gocui.View) error {
	return gui.globalController.FocusTags()
}

func (gui *Gui) focusPreview(g *gocui.Gui, v *gocui.View) error {
	return gui.globalController.FocusPreview()
}

func (gui *Gui) focusSearchFilter(g *gocui.Gui, v *gocui.View) error {
	if gui.state.SearchQuery != "" {
		// Re-run search to restore results to Preview pane
		query, sort := extractSort(gui.state.SearchQuery)
		opts := gui.buildSearchOptions()
		opts.Sort = sort
		notes, err := gui.ruinCmd.Search.Search(query, opts)
		if err == nil {
			gui.helpers.Preview().ShowCardList(" Search: "+gui.state.SearchQuery+" ", notes)
		}
		gui.setContext(SearchFilterContext)
	}
	return nil
}

func (gui *Gui) openSearch(g *gocui.Gui, v *gocui.View) error {
	if gui.state.popupActive() {
		return nil
	}
	cs := NewCompletionState()
	cs.FallbackCandidates = ambientDateCandidates
	gui.state.SearchCompletion = cs
	gui.pushContext(SearchContext)
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
		StripGlobalTags: !gui.contexts.Preview.ShowGlobalTags,
		StripTitle:      !gui.contexts.Preview.ShowTitle,
	}
}

func (gui *Gui) executeSearch(g *gocui.Gui, v *gocui.View) error {
	raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	if raw == "" {
		return gui.cancelSearch(g, v)
	}

	query, sort := extractSort(raw)
	opts := gui.buildSearchOptions()
	opts.Sort = sort
	notes, err := gui.ruinCmd.Search.Search(query, opts)
	if err != nil {
		gui.showError(err)
		return nil
	}

	// Store full input for the search filter pane display
	gui.state.SearchQuery = raw
	gui.state.SearchCompletion = NewCompletionState()
	g.Cursor = false

	// Display results in Preview pane
	gui.helpers.Preview().PushNavHistory()
	gui.helpers.Preview().ShowCardList(" Search: "+query+" ", notes)

	gui.replaceContext(PreviewContext)

	return nil
}

func (gui *Gui) cancelSearch(g *gocui.Gui, v *gocui.View) error {
	gui.state.SearchCompletion = NewCompletionState()
	g.Cursor = false
	gui.popContext()
	return nil
}

func (gui *Gui) clearSearch(g *gocui.Gui, v *gocui.View) error {
	gui.state.SearchQuery = ""
	gui.contexts.Notes.CurrentTab = context.NotesTabAll
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

	gui.refreshTags(false)
	gui.refreshQueries(false)
	gui.refreshParents(false)
	gui.helpers.Preview().ReloadContent()
	gui.renderAll()
	return nil
}
