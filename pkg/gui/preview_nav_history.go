package gui

import (
	"strings"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// pushNavHistory captures the current preview state onto the nav history stack.
// Call this before changing Preview.Cards for a new navigation.
func (c *PreviewController) pushNavHistory() {
	if len(c.gui.state.Preview.Cards) == 0 && len(c.gui.state.Preview.PickResults) == 0 {
		return
	}

	title := ""
	if c.gui.views.Preview != nil {
		title = c.gui.views.Preview.Title
	}

	entry := NavEntry{
		Cards:             append([]models.Note(nil), c.gui.state.Preview.Cards...),
		SelectedCardIndex: c.gui.state.Preview.SelectedCardIndex,
		CursorLine:        c.gui.state.Preview.CursorLine,
		ScrollOffset:      c.gui.state.Preview.ScrollOffset,
		Mode:              c.gui.state.Preview.Mode,
		Title:             title,
		PickResults:       append([]models.PickResult(nil), c.gui.state.Preview.PickResults...),
	}

	// Truncate any forward entries
	if c.gui.state.NavIndex >= 0 && c.gui.state.NavIndex < len(c.gui.state.NavHistory)-1 {
		c.gui.state.NavHistory = c.gui.state.NavHistory[:c.gui.state.NavIndex+1]
	}

	c.gui.state.NavHistory = append(c.gui.state.NavHistory, entry)
	c.gui.state.NavIndex = len(c.gui.state.NavHistory) - 1

	// Cap at 50 entries
	if len(c.gui.state.NavHistory) > 50 {
		c.gui.state.NavHistory = c.gui.state.NavHistory[len(c.gui.state.NavHistory)-50:]
		c.gui.state.NavIndex = len(c.gui.state.NavHistory) - 1
	}
}

// captureCurrentNavEntry returns a NavEntry for the current preview state.
func (c *PreviewController) captureCurrentNavEntry() NavEntry {
	title := ""
	if c.gui.views.Preview != nil {
		title = c.gui.views.Preview.Title
	}
	return NavEntry{
		Cards:             append([]models.Note(nil), c.gui.state.Preview.Cards...),
		SelectedCardIndex: c.gui.state.Preview.SelectedCardIndex,
		CursorLine:        c.gui.state.Preview.CursorLine,
		ScrollOffset:      c.gui.state.Preview.ScrollOffset,
		Mode:              c.gui.state.Preview.Mode,
		Title:             title,
		PickResults:       append([]models.PickResult(nil), c.gui.state.Preview.PickResults...),
	}
}

// restoreNavEntry restores preview state from a NavEntry.
func (c *PreviewController) restoreNavEntry(entry NavEntry) {
	c.gui.state.Preview.Mode = entry.Mode
	c.gui.state.Preview.Cards = append([]models.Note(nil), entry.Cards...)
	c.gui.state.Preview.PickResults = append([]models.PickResult(nil), entry.PickResults...)
	c.gui.state.Preview.SelectedCardIndex = entry.SelectedCardIndex
	c.gui.state.Preview.CursorLine = entry.CursorLine
	c.gui.state.Preview.ScrollOffset = entry.ScrollOffset
	if c.gui.views.Preview != nil {
		c.gui.views.Preview.Title = entry.Title
		c.gui.views.Preview.SetOrigin(0, entry.ScrollOffset)
	}
	c.gui.renderPreview()
}

func (c *PreviewController) navBack(g *gocui.Gui, v *gocui.View) error {
	if c.gui.state.NavIndex < 0 || len(c.gui.state.NavHistory) == 0 {
		return nil
	}

	// Save current state at current position
	c.gui.state.NavHistory[c.gui.state.NavIndex] = c.captureCurrentNavEntry()

	// If we're at the end and haven't pushed current yet, push it first
	if c.gui.state.NavIndex == len(c.gui.state.NavHistory)-1 {
		// We're at the top — push current state as forward entry
		c.gui.state.NavHistory = append(c.gui.state.NavHistory, c.captureCurrentNavEntry())
		// NavIndex stays, we go back
	}

	if c.gui.state.NavIndex <= 0 {
		return nil
	}

	c.gui.state.NavIndex--
	c.restoreNavEntry(c.gui.state.NavHistory[c.gui.state.NavIndex])
	return nil
}

func (c *PreviewController) navForward(g *gocui.Gui, v *gocui.View) error {
	if c.gui.state.NavIndex >= len(c.gui.state.NavHistory)-1 {
		return nil
	}

	// Save current state at current position
	c.gui.state.NavHistory[c.gui.state.NavIndex] = c.captureCurrentNavEntry()

	c.gui.state.NavIndex++
	c.restoreNavEntry(c.gui.state.NavHistory[c.gui.state.NavIndex])
	return nil
}

// showNavHistory shows the navigation history stack in a menu dialog.
func (c *PreviewController) showNavHistory() error {
	if len(c.gui.state.NavHistory) == 0 {
		return nil
	}

	var items []MenuItem
	for i := len(c.gui.state.NavHistory) - 1; i >= 0; i-- {
		entry := c.gui.state.NavHistory[i]
		label := strings.TrimSpace(entry.Title)
		if label == "" {
			label = "(untitled)"
		}
		if i == c.gui.state.NavIndex {
			label = "> " + label
		}
		idx := i
		items = append(items, MenuItem{
			Label: label,
			OnRun: func() error {
				c.gui.state.NavHistory[c.gui.state.NavIndex] = c.captureCurrentNavEntry()
				c.gui.state.NavIndex = idx
				c.restoreNavEntry(c.gui.state.NavHistory[idx])
				return nil
			},
		})
	}

	c.gui.state.Dialog = &DialogState{
		Active:        true,
		Type:          "menu",
		Title:         "Navigation History",
		MenuItems:     items,
		MenuSelection: 0,
	}
	return nil
}

// openNoteByUUID loads a note by UUID and displays it in the preview.
func (c *PreviewController) openNoteByUUID(uuid string) error {
	opts := c.gui.buildSearchOptions()
	note, err := c.gui.ruinCmd.Search.Get(uuid, opts)
	if err != nil || note == nil {
		return nil
	}
	c.pushNavHistory()
	c.gui.state.Preview.Mode = PreviewModeCardList
	c.gui.state.Preview.Cards = []models.Note{*note}
	c.gui.state.Preview.SelectedCardIndex = 0
	c.gui.state.Preview.CursorLine = 1
	c.gui.state.Preview.ScrollOffset = 0
	if c.gui.views.Preview != nil {
		c.gui.views.Preview.Title = " " + note.Title + " "
	}
	c.gui.setContext(PreviewContext)
	c.gui.renderPreview()
	return nil
}
