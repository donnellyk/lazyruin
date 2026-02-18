package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// PreviewController holds preview-specific handler methods.
// It wraps *Gui for access to shared state and services.
type PreviewController struct {
	gui *Gui
}

// NewPreviewController creates a new PreviewController.
func NewPreviewController(gui *Gui) *PreviewController {
	return &PreviewController{gui: gui}
}

// multiCardCount returns the number of items in the current preview.
func (c *PreviewController) multiCardCount() int {
	switch c.gui.state.Preview.Mode {
	case PreviewModeCardList:
		return len(c.gui.state.Preview.Cards)
	case PreviewModePickResults:
		return len(c.gui.state.Preview.PickResults)
	default:
		return 0
	}
}

// isContentLine returns true if lineNum is a body/content line (not a separator or blank).
func (c *PreviewController) isContentLine(lineNum int) bool {
	for _, r := range c.gui.state.Preview.CardLineRanges {
		if lineNum > r[0] && lineNum < r[1]-1 {
			return true
		}
	}
	return false
}

// syncCardIndexFromCursor updates SelectedCardIndex based on CursorLine position.
func (c *PreviewController) syncCardIndexFromCursor() {
	ranges := c.gui.state.Preview.CardLineRanges
	cursor := c.gui.state.Preview.CursorLine
	for i, r := range ranges {
		if cursor >= r[0] && cursor < r[1] {
			c.gui.state.Preview.SelectedCardIndex = i
			return
		}
	}
	// On blank line between cards - attribute to next card
	for i := 0; i < len(ranges)-1; i++ {
		if cursor >= ranges[i][1] && cursor < ranges[i+1][0] {
			c.gui.state.Preview.SelectedCardIndex = i + 1
			return
		}
	}
}

func (c *PreviewController) previewDown(g *gocui.Gui, v *gocui.View) error {
	ranges := c.gui.state.Preview.CardLineRanges
	if len(ranges) > 0 {
		maxLine := ranges[len(ranges)-1][1] - 1
		cursor := c.gui.state.Preview.CursorLine
		for cursor < maxLine {
			cursor++
			if c.isContentLine(cursor) {
				break
			}
		}
		if c.isContentLine(cursor) && cursor != c.gui.state.Preview.CursorLine {
			c.gui.state.Preview.CursorLine = cursor
			c.syncCardIndexFromCursor()
			c.gui.renderPreview()
		}
	}
	return nil
}

func (c *PreviewController) previewUp(g *gocui.Gui, v *gocui.View) error {
	cursor := c.gui.state.Preview.CursorLine
	for cursor > 0 {
		cursor--
		if c.isContentLine(cursor) {
			break
		}
	}
	if c.isContentLine(cursor) && cursor != c.gui.state.Preview.CursorLine {
		c.gui.state.Preview.CursorLine = cursor
		c.syncCardIndexFromCursor()
		c.gui.renderPreview()
	}
	return nil
}

// previewCardDown jumps to the next card (J).
func (c *PreviewController) previewCardDown(g *gocui.Gui, v *gocui.View) error {
	if listMove(&c.gui.state.Preview.SelectedCardIndex, c.multiCardCount(), 1) {
		ranges := c.gui.state.Preview.CardLineRanges
		idx := c.gui.state.Preview.SelectedCardIndex
		if idx < len(ranges) {
			c.gui.state.Preview.CursorLine = ranges[idx][0] + 1 // first content line
		}
		c.gui.renderPreview()
	}
	return nil
}

// previewCardUp jumps to the previous card (K).
func (c *PreviewController) previewCardUp(g *gocui.Gui, v *gocui.View) error {
	if listMove(&c.gui.state.Preview.SelectedCardIndex, c.multiCardCount(), -1) {
		ranges := c.gui.state.Preview.CardLineRanges
		idx := c.gui.state.Preview.SelectedCardIndex
		if idx < len(ranges) {
			c.gui.state.Preview.CursorLine = ranges[idx][0] + 1 // first content line
		}
		c.gui.renderPreview()
	}
	return nil
}

