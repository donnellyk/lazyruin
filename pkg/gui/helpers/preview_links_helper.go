package helpers

import (
	"os/exec"
	"regexp"
	"strings"

	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// PreviewLinksHelper handles link extraction, highlighting, and following
// for the preview panel.
type PreviewLinksHelper struct {
	c *HelperCommon
}

// NewPreviewLinksHelper creates a new PreviewLinksHelper.
func NewPreviewLinksHelper(c *HelperCommon) *PreviewLinksHelper {
	return &PreviewLinksHelper{c: c}
}

func (self *PreviewLinksHelper) ctx() *context.PreviewContext {
	return self.c.GuiCommon().Contexts().Preview
}

func (self *PreviewLinksHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// ExtractLinks parses the preview content for wiki-links and URLs.
func (self *PreviewLinksHelper) ExtractLinks() {
	pc := self.ctx()
	pc.Links = nil
	v := self.view()
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
			pc.Links = append(pc.Links, context.PreviewLink{
				Text: text, Line: lineNum, Col: match[0], Len: match[1] - match[0],
			})
		}
		for _, match := range urlRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			pc.Links = append(pc.Links, context.PreviewLink{
				Text: text, Line: lineNum, Col: match[0], Len: match[1] - match[0],
			})
		}
	}
}

// HighlightNextLink cycles to the next link.
func (self *PreviewLinksHelper) HighlightNextLink() error {
	self.ExtractLinks()
	pc := self.ctx()
	links := pc.Links
	if len(links) == 0 {
		return nil
	}
	cur := pc.RenderedLink
	next := cur + 1
	if next >= len(links) {
		next = 0
	}
	pc.HighlightedLink = next
	pc.CursorLine = links[next].Line
	self.c.Helpers().PreviewNav().SyncCardIndexFromCursor()
	self.c.GuiCommon().RenderPreview()
	return nil
}

// HighlightPrevLink cycles to the previous link.
func (self *PreviewLinksHelper) HighlightPrevLink() error {
	self.ExtractLinks()
	pc := self.ctx()
	links := pc.Links
	if len(links) == 0 {
		return nil
	}
	cur := pc.RenderedLink
	prev := cur - 1
	if prev < 0 {
		prev = len(links) - 1
	}
	pc.HighlightedLink = prev
	pc.CursorLine = links[prev].Line
	self.c.Helpers().PreviewNav().SyncCardIndexFromCursor()
	self.c.GuiCommon().RenderPreview()
	return nil
}

// OpenLink opens the currently highlighted link.
func (self *PreviewLinksHelper) OpenLink() error {
	pc := self.ctx()
	links := pc.Links
	hl := pc.RenderedLink
	if hl < 0 || hl >= len(links) {
		return nil
	}
	return self.FollowLink(links[hl])
}

// FollowLink navigates to a wiki-link target or opens a URL in the browser.
func (self *PreviewLinksHelper) FollowLink(link context.PreviewLink) error {
	text := link.Text

	if strings.HasPrefix(text, "[[") && strings.HasSuffix(text, "]]") {
		target := text[2 : len(text)-2]
		if i := strings.Index(target, "#"); i >= 0 {
			target = target[:i]
		}
		if target == "" {
			return nil
		}
		opts := self.c.GuiCommon().BuildSearchOptions()
		note, err := self.c.RuinCmd().Search.GetByTitle(target, opts)
		if err != nil || note == nil {
			return nil
		}
		nav := self.c.Helpers().PreviewNav()
		nav.PushNavHistory()
		pc := self.ctx()
		pc.Mode = context.PreviewModeCardList
		pc.Cards = []models.Note{*note}
		pc.SelectedCardIndex = 0
		pc.CursorLine = 1
		pc.ScrollOffset = 0
		if v := self.view(); v != nil {
			v.Title = " " + note.Title + " "
		}
		self.c.GuiCommon().RenderPreview()
		return nil
	}

	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		exec.Command("open", text).Start()
	}
	return nil
}
