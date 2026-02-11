package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// isMultiCardPreview returns true when the preview shows multiple selectable items.
func (gui *Gui) isMultiCardPreview() bool {
	return gui.state.Preview.Mode == PreviewModeCardList || gui.state.Preview.Mode == PreviewModePickResults
}

// multiCardCount returns the number of items in the current multi-card preview.
func (gui *Gui) multiCardCount() int {
	switch gui.state.Preview.Mode {
	case PreviewModeCardList:
		return len(gui.state.Preview.Cards)
	case PreviewModePickResults:
		return len(gui.state.Preview.PickResults)
	default:
		return 0
	}
}

// isContentLine returns true if lineNum is a body/content line (not a separator or blank).
func (gui *Gui) isContentLine(lineNum int) bool {
	for _, r := range gui.state.Preview.CardLineRanges {
		if lineNum > r[0] && lineNum < r[1]-1 {
			return true
		}
	}
	return false
}

// syncCardIndexFromCursor updates SelectedCardIndex based on CursorLine position.
func (gui *Gui) syncCardIndexFromCursor() {
	ranges := gui.state.Preview.CardLineRanges
	cursor := gui.state.Preview.CursorLine
	for i, r := range ranges {
		if cursor >= r[0] && cursor < r[1] {
			gui.state.Preview.SelectedCardIndex = i
			return
		}
	}
	// On blank line between cards - attribute to next card
	for i := 0; i < len(ranges)-1; i++ {
		if cursor >= ranges[i][1] && cursor < ranges[i+1][0] {
			gui.state.Preview.SelectedCardIndex = i + 1
			return
		}
	}
}

func (gui *Gui) previewDown(g *gocui.Gui, v *gocui.View) error {
	if gui.isMultiCardPreview() {
		ranges := gui.state.Preview.CardLineRanges
		if len(ranges) > 0 {
			maxLine := ranges[len(ranges)-1][1] - 1
			cursor := gui.state.Preview.CursorLine
			for cursor < maxLine {
				cursor++
				if gui.isContentLine(cursor) {
					break
				}
			}
			if gui.isContentLine(cursor) && cursor != gui.state.Preview.CursorLine {
				gui.state.Preview.CursorLine = cursor
				gui.syncCardIndexFromCursor()
				gui.renderPreview()
			}
		}
	} else {
		gui.state.Preview.ScrollOffset++
		gui.renderPreview()
	}
	return nil
}

func (gui *Gui) previewUp(g *gocui.Gui, v *gocui.View) error {
	if gui.isMultiCardPreview() {
		cursor := gui.state.Preview.CursorLine
		for cursor > 0 {
			cursor--
			if gui.isContentLine(cursor) {
				break
			}
		}
		if gui.isContentLine(cursor) && cursor != gui.state.Preview.CursorLine {
			gui.state.Preview.CursorLine = cursor
			gui.syncCardIndexFromCursor()
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

// previewCardDown jumps to the next card (J).
func (gui *Gui) previewCardDown(g *gocui.Gui, v *gocui.View) error {
	if gui.isMultiCardPreview() {
		if listMove(&gui.state.Preview.SelectedCardIndex, gui.multiCardCount(), 1) {
			ranges := gui.state.Preview.CardLineRanges
			idx := gui.state.Preview.SelectedCardIndex
			if idx < len(ranges) {
				gui.state.Preview.CursorLine = ranges[idx][0] + 1 // first content line
			}
			gui.renderPreview()
		}
	}
	return nil
}

// previewNextHeader jumps to the next markdown header (]).
func (gui *Gui) previewNextHeader(g *gocui.Gui, v *gocui.View) error {
	cursor := gui.state.Preview.CursorLine
	for _, h := range gui.state.Preview.HeaderLines {
		if h > cursor {
			gui.state.Preview.CursorLine = h
			gui.syncCardIndexFromCursor()
			gui.renderPreview()
			return nil
		}
	}
	return nil
}

// previewPrevHeader jumps to the previous markdown header ([).
func (gui *Gui) previewPrevHeader(g *gocui.Gui, v *gocui.View) error {
	cursor := gui.state.Preview.CursorLine
	for i := len(gui.state.Preview.HeaderLines) - 1; i >= 0; i-- {
		if gui.state.Preview.HeaderLines[i] < cursor {
			gui.state.Preview.CursorLine = gui.state.Preview.HeaderLines[i]
			gui.syncCardIndexFromCursor()
			gui.renderPreview()
			return nil
		}
	}
	return nil
}

// previewCardUp jumps to the previous card (K).
func (gui *Gui) previewCardUp(g *gocui.Gui, v *gocui.View) error {
	if gui.isMultiCardPreview() {
		if listMove(&gui.state.Preview.SelectedCardIndex, gui.multiCardCount(), -1) {
			ranges := gui.state.Preview.CardLineRanges
			idx := gui.state.Preview.SelectedCardIndex
			if idx < len(ranges) {
				gui.state.Preview.CursorLine = ranges[idx][0] + 1 // first content line
			}
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
	if !gui.isMultiCardPreview() {
		gui.setContext(PreviewContext)
		return nil
	}

	_, cy := v.Cursor()
	_, oy := v.Origin()
	absY := cy + oy

	// Snap click to nearest content line within the card
	clickLine := absY
	for i, lr := range gui.state.Preview.CardLineRanges {
		if absY >= lr[0] && absY < lr[1] {
			gui.state.Preview.SelectedCardIndex = i
			if !gui.isContentLine(clickLine) {
				clickLine = lr[0] + 1 // first content line
			}
			break
		}
	}
	gui.state.Preview.CursorLine = clickLine

	gui.setContext(PreviewContext)
	gui.renderPreview()
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
	gui.state.Preview.CursorLine = 1
	gui.state.Preview.ScrollOffset = 0
	gui.state.Preview.EditMode = false
	if gui.views.Preview != nil {
		gui.views.Preview.Title = title
		gui.renderPreview()
	}
}