// previewNextHeader jumps to the next markdown header (}).
func (c *PreviewController) previewNextHeader(g *gocui.Gui, v *gocui.View) error {
	cursor := c.gui.state.Preview.CursorLine
	for _, h := range c.gui.state.Preview.HeaderLines {
		if h > cursor {
			c.gui.state.Preview.CursorLine = h
			c.syncCardIndexFromCursor()
			c.gui.renderPreview()
			return nil
		}
	}
	return nil
}

// previewPrevHeader jumps to the previous markdown header ({).
func (c *PreviewController) previewPrevHeader(g *gocui.Gui, v *gocui.View) error {
	cursor := c.gui.state.Preview.CursorLine
	for i := len(c.gui.state.Preview.HeaderLines) - 1; i >= 0; i-- {
		if c.gui.state.Preview.HeaderLines[i] < cursor {
			c.gui.state.Preview.CursorLine = c.gui.state.Preview.HeaderLines[i]
			c.syncCardIndexFromCursor()
			c.gui.renderPreview()
			return nil
		}
	}
	return nil
}

func (c *PreviewController) previewScrollDown(g *gocui.Gui, v *gocui.View) error {
	if c.gui.state.ActiveOverlay == OverlayPalette {
		if c.gui.views.PaletteList != nil {
			scrollViewport(c.gui.views.PaletteList, 3)
		}
		return nil
	}
	if v == nil || v.Name() != PreviewView {
		return nil
	}
	c.gui.state.Preview.ScrollOffset += 3
	v.SetOrigin(0, c.gui.state.Preview.ScrollOffset)
	return nil
}

func (c *PreviewController) previewScrollUp(g *gocui.Gui, v *gocui.View) error {
	if c.gui.state.ActiveOverlay == OverlayPalette {
		if c.gui.views.PaletteList != nil {
			scrollViewport(c.gui.views.PaletteList, -3)
		}
		return nil
	}
	if v == nil || v.Name() != PreviewView {
		return nil
	}
	c.gui.state.Preview.ScrollOffset -= 3
	if c.gui.state.Preview.ScrollOffset < 0 {
		c.gui.state.Preview.ScrollOffset = 0
	}
	v.SetOrigin(0, c.gui.state.Preview.ScrollOffset)
	return nil
}

func (c *PreviewController) previewClick(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	ox, oy := v.Origin()
	absX := cx + ox
	absY := cy + oy

	// Check if click lands on a link
	c.extractLinks()
	for _, link := range c.gui.state.Preview.Links {
		if link.Line == absY && absX >= link.Col && absX < link.Col+link.Len {
			return c.followLink(link)
		}
	}

	// Snap click to nearest content line within the card
	clickLine := absY
	for i, lr := range c.gui.state.Preview.CardLineRanges {
		if absY >= lr[0] && absY < lr[1] {
			c.gui.state.Preview.SelectedCardIndex = i
			if !c.isContentLine(clickLine) {
				clickLine = lr[0] + 1 // first content line
			}
			break
		}
	}
	c.gui.state.Preview.CursorLine = clickLine

	c.gui.setContext(PreviewContext)
	c.gui.renderPreview()
	return nil
}

func (c *PreviewController) previewBack(g *gocui.Gui, v *gocui.View) error {
	c.gui.popContext()
	return nil
}

// currentPreviewCard returns the currently selected card, or nil if none.
func (c *PreviewController) currentPreviewCard() *models.Note {
	idx := c.gui.state.Preview.SelectedCardIndex
	if idx >= len(c.gui.state.Preview.Cards) {
		return nil
	}
	return &c.gui.state.Preview.Cards[idx]
}

func (c *PreviewController) focusNoteFromPreview(g *gocui.Gui, v *gocui.View) error {
	if len(c.gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := c.gui.state.Preview.Cards[c.gui.state.Preview.SelectedCardIndex]

	for i, note := range c.gui.state.Notes.Items {
		if note.UUID == card.UUID {
			c.gui.state.Notes.SelectedIndex = i
			c.gui.setContext(NotesContext)
			c.gui.renderNotes()
			return nil
		}
	}

	return nil
}

// openCardInEditor opens the currently selected card in $EDITOR.
func (c *PreviewController) openCardInEditor(g *gocui.Gui, v *gocui.View) error {
	card := c.currentPreviewCard()
	if card == nil {
		return nil
	}
	return c.gui.openInEditor(card.Path)
}
