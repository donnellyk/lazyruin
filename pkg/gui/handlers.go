package gui

import (
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
		// Already focused - cycle through tabs
		gui.cycleNotesTab()
		return nil
	}
	gui.setContext(NotesContext)
	return nil
}

func (gui *Gui) notesClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	idx := (cy + oy) / 2 // 2 lines per note
	if idx >= 0 && idx < len(gui.state.Notes.Items) {
		gui.state.Notes.SelectedIndex = idx
	}
	gui.setContext(NotesContext)
	return nil
}

// cycleNotesTab cycles through All -> Today -> Recent tabs
func (gui *Gui) cycleNotesTab() {
	idx := (gui.notesTabIndex() + 1) % len(notesTabs)
	gui.state.Notes.CurrentTab = notesTabs[idx]
	gui.loadNotesForCurrentTab()
}

// switchNotesTabByIndex switches to a specific tab by index (for tab click)
func (gui *Gui) switchNotesTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(notesTabs) {
		return nil
	}
	gui.state.Notes.CurrentTab = notesTabs[tabIndex]
	gui.loadNotesForCurrentTab()
	gui.setContext(NotesContext)
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

// loadNotesForCurrentTab loads notes based on the current tab
func (gui *Gui) loadNotesForCurrentTab() {
	var notes []models.Note
	var err error

	opts := gui.buildSearchOptions()
	opts.Sort = "created:desc"

	switch gui.state.Notes.CurrentTab {
	case NotesTabAll:
		opts.Limit = 50
		notes, err = gui.ruinCmd.Search.Search("created:10000d", opts)
	case NotesTabToday:
		notes, err = gui.ruinCmd.Search.Today()
	case NotesTabRecent:
		opts.Limit = 20
		notes, err = gui.ruinCmd.Search.Search("created:7d", opts)
	}

	if err == nil {
		gui.state.Notes.Items = notes
		gui.state.Notes.SelectedIndex = 0
	}
	gui.renderNotes()
	gui.updateNotesTab()
	gui.updatePreviewForNotes()
}

func (gui *Gui) focusQueries(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(QueriesContext)
	return nil
}

func (gui *Gui) queriesClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	idx := (cy + oy) / 2 // 2 lines per query
	if idx >= 0 && idx < len(gui.state.Queries.Items) {
		gui.state.Queries.SelectedIndex = idx
	}
	gui.setContext(QueriesContext)
	return nil
}

func (gui *Gui) focusTags(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(TagsContext)
	return nil
}

func (gui *Gui) tagsClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	idx := cy + oy // 1 line per tag
	if idx >= 0 && idx < len(gui.state.Tags.Items) {
		gui.state.Tags.SelectedIndex = idx
	}
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

func (gui *Gui) editNote(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}

	note := gui.state.Notes.Items[gui.state.Notes.SelectedIndex]
	return gui.openInEditor(note.Path)
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
	opts := gui.buildSearchOptions()
	notes, err := gui.ruinCmd.Search.Search(tag.Name, opts)
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

func (gui *Gui) previewScrollDown(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() != PreviewView {
		return nil
	}
	gui.state.Preview.ScrollOffset += 3
	v.SetOrigin(0, gui.state.Preview.ScrollOffset)
	return nil
}

func (gui *Gui) previewScrollUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() != PreviewView {
		return nil
	}
	gui.state.Preview.ScrollOffset -= 3
	if gui.state.Preview.ScrollOffset < 0 {
		gui.state.Preview.ScrollOffset = 0
	}
	v.SetOrigin(0, gui.state.Preview.ScrollOffset)
	return nil
}

func (gui *Gui) previewClick(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Preview.Mode != PreviewModeCardList {
		gui.setContext(PreviewContext)
		return nil
	}

	_, cy := v.Cursor()
	_, oy := v.Origin()
	absY := cy + oy

	for i, lr := range gui.state.Preview.CardLineRanges {
		if absY >= lr[0] && absY < lr[1] {
			gui.state.Preview.SelectedCardIndex = i
			gui.setContext(PreviewContext)
			return nil
		}
	}

	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) previewBack(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Preview.EditMode {
		gui.state.Preview.EditMode = false
		gui.refreshNotes()
	}
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
	gui.reloadContent()
	return nil
}

func (gui *Gui) toggleGlobalTags(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.ShowGlobalTags = !gui.state.Preview.ShowGlobalTags
	gui.reloadContent()
	return nil
}

