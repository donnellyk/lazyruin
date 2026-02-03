package gui

import (
	"os"
	"os/exec"
	"strings"

	"kvnd/lazyruin/pkg/commands"

	"github.com/jesseduffield/gocui"
)

// Global handlers

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	order := []ContextKey{NotesContext, QueriesContext, TagsContext, PreviewContext}

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

func (gui *Gui) focusNotes(g *gocui.Gui, v *gocui.View) error {
	gui.setContext(NotesContext)
	return nil
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

func (gui *Gui) editNote(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}

	note := gui.state.Notes.Items[gui.state.Notes.SelectedIndex]
	return gui.openInEditor(note.Path)
}

func (gui *Gui) openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	gui.g.Close()

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	// Re-initialize the GUI
	newG, initErr := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
	})
	if initErr != nil {
		return initErr
	}

	gui.g = newG
	newG.Mouse = true
	newG.Cursor = false
	newG.SetManager(gocui.ManagerFunc(gui.layout))

	if keybindErr := gui.setupKeybindings(); keybindErr != nil {
		return keybindErr
	}

	gui.refreshAll()

	return err
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

// Preview update helpers

func (gui *Gui) updatePreviewForNotes() {
	gui.state.Preview.Mode = PreviewModeSingleNote
	gui.state.Preview.ScrollOffset = 0
	gui.views.Preview.Title = " Preview "
	gui.renderPreview()
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
	gui.views.Preview.Title = " Preview: #" + tag.Name + " "
	gui.renderPreview()
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
	gui.views.Preview.Title = " Preview: " + query.Name + " "
	gui.renderPreview()
}
