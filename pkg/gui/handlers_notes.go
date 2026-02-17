package gui

import (
	"os/exec"
	"strings"
	"time"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) notesClick(g *gocui.Gui, v *gocui.View) error {
	idx := listClickIndex(v, 3)
	if idx >= 0 && idx < len(gui.state.Notes.Items) {
		gui.state.Notes.SelectedIndex = idx
	}
	gui.setContext(NotesContext)
	return nil
}

func (gui *Gui) notesWheelDown(g *gocui.Gui, v *gocui.View) error {
	scrollViewport(v, 3)
	return nil
}

func (gui *Gui) notesWheelUp(g *gocui.Gui, v *gocui.View) error {
	scrollViewport(v, -3)
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

// loadNotesForCurrentTab loads notes based on the current tab,
// resets selection, renders the list, and updates the preview.
func (gui *Gui) loadNotesForCurrentTab() {
	gui.fetchNotesForCurrentTab(false)
	gui.updatePreviewForNotes()
}

// fetchNotesForCurrentTab loads notes for the current tab and renders the list.
// If preserve is true, the current selection index is kept; otherwise it resets to 0.
func (gui *Gui) fetchNotesForCurrentTab(preserve bool) {
	savedIdx := gui.state.Notes.SelectedIndex

	var notes []models.Note
	var err error

	opts := gui.buildSearchOptions()
	opts.Sort = "created:desc"
	opts.IncludeContent = true
	opts.StripTitle = true
	opts.StripGlobalTags = true

	switch gui.state.Notes.CurrentTab {
	case NotesTabAll:
		opts.Limit = 50
		opts.Everything = true
		notes, err = gui.ruinCmd.Search.Search("", opts)
	case NotesTabToday:
		notes, err = gui.ruinCmd.Search.Search("created:today", opts)
	case NotesTabRecent:
		opts.Limit = 20
		recentDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		notes, err = gui.ruinCmd.Search.Search("after:"+recentDate, opts)
	}

	if err == nil {
		gui.state.Notes.Items = notes
		if preserve && savedIdx < len(notes) {
			gui.state.Notes.SelectedIndex = savedIdx
		} else {
			gui.state.Notes.SelectedIndex = 0
		}
	}
	gui.renderNotes()
	gui.updateNotesTab()
}

func (gui *Gui) notesPanel() *listPanel {
	return &listPanel{
		selectedIndex: &gui.state.Notes.SelectedIndex,
		itemCount:     func() int { return len(gui.state.Notes.Items) },
		render:        gui.renderNotes,
		updatePreview: gui.updatePreviewForNotes,
		context:       NotesContext,
	}
}

func (gui *Gui) notesDown(g *gocui.Gui, v *gocui.View) error {
	return gui.notesPanel().listDown(g, v)
}

func (gui *Gui) notesUp(g *gocui.Gui, v *gocui.View) error {
	return gui.notesPanel().listUp(g, v)
}

func (gui *Gui) notesTop(g *gocui.Gui, v *gocui.View) error {
	return gui.notesPanel().listTop(g, v)
}

func (gui *Gui) notesBottom(g *gocui.Gui, v *gocui.View) error {
	return gui.notesPanel().listBottom(g, v)
}

func (gui *Gui) viewNoteInPreview(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Notes.Items) == 0 {
		return nil
	}
	gui.setContext(PreviewContext)
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
	return gui.openCapture(g, v)
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
		err := gui.ruinCmd.Note.Delete(note.UUID)
		if err != nil {
			gui.showError(err)
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