// reloadContent reloads notes from CLI with current toggle settings,
// preserving selection indices and preview mode.
func (gui *Gui) reloadContent() {
	// Reload notes for the Notes pane, preserving selection
	savedNoteIdx := gui.state.Notes.SelectedIndex
	gui.loadNotesForCurrentTabPreserve()
	if savedNoteIdx < len(gui.state.Notes.Items) {
		gui.state.Notes.SelectedIndex = savedNoteIdx
	}
	gui.renderNotes()

	// Reload cards in Preview pane if in card list mode
	if gui.state.Preview.Mode == PreviewModeCardList && len(gui.state.Preview.Cards) > 0 {
		savedCardIdx := gui.state.Preview.SelectedCardIndex
		gui.reloadPreviewCards()
		if savedCardIdx < len(gui.state.Preview.Cards) {
			gui.state.Preview.SelectedCardIndex = savedCardIdx
		}
		gui.renderPreview()
	} else {
		gui.renderPreview()
	}
}

// loadNotesForCurrentTabPreserve reloads notes without resetting selection or touching preview.
func (gui *Gui) loadNotesForCurrentTabPreserve() {
	var notes []models.Note
	var err error

	opts := gui.buildSearchOptions()
	opts.Sort = "created:desc"

	switch gui.state.Notes.CurrentTab {
	case NotesTabAll:
		opts.Limit = 50
		notes, err = gui.ruinCmd.Search.Search("created:10000d", opts)
	case NotesTabToday:
		notes, err = gui.ruinCmd.Search.Today()
	case NotesTabRecent:
		opts.Limit = 20
		notes, err = gui.ruinCmd.Search.Search("created:7d", opts)
	}

	if err == nil {
		gui.state.Notes.Items = notes
	}
}

// reloadPreviewCards reloads the preview cards based on what generated them
func (gui *Gui) reloadPreviewCards() {
	opts := gui.buildSearchOptions()

	// If there's an active search query, reload search results
	if gui.state.SearchQuery != "" {
		notes, err := gui.ruinCmd.Search.Search(gui.state.SearchQuery, opts)
		if err == nil {
			gui.state.Preview.Cards = notes
		}
		gui.renderPreview()
		return
	}

	// Otherwise, reload based on previous context
	switch gui.state.PreviousContext {
	case TagsContext:
		if len(gui.state.Tags.Items) > 0 {
			tag := gui.state.Tags.Items[gui.state.Tags.SelectedIndex]
			notes, err := gui.ruinCmd.Search.Search(tag.Name, opts)
			if err == nil {
				gui.state.Preview.Cards = notes
			}
		}
	case QueriesContext:
		if len(gui.state.Queries.Items) > 0 {
			query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
			// Query run doesn't support strip options, so we re-search
			notes, err := gui.ruinCmd.Queries.Run(query.Name)
			if err == nil {
				gui.state.Preview.Cards = notes
			}
		}
	}

	gui.renderPreview()
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
	opts := gui.buildSearchOptions()
	notes, _ := gui.ruinCmd.Search.Search(tag.Name, opts)

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

// Edit mode handlers

func (gui *Gui) editNotesInPreview(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}

	gui.state.Preview.Cards = make([]models.Note, len(gui.state.Notes.Items))
	copy(gui.state.Preview.Cards, gui.state.Notes.Items)
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.EditMode = true
	gui.state.Preview.SelectedCardIndex = gui.state.Notes.SelectedIndex
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Edit Mode "
	}
	gui.renderPreview()
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) deleteCardFromPreview(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	if len(gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := gui.state.Preview.Cards[gui.state.Preview.SelectedCardIndex]
	title := card.Title
	if title == "" {
		title = card.Path
	}
	if len(title) > 30 {
		title = title[:30] + "..."
	}

	gui.showConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
		err := os.Remove(card.Path)
		if err != nil {
			return nil
		}
		idx := gui.state.Preview.SelectedCardIndex
		gui.state.Preview.Cards = append(gui.state.Preview.Cards[:idx], gui.state.Preview.Cards[idx+1:]...)
		if gui.state.Preview.SelectedCardIndex >= len(gui.state.Preview.Cards) && gui.state.Preview.SelectedCardIndex > 0 {
			gui.state.Preview.SelectedCardIndex--
		}
		gui.refreshNotes()
		gui.renderPreview()
		return nil
	})
	return nil
}

func (gui *Gui) moveCardUp(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	idx := gui.state.Preview.SelectedCardIndex
	if idx <= 0 {
		return nil
	}
	gui.state.Preview.Cards[idx], gui.state.Preview.Cards[idx-1] = gui.state.Preview.Cards[idx-1], gui.state.Preview.Cards[idx]
	gui.state.Preview.SelectedCardIndex--
	gui.renderPreview()
	return nil
}

