package gui

import (
	"github.com/jesseduffield/gocui"
)

func (c *PreviewController) deleteCardFromPreview(g *gocui.Gui, v *gocui.View) error {
	if len(c.gui.state.Preview.Cards) == 0 {
		return nil
	}

	card := c.gui.state.Preview.Cards[c.gui.state.Preview.SelectedCardIndex]
	title := card.Title
	if title == "" {
		title = card.Path
	}
	if len(title) > 30 {
		title = title[:30] + "..."
	}

	c.gui.showConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
		err := c.gui.ruinCmd.Note.Delete(card.UUID)
		if err != nil {
			c.gui.showError(err)
			return nil
		}
		idx := c.gui.state.Preview.SelectedCardIndex
		c.gui.state.Preview.Cards = append(c.gui.state.Preview.Cards[:idx], c.gui.state.Preview.Cards[idx+1:]...)
		if c.gui.state.Preview.SelectedCardIndex >= len(c.gui.state.Preview.Cards) && c.gui.state.Preview.SelectedCardIndex > 0 {
			c.gui.state.Preview.SelectedCardIndex--
		}
		c.gui.refreshNotes(false)
		c.gui.renderPreview()
		return nil
	})
	return nil
}

func (c *PreviewController) moveCardHandler(g *gocui.Gui, v *gocui.View) error {
	if len(c.gui.state.Preview.Cards) <= 1 {
		return nil
	}
	c.showMoveOverlay()
	return nil
}

func (c *PreviewController) showMoveOverlay() {
	c.gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "Move",
		MenuItems: []MenuItem{
			{Label: "Move card up", Key: "u", OnRun: func() error { return c.moveCard("up") }},
			{Label: "Move card down", Key: "d", OnRun: func() error { return c.moveCard("down") }},
		},
		MenuSelection: 0,
	}
}

func (c *PreviewController) moveCard(direction string) error {
	idx := c.gui.state.Preview.SelectedCardIndex
	if direction == "up" {
		if idx <= 0 {
			return nil
		}
		c.gui.state.Preview.Cards[idx], c.gui.state.Preview.Cards[idx-1] = c.gui.state.Preview.Cards[idx-1], c.gui.state.Preview.Cards[idx]
		c.gui.state.Preview.SelectedCardIndex--
	} else {
		if idx >= len(c.gui.state.Preview.Cards)-1 {
			return nil
		}
		c.gui.state.Preview.Cards[idx], c.gui.state.Preview.Cards[idx+1] = c.gui.state.Preview.Cards[idx+1], c.gui.state.Preview.Cards[idx]
		c.gui.state.Preview.SelectedCardIndex++
	}

	if c.gui.state.Preview.TemporarilyMoved == nil {
		c.gui.state.Preview.TemporarilyMoved = make(map[int]bool)
	}
	c.gui.state.Preview.TemporarilyMoved[c.gui.state.Preview.SelectedCardIndex] = true

	// Render once to compute CardLineRanges for the new order,
	// then move cursor to the card's new position and render again.
	c.gui.renderPreview()
	newIdx := c.gui.state.Preview.SelectedCardIndex
	if newIdx < len(c.gui.state.Preview.CardLineRanges) {
		c.gui.state.Preview.CursorLine = c.gui.state.Preview.CardLineRanges[newIdx][0] + 1
	}
	c.gui.renderPreview()
	return nil
}

func (c *PreviewController) mergeCardHandler(g *gocui.Gui, v *gocui.View) error {
	if len(c.gui.state.Preview.Cards) <= 1 {
		return nil
	}
	c.showMergeOverlay()
	return nil
}

func (c *PreviewController) executeMerge(direction string) error {
	idx := c.gui.state.Preview.SelectedCardIndex
	var targetIdx, sourceIdx int
	if direction == "down" {
		if idx >= len(c.gui.state.Preview.Cards)-1 {
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

	target := c.gui.state.Preview.Cards[targetIdx]
	source := c.gui.state.Preview.Cards[sourceIdx]

	result, err := c.gui.ruinCmd.Note.Merge(target.UUID, source.UUID, true, false)
	if err != nil {
		c.gui.showError(err)
		return nil
	}

	// Clear target content so it re-reads from disk
	c.gui.state.Preview.Cards[targetIdx].Content = ""
	if len(result.TagsMerged) > 0 {
		c.gui.state.Preview.Cards[targetIdx].Tags = result.TagsMerged
	}

	// Remove source from cards
	c.gui.state.Preview.Cards = append(c.gui.state.Preview.Cards[:sourceIdx], c.gui.state.Preview.Cards[sourceIdx+1:]...)
	if c.gui.state.Preview.SelectedCardIndex >= len(c.gui.state.Preview.Cards) {
		c.gui.state.Preview.SelectedCardIndex = len(c.gui.state.Preview.Cards) - 1
	}
	if c.gui.state.Preview.SelectedCardIndex < 0 {
		c.gui.state.Preview.SelectedCardIndex = 0
	}

	c.gui.refreshNotes(false)
	c.gui.renderPreview()
	return nil
}

// showMergeOverlay shows the merge direction menu.
func (c *PreviewController) showMergeOverlay() {
	c.gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "Merge",
		MenuItems: []MenuItem{
			{Label: "Merge card below into this one", Key: "d", OnRun: func() error { return c.executeMerge("down") }},
			{Label: "Merge card above into this one", Key: "u", OnRun: func() error { return c.executeMerge("up") }},
		},
		MenuSelection: 0,
	}
}

// orderCards persists the current card order to frontmatter order fields.
func (c *PreviewController) orderCards() error {
	for i, card := range c.gui.state.Preview.Cards {
		if err := c.gui.ruinCmd.Note.SetOrder(card.UUID, i+1); err != nil {
			c.gui.showError(err)
			return nil
		}
	}
	c.gui.state.Preview.TemporarilyMoved = nil
	c.reloadContent()
	return nil
}
