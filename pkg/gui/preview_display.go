package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (c *PreviewController) toggleMarkdown(g *gocui.Gui, v *gocui.View) error {
	c.gui.state.Preview.RenderMarkdown = !c.gui.state.Preview.RenderMarkdown
	c.gui.renderPreview()
	return nil
}

func (c *PreviewController) toggleFrontmatter(g *gocui.Gui, v *gocui.View) error {
	c.gui.state.Preview.ShowFrontmatter = !c.gui.state.Preview.ShowFrontmatter
	c.gui.renderPreview()
	return nil
}

func (c *PreviewController) toggleTitle(g *gocui.Gui, v *gocui.View) error {
	c.gui.state.Preview.ShowTitle = !c.gui.state.Preview.ShowTitle
	c.reloadContent()
	return nil
}

func (c *PreviewController) toggleGlobalTags(g *gocui.Gui, v *gocui.View) error {
	c.gui.state.Preview.ShowGlobalTags = !c.gui.state.Preview.ShowGlobalTags
	c.reloadContent()
	return nil
}

// viewOptionsDialog shows the view options menu (displaced toggles).
func (c *PreviewController) viewOptionsDialog(g *gocui.Gui, v *gocui.View) error {
	fmLabel := "Show frontmatter"
	if c.gui.state.Preview.ShowFrontmatter {
		fmLabel = "Hide frontmatter"
	}
	titleLabel := "Show title"
	if c.gui.state.Preview.ShowTitle {
		titleLabel = "Hide title"
	}
	tagsLabel := "Show global tags"
	if c.gui.state.Preview.ShowGlobalTags {
		tagsLabel = "Hide global tags"
	}
	mdLabel := "Render markdown"
	if c.gui.state.Preview.RenderMarkdown {
		mdLabel = "Raw markdown"
	}

	c.gui.state.Dialog = &DialogState{
		Active: true,
		Type:   "menu",
		Title:  "View Options",
		MenuItems: []MenuItem{
			{Label: fmLabel, Key: "f", OnRun: func() error { return c.toggleFrontmatter(nil, nil) }},
			{Label: titleLabel, Key: "t", OnRun: func() error { return c.toggleTitle(nil, nil) }},
			{Label: tagsLabel, Key: "T", OnRun: func() error { return c.toggleGlobalTags(nil, nil) }},
			{Label: mdLabel, Key: "M", OnRun: func() error { return c.toggleMarkdown(nil, nil) }},
		},
		MenuSelection: 0,
	}
	return nil
}

// reloadContent reloads notes from CLI with current toggle settings,
// preserving selection indices and preview mode.
func (c *PreviewController) reloadContent() {
	// Reload notes for the Notes pane, preserving selection
	c.gui.fetchNotesForCurrentTab(true)

	// Reload cards in Preview pane
	if len(c.gui.state.Preview.Cards) > 0 {
		savedCardIdx := c.gui.state.Preview.SelectedCardIndex
		c.reloadPreviewCards()
		if savedCardIdx < len(c.gui.state.Preview.Cards) {
			c.gui.state.Preview.SelectedCardIndex = savedCardIdx
		}
	}
	c.gui.renderPreview()
}

// reloadPreviewCards reloads the preview cards based on what generated them
func (c *PreviewController) reloadPreviewCards() {
	c.gui.state.Preview.TemporarilyMoved = nil
	opts := c.gui.buildSearchOptions()

	// If there's an active search query, reload search results
	if c.gui.state.SearchQuery != "" {
		notes, err := c.gui.ruinCmd.Search.Search(c.gui.state.SearchQuery, opts)
		if err == nil {
			c.gui.state.Preview.Cards = notes
		}
		c.gui.renderPreview()
		return
	}

	// Otherwise, reload based on previous context
	switch c.gui.state.previousContext() {
	case NotesContext:
		// The notes list was already refreshed by reloadContent().
		// Find the updated note(s) by UUID.
		c.reloadPreviewCardsFromNotes()
	case TagsContext:
		if len(c.gui.state.Tags.Items) > 0 {
			tag := c.gui.state.Tags.Items[c.gui.state.Tags.SelectedIndex]
			notes, err := c.gui.ruinCmd.Search.Search(tag.Name, opts)
			if err == nil {
				c.gui.state.Preview.Cards = notes
			}
		}
	case QueriesContext:
		if c.gui.state.Queries.CurrentTab == QueriesTabParents {
			if len(c.gui.state.Parents.Items) > 0 {
				parent := c.gui.state.Parents.Items[c.gui.state.Parents.SelectedIndex]
				composed, err := c.gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
				if err == nil {
					c.gui.state.Preview.Cards = []models.Note{composed}
				}
			}
		} else if len(c.gui.state.Queries.Items) > 0 {
			query := c.gui.state.Queries.Items[c.gui.state.Queries.SelectedIndex]
			notes, err := c.gui.ruinCmd.Queries.Run(query.Name, opts)
			if err == nil {
				c.gui.state.Preview.Cards = notes
			}
		}
	default:
		c.reloadPreviewCardsFromNotes()
	}

	c.gui.renderPreview()
}

// reloadPreviewCardsFromNotes re-fetches each preview card by UUID using
// the current search options (respecting title/tag toggle state).
func (c *PreviewController) reloadPreviewCardsFromNotes() {
	opts := c.gui.buildSearchOptions()
	updated := make([]models.Note, 0, len(c.gui.state.Preview.Cards))
	for _, card := range c.gui.state.Preview.Cards {
		fresh, err := c.gui.ruinCmd.Search.Get(card.UUID, opts)
		if err == nil && fresh != nil {
			// ruin get doesn't return inline_tags; preserve from original
			if len(fresh.InlineTags) == 0 && len(card.InlineTags) > 0 {
				fresh.InlineTags = card.InlineTags
			}
			updated = append(updated, *fresh)
		} else {
			// Fallback: clear content so buildCardContent reads from disk
			card.Content = ""
			updated = append(updated, card)
		}
	}
	c.gui.state.Preview.Cards = updated
}

func (c *PreviewController) updatePreviewForNotes() {
	if len(c.gui.state.Notes.Items) == 0 {
		return
	}
	idx := c.gui.state.Notes.SelectedIndex
	if idx >= len(c.gui.state.Notes.Items) {
		return
	}
	note := c.gui.state.Notes.Items[idx]
	c.pushNavHistory()
	c.gui.state.Preview.Mode = PreviewModeCardList
	c.gui.state.Preview.Cards = []models.Note{note}
	c.gui.state.Preview.SelectedCardIndex = 0
	c.gui.state.Preview.CursorLine = 1
	c.gui.state.Preview.ScrollOffset = 0
	if c.gui.views.Preview != nil {
		c.gui.views.Preview.Title = " " + note.Title + " "
		c.gui.renderPreview()
	}
}

// updatePreviewCardList is a shared helper for updating the preview with a card list.
func (c *PreviewController) updatePreviewCardList(title string, loadFn func() ([]models.Note, error)) {
	notes, err := loadFn()
	if err != nil {
		return
	}
	c.pushNavHistory()
	c.gui.state.Preview.Mode = PreviewModeCardList
	c.gui.state.Preview.Cards = notes
	c.gui.state.Preview.SelectedCardIndex = 0
	c.gui.state.Preview.CursorLine = 1
	c.gui.state.Preview.ScrollOffset = 0
	if c.gui.views.Preview != nil {
		c.gui.views.Preview.Title = title
		c.gui.renderPreview()
	}
}
