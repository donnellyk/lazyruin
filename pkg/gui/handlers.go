package gui

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

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
		if gui.state.SearchQuery != "" {
			// If search is active, pressing 1 reloads current tab (clears search)
			gui.loadNotesForCurrentTab()
		} else {
			// Already focused, no search - cycle through tabs
			gui.cycleNotesTab()
		}
		return nil
	}
	gui.setContext(NotesContext)
	return nil
}

// cycleNotesTab cycles through All -> Today -> Recent tabs
func (gui *Gui) cycleNotesTab() {
	tabs := []NotesTab{NotesTabAll, NotesTabToday, NotesTabRecent}
	for i, tab := range tabs {
		if tab == gui.state.Notes.CurrentTab {
			gui.state.Notes.CurrentTab = tabs[(i+1)%len(tabs)]
			break
		}
	}
	gui.loadNotesForCurrentTab()
}

// loadNotesForCurrentTab loads notes based on the current tab
// This clears any active search since tabs always show their full results
func (gui *Gui) loadNotesForCurrentTab() {
	// Clear search when switching tabs - tabs always show full results
	gui.state.SearchQuery = ""

	var notes []models.Note
	var err error

	switch gui.state.Notes.CurrentTab {
	case NotesTabAll:
		notes, err = gui.ruinCmd.Search.All(50)
	case NotesTabToday:
		notes, err = gui.ruinCmd.Search.Today()
	case NotesTabRecent:
		notes, err = gui.ruinCmd.Search.Recent(20)
	}

	if err == nil {
		gui.state.Notes.Items = notes
		gui.state.Notes.SelectedIndex = 0
	}
	gui.renderNotes()
	gui.updateNotesTitle()
	gui.updatePreviewForNotes()
}

func (gui *Gui) focusQueries(g *gocui.Gui, v *gocui.View) error {
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

// Notes handlers

func (gui *Gui) notesDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Notes.SelectedIndex < len(gui.state.Notes.Items)-1 {
		gui.state.Notes.SelectedIndex++
		gui.renderNotes()
		gui.updatePreviewForNotes()
	}
	return nil
}

func (gui *Gui) notesUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Notes.SelectedIndex > 0 {
		gui.state.Notes.SelectedIndex--
		gui.renderNotes()
		gui.updatePreviewForNotes()
	}
	return nil
}

func (gui *Gui) notesTop(g *gocui.Gui, v *gocui.View) error {
	gui.state.Notes.SelectedIndex = 0
	gui.renderNotes()
	gui.updatePreviewForNotes()
	return nil
}

func (gui *Gui) notesBottom(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) > 0 {
		gui.state.Notes.SelectedIndex = len(gui.state.Notes.Items) - 1
		gui.renderNotes()
		gui.updatePreviewForNotes()
	}
	return nil
}

// ErrEditFile signals that we need to edit a file (exit main loop, run editor, restart)
var ErrEditFile = errors.New("edit file")

func (gui *Gui) editNote(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}

	note := gui.state.Notes.Items[gui.state.Notes.SelectedIndex]
	gui.state.EditFilePath = note.Path
	return ErrEditFile
}

func (gui *Gui) runEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Queries handlers

func (gui *Gui) queriesDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.SelectedIndex < len(gui.state.Queries.Items)-1 {
		gui.state.Queries.SelectedIndex++
		gui.renderQueries()
		gui.updatePreviewForQueries()
	}
	return nil
}

func (gui *Gui) queriesUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.SelectedIndex > 0 {
		gui.state.Queries.SelectedIndex--
		gui.renderQueries()
		gui.updatePreviewForQueries()
	}
	return nil
}

func (gui *Gui) runQuery(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Queries.Items) == 0 {
		return nil
	}

	query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
	notes, err := gui.ruinCmd.Queries.Run(query.Name)
	if err != nil {
		return nil
	}

	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Preview: " + query.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)

	return nil
}

// Tags handlers

func (gui *Gui) tagsDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Tags.SelectedIndex < len(gui.state.Tags.Items)-1 {
		gui.state.Tags.SelectedIndex++
		gui.renderTags()
		gui.updatePreviewForTags()
	}
	return nil
}

func (gui *Gui) tagsUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Tags.SelectedIndex > 0 {
		gui.state.Tags.SelectedIndex--
		gui.renderTags()
		gui.updatePreviewForTags()
	}
	return nil
}

func (gui *Gui) filterByTag(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Tags.Items) == 0 {
		return nil
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]
	notes, err := gui.ruinCmd.Search.ByTag(tag.Name)
	if err != nil {
		return nil
	}

	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Preview: #" + tag.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)

	return nil
}

// Preview handlers

func (gui *Gui) previewDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Preview.Mode == PreviewModeCardList {
		if gui.state.Preview.SelectedCardIndex < len(gui.state.Preview.Cards)-1 {
			gui.state.Preview.SelectedCardIndex++
			gui.renderPreview()
		}
	} else {
		gui.state.Preview.ScrollOffset++
		gui.renderPreview()
	}
	return nil
}

func (gui *Gui) previewUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Preview.Mode == PreviewModeCardList {
		if gui.state.Preview.SelectedCardIndex > 0 {
			gui.state.Preview.SelectedCardIndex--
			gui.renderPreview()
		}
	} else {
		if gui.state.Preview.ScrollOffset > 0 {
			gui.state.Preview.ScrollOffset--
			gui.renderPreview()
		}
	}
	return nil
}

