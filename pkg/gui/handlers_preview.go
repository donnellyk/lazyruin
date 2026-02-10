package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) previewDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Preview.Mode == PreviewModeCardList {
		if listMove(&gui.state.Preview.SelectedCardIndex, len(gui.state.Preview.Cards), 1) {
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
		if listMove(&gui.state.Preview.SelectedCardIndex, len(gui.state.Preview.Cards), -1) {
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
		gui.refreshNotes(false)
	}
	gui.setContext(gui.state.PreviousContext)
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

func (gui *Gui) toggleMarkdown(g *gocui.Gui, v *gocui.View) error {
	gui.state.Preview.RenderMarkdown = !gui.state.Preview.RenderMarkdown
	gui.renderPreview()
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
		if gui.state.Queries.CurrentTab == QueriesTabParents {
			if len(gui.state.Parents.Items) > 0 {
				parent := gui.state.Parents.Items[gui.state.Parents.SelectedIndex]
				composed, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
				if err == nil {
					gui.state.Preview.Cards = []models.Note{composed}
				}
			}
		} else if len(gui.state.Queries.Items) > 0 {
			query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
			notes, err := gui.ruinCmd.Queries.Run(query.Name)
			if err == nil {
				gui.state.Preview.Cards = notes
			}
		}
	}

	gui.renderPreview()
}

func (gui *Gui) updatePreviewForNotes() {
	gui.state.Preview.Mode = PreviewModeSingleNote
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Preview "
		gui.renderPreview()
	}
}

// updatePreviewCardList is a shared helper for updating the preview with a card list.
func (gui *Gui) updatePreviewCardList(title string, loadFn func() ([]models.Note, error)) {
	notes, err := loadFn()
	if err != nil {
		return
	}
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.EditMode = false
	if gui.views.Preview != nil {
		gui.views.Preview.Title = title
		gui.renderPreview()
	}
}
