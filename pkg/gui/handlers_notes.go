package gui

import (
	"os/exec"
	"strings"
	"time"

	guictx "kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) notesClick(g *gocui.Gui, v *gocui.View) error {
	notesCtx := gui.contexts.Notes
	idx := listClickIndex(gui.views.Notes, 3)
	if idx >= 0 && idx < len(notesCtx.Items) {
		notesCtx.SetSelectedLineIdx(idx)
		gui.syncNotesToLegacy()
	}
	gui.setContext(NotesContext)
	return nil
}

// cycleNotesTab cycles through All -> Today -> Recent tabs
func (gui *Gui) cycleNotesTab() {
	notesCtx := gui.contexts.Notes
	idx := (notesCtx.TabIndex() + 1) % len(guictx.NotesTabs)
	notesCtx.CurrentTab = guictx.NotesTabs[idx]
	notesCtx.SetSelectedLineIdx(0)
	gui.syncNotesToLegacy()
	gui.loadNotesForCurrentTab()
}

// switchNotesTabByIndex switches to a specific tab by index (for tab click)
func (gui *Gui) switchNotesTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(guictx.NotesTabs) {
		return nil
	}
	notesCtx := gui.contexts.Notes
	notesCtx.CurrentTab = guictx.NotesTabs[tabIndex]
	notesCtx.SetSelectedLineIdx(0)
	gui.syncNotesToLegacy()
	gui.loadNotesForCurrentTab()
	gui.setContext(NotesContext)
	return nil
}

// loadNotesForCurrentTab loads notes based on the current tab,
// resets selection, renders the list, and updates the preview.
func (gui *Gui) loadNotesForCurrentTab() {
	gui.fetchNotesForCurrentTab(false)
	gui.preview.updatePreviewForNotes()
}

// fetchNotesForCurrentTab loads notes for the current tab and renders the list.
// If preserve is true, the current selection is preserved by UUID; otherwise it resets to 0.
func (gui *Gui) fetchNotesForCurrentTab(preserve bool) {
	notesCtx := gui.contexts.Notes
	prevID := ""
	if preserve {
		prevID = notesCtx.GetSelectedItemId()
	}

	var notes []models.Note
	var err error

	opts := gui.buildSearchOptions()
	opts.Sort = "created:desc"
	opts.IncludeContent = true
	opts.StripTitle = true
	opts.StripGlobalTags = true

	switch notesCtx.CurrentTab {
	case guictx.NotesTabAll:
		opts.Limit = 50
		opts.Everything = true
		notes, err = gui.ruinCmd.Search.Search("", opts)
	case guictx.NotesTabToday:
		notes, err = gui.ruinCmd.Search.Search("created:today", opts)
	case guictx.NotesTabRecent:
		opts.Limit = 20
		recentDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		notes, err = gui.ruinCmd.Search.Search("after:"+recentDate, opts)
	}

	if err == nil {
		notesCtx.Items = notes
		if preserve && prevID != "" {
			if newIdx := notesCtx.GetList().FindIndexById(prevID); newIdx >= 0 {
				notesCtx.SetSelectedLineIdx(newIdx)
			} else {
				notesCtx.SetSelectedLineIdx(0)
			}
		} else {
			notesCtx.SetSelectedLineIdx(0)
		}
		notesCtx.ClampSelection()
		gui.syncNotesToLegacy()
	}
	gui.renderNotes()
	gui.updateNotesTab()
}

func (gui *Gui) viewNoteInPreview(g *gocui.Gui, v *gocui.View) error {
	if len(gui.contexts.Notes.Items) == 0 {
		return nil
	}
	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) editNote(g *gocui.Gui, v *gocui.View) error {
	notesCtx := gui.contexts.Notes
	note := notesCtx.Selected()
	if note == nil {
		return nil
	}
	return gui.openInEditor(note.Path)
}

func (gui *Gui) newNote(g *gocui.Gui, v *gocui.View) error {
	return gui.openCapture(g, v)
}

func (gui *Gui) deleteNote(g *gocui.Gui, v *gocui.View) error {
	notesCtx := gui.contexts.Notes
	note := notesCtx.Selected()
	if note == nil {
		return nil
	}

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
	notesCtx := gui.contexts.Notes
	note := notesCtx.Selected()
	if note == nil {
		return nil
	}

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