func (gui *Gui) moveCardDown(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	idx := gui.state.Preview.SelectedCardIndex
	if idx >= len(gui.state.Preview.Cards)-1 {
		return nil
	}
	gui.state.Preview.Cards[idx], gui.state.Preview.Cards[idx+1] = gui.state.Preview.Cards[idx+1], gui.state.Preview.Cards[idx]
	gui.state.Preview.SelectedCardIndex++
	gui.renderPreview()
	return nil
}

func (gui *Gui) mergeCardHandler(g *gocui.Gui, v *gocui.View) error {
	if !gui.state.Preview.EditMode {
		return nil
	}
	if len(gui.state.Preview.Cards) <= 1 {
		return nil
	}
	gui.showMergeOverlay()
	return nil
}

func (gui *Gui) executeMerge(direction string) error {
	idx := gui.state.Preview.SelectedCardIndex
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(gui.state.Preview.Cards)-1 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx + 1
	} else {
		if idx <= 0 {
			return nil
		}
		targetIdx = idx
		sourceIdx = idx - 1
	}

	target := gui.state.Preview.Cards[targetIdx]
	source := gui.state.Preview.Cards[sourceIdx]

	// Read both files' raw content (after stripping frontmatter)
	targetContent, err := gui.loadNoteContent(target.Path)
	if err != nil {
		return nil
	}
	sourceContent, err := gui.loadNoteContent(source.Path)
	if err != nil {
		return nil
	}

	// Merge tags (union)
	tagSet := make(map[string]bool)
	for _, t := range target.Tags {
		tagSet[t] = true
	}
	for _, t := range source.Tags {
		tagSet[t] = true
	}
	var mergedTags []string
	for t := range tagSet {
		mergedTags = append(mergedTags, t)
	}

	// Combine content
	combined := strings.TrimRight(targetContent, "\n") + "\n\n" + strings.TrimRight(sourceContent, "\n") + "\n"

	// Rewrite target file
	err = gui.writeNoteFile(target.Path, combined, mergedTags)
	if err != nil {
		return nil
	}

	// Delete source file
	os.Remove(source.Path)

	// Remove source from cards
	gui.state.Preview.Cards = append(gui.state.Preview.Cards[:sourceIdx], gui.state.Preview.Cards[sourceIdx+1:]...)
	if gui.state.Preview.SelectedCardIndex >= len(gui.state.Preview.Cards) {
		gui.state.Preview.SelectedCardIndex = len(gui.state.Preview.Cards) - 1
	}
	if gui.state.Preview.SelectedCardIndex < 0 {
		gui.state.Preview.SelectedCardIndex = 0
	}

	gui.refreshNotes()
	gui.renderPreview()
	return nil
}

// writeNoteFile rewrites a note file preserving uuid/created/updated, with merged tags and new content.
func (gui *Gui) writeNoteFile(path, content string, tags []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Extract existing frontmatter fields
	raw := string(data)
	uuid := ""
	created := ""
	updated := ""
	title := ""

	if strings.HasPrefix(raw, "---") {
		rest := raw[3:]
		if idx := strings.Index(rest, "\n---"); idx != -1 {
			fmBlock := rest[:idx]
			for _, line := range strings.Split(fmBlock, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "uuid:") {
					uuid = strings.TrimSpace(strings.TrimPrefix(line, "uuid:"))
				} else if strings.HasPrefix(line, "created:") {
					created = strings.TrimSpace(strings.TrimPrefix(line, "created:"))
				} else if strings.HasPrefix(line, "updated:") {
					updated = strings.TrimSpace(strings.TrimPrefix(line, "updated:"))
				} else if strings.HasPrefix(line, "title:") {
					title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				}
			}
		}
	}

	// Build new frontmatter
	var fm strings.Builder
	fm.WriteString("---\n")
	if uuid != "" {
		fm.WriteString("uuid: " + uuid + "\n")
	}
	if created != "" {
		fm.WriteString("created: " + created + "\n")
	}
	if updated != "" {
		fm.WriteString("updated: " + updated + "\n")
	}
	if title != "" {
		fm.WriteString("title: " + title + "\n")
	}
	if len(tags) > 0 {
		fm.WriteString("tags:\n")
		for _, t := range tags {
			fm.WriteString("  - " + t + "\n")
		}
	} else {
		fm.WriteString("tags: []\n")
	}
	fm.WriteString("---\n")

	return os.WriteFile(path, []byte(fm.String()+content), 0644)
}

// Helper to check if a command is available
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
