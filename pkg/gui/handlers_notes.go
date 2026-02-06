package gui

import (
	"os"
	"os/exec"
	"strings"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

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

func (gui *Gui) notesDown(g *gocui.Gui, v *gocui.View) error {
	if listMove(&gui.state.Notes.SelectedIndex, len(gui.state.Notes.Items), 1) {
		gui.renderNotes()
		gui.updatePreviewForNotes()
	}
	return nil
}

func (gui *Gui) notesUp(g *gocui.Gui, v *gocui.View) error {
	if listMove(&gui.state.Notes.SelectedIndex, len(gui.state.Notes.Items), -1) {
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

func (gui *Gui) newNote(g *gocui.Gui, v *gocui.View) error {
	gui.showInput("New Note", "Enter note content:", func(content string) error {
		if content == "" {
			return nil
		}
		_, err := gui.ruinCmd.Execute("log", content)
		if err != nil {
			return nil
		}
		gui.refreshNotes(false)
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
		gui.refreshNotes(false)
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