func (gui *Gui) previewBack(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(gui.state.PreviousContext)
	return nil
}

func (gui *Gui) toggleFrontmatter(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowFrontmatter = !gui.state.Preview.ShowFrontmatter
	gui.renderPreview()
	return nil
}

func (gui *Gui) toggleTitle(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowTitle = !gui.state.Preview.ShowTitle
	gui.renderPreview()
	return nil
}

func (gui *Gui) toggleGlobalTags(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowGlobalTags = !gui.state.Preview.ShowGlobalTags
	gui.renderPreview()
	return nil
}

func (gui *Gui) focusNoteFromPreview(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Preview.Mode != PreviewModeCardList {
		return nil
	}

	if len(gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := gui.state.Preview.Cards[gui.state.Preview.SelectedCardIndex]

	for i, note := range gui.state.Notes.Items {
		if note.UUID == card.UUID {
			gui.state.Notes.SelectedIndex = i
			gui.setContext(NotesContext)
			gui.renderNotes()
			return nil
		}
	}

	return nil
}

// Search handlers

func (gui *Gui) executeSearch(g *gocui.Gui, v *gocui.View) error {
	query := strings.TrimSpace(v.Buffer())
	if query == "" {
		return gui.cancelSearch(g, v)
	}

	notes, err := gui.ruinCmd.Search.Search(query, commands.SearchOptions{})
	if err != nil {
		return nil
	}

	gui.state.Notes.Items = notes
	gui.state.Notes.SelectedIndex = 0
	gui.state.SearchQuery = query
	gui.state.SearchMode = false

	gui.setContext(NotesContext)
	gui.renderNotes()
	gui.updatePreviewForNotes()

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

// Preview update helpers

func (gui *Gui) updatePreviewForNotes() {
	gui.state.Preview.Mode = PreviewModeSingleNote
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Preview "
		gui.renderPreview()
	}
}

func (gui *Gui) updatePreviewForTags() {
	if len(gui.state.Tags.Items) == 0 {
		return
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]
	notes, _ := gui.ruinCmd.Search.ByTag(tag.Name)

	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Preview: #" + tag.Name + " "
		gui.renderPreview()
	}
}

func (gui *Gui) updatePreviewForQueries() {
	if len(gui.state.Queries.Items) == 0 {
		return
	}

	query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
	notes, _ := gui.ruinCmd.Queries.Run(query.Name)

	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Preview: " + query.Name + " "
		gui.renderPreview()
	}
}

// Help handler

func (gui *Gui) showHelpHandler(g *gocui.Gui, v *gocui.View) error {
	gui.showHelp()
	return nil
}

// Note action handlers

func (gui *Gui) newNote(g *gocui.Gui, v *gocui.View) error {
	gui.showInput("New Note", "Enter note content:", func(content string) error {
		if content == "" {
			return nil
		}
		_, err := gui.ruinCmd.Execute("log", content)
		if err != nil {
			return nil
		}
		gui.refreshNotes()
		return nil
	})
	return nil
}

func (gui *Gui) deleteNote(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}

	note := gui.state.Notes.Items[gui.state.Notes.SelectedIndex]
	title := note.Title
	if title == "" {
		title = note.Path
	}
	if len(title) > 30 {
		title = title[:30] + "..."
	}

	gui.showConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
		err := os.Remove(note.Path)
		if err != nil {
			return nil
		}
		gui.refreshNotes()
		return nil
	})
	return nil
}

func (gui *Gui) copyNotePath(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}

	note := gui.state.Notes.Items[gui.state.Notes.SelectedIndex]

	// Use pbcopy on macOS, xclip on Linux
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("pbcopy"):
		cmd = exec.Command("pbcopy")
	case isCommandAvailable("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case isCommandAvailable("xsel"):
		cmd = exec.Command("xsel", "--clipboard", "--input")
	default:
		return nil
	}

	cmd.Stdin = strings.NewReader(note.Path)
	cmd.Run()

	return nil
}

// Tag action handlers

func (gui *Gui) renameTag(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Tags.Items) == 0 {
		return nil
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]

	gui.showInput("Rename Tag", "New name for #"+tag.Name+":", func(newName string) error {
		if newName == "" || newName == tag.Name {
			return nil
		}
		err := gui.ruinCmd.Tags.Rename(tag.Name, newName)
		if err != nil {
			return nil
		}
		gui.refreshTags()
		gui.refreshNotes()
		return nil
	})
	return nil
}

func (gui *Gui) deleteTag(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Tags.Items) == 0 {
		return nil
	}

	tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]

	gui.showConfirm("Delete Tag", "Delete #"+tag.Name+" from all notes?", func() error {
		err := gui.ruinCmd.Tags.Delete(tag.Name)
		if err != nil {
			return nil
		}
		gui.refreshTags()
		gui.refreshNotes()
		return nil
	})
	return nil
}

// Query action handlers

func (gui *Gui) deleteQuery(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Queries.Items) == 0 {
		return nil
	}

	query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]

	gui.showConfirm("Delete Query", "Delete query \""+query.Name+"\"?", func() error {
		err := gui.ruinCmd.Queries.Delete(query.Name)
		if err != nil {
			return nil
		}
		gui.refreshQueries()
		return nil
	})
	return nil
}

// Helper to check if a command is available
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
