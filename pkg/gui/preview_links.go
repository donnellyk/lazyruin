package gui

import (
	"os/exec"
	"regexp"
	"strings"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// extractLinks parses the preview content for wiki-links ([[...]]) and URLs.
func (c *PreviewController) extractLinks() {
	c.gui.state.Preview.Links = nil
	v := c.gui.views.Preview
	if v == nil {
		return
	}

	lines := v.ViewBufferLines()
	wikiRe := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	urlRe := regexp.MustCompile(`https?://[^\s)\]>]+`)

	for lineNum, line := range lines {
		plain := stripAnsi(line)
		for _, match := range wikiRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			c.gui.state.Preview.Links = append(c.gui.state.Preview.Links, PreviewLink{
				Text: text,
				Line: lineNum,
				Col:  match[0],
				Len:  match[1] - match[0],
			})
		}
		for _, match := range urlRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			c.gui.state.Preview.Links = append(c.gui.state.Preview.Links, PreviewLink{
				Text: text,
				Line: lineNum,
				Col:  match[0],
				Len:  match[1] - match[0],
			})
		}
	}
}

// highlightNextLink cycles to the next link (l).
func (c *PreviewController) highlightNextLink(g *gocui.Gui, v *gocui.View) error {
	c.extractLinks()
	links := c.gui.state.Preview.Links
	if len(links) == 0 {
		return nil
	}
	cur := c.gui.state.Preview.renderedLink
	next := cur + 1
	if next >= len(links) {
		next = 0
	}
	c.gui.state.Preview.HighlightedLink = next
	c.gui.state.Preview.CursorLine = links[next].Line
	c.syncCardIndexFromCursor()
	c.gui.renderPreview()
	return nil
}

// highlightPrevLink cycles to the previous link (L).
func (c *PreviewController) highlightPrevLink(g *gocui.Gui, v *gocui.View) error {
	c.extractLinks()
	links := c.gui.state.Preview.Links
	if len(links) == 0 {
		return nil
	}
	cur := c.gui.state.Preview.renderedLink
	prev := cur - 1
	if prev < 0 {
		prev = len(links) - 1
	}
	c.gui.state.Preview.HighlightedLink = prev
	c.gui.state.Preview.CursorLine = links[prev].Line
	c.syncCardIndexFromCursor()
	c.gui.renderPreview()
	return nil
}

// openLink opens the currently highlighted link.
func (c *PreviewController) openLink(g *gocui.Gui, v *gocui.View) error {
	links := c.gui.state.Preview.Links
	hl := c.gui.state.Preview.renderedLink
	if hl < 0 || hl >= len(links) {
		return nil
	}
	return c.followLink(links[hl])
}

// followLink navigates to a wiki-link target or opens a URL in the browser.
func (c *PreviewController) followLink(link PreviewLink) error {
	text := link.Text

	// Wiki-link: strip [[ and ]]
	if strings.HasPrefix(text, "[[") && strings.HasSuffix(text, "]]") {
		target := text[2 : len(text)-2]
		// Strip header fragment
		if i := strings.Index(target, "#"); i >= 0 {
			target = target[:i]
		}
		if target == "" {
			return nil
		}
		opts := c.gui.buildSearchOptions()
		note, err := c.gui.ruinCmd.Search.GetByTitle(target, opts)
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
		c.gui.renderPreview()
		return nil
	}

	// URL: open in browser
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		exec.Command("open", text).Start()
	}
	return nil
}
