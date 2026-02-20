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

func (self *PreviewLinksHelper) activeCtx() context.IPreviewContext {
	return self.c.GuiCommon().Contexts().ActivePreview()
}

func (self *PreviewLinksHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// ExtractLinks parses the preview content for wiki-links and URLs.
func (self *PreviewLinksHelper) ExtractLinks() {
	ns := self.activeCtx().NavState()
	ns.Links = nil
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
			ns.Links = append(ns.Links, context.PreviewLink{
				Text: text, Line: lineNum, Col: match[0], Len: match[1] - match[0],
			})
		}
		for _, match := range urlRe.FindAllStringIndex(plain, -1) {
			text := plain[match[0]:match[1]]
			ns.Links = append(ns.Links, context.PreviewLink{
				Text: text, Line: lineNum, Col: match[0], Len: match[1] - match[0],
			})
		}
	}
}

// HighlightNextLink cycles to the next link.
func (self *PreviewLinksHelper) HighlightNextLink() error {
	self.ExtractLinks()
	ns := self.activeCtx().NavState()
	links := ns.Links
	if len(links) == 0 {
		return nil
	}
	cur := ns.RenderedLink
	next := cur + 1
	if next >= len(links) {
		next = 0
	}
	ns.HighlightedLink = next
	ns.CursorLine = links[next].Line
	self.c.Helpers().PreviewNav().SyncCardIndexFromCursor()
	self.c.GuiCommon().RenderPreview()
	return nil
}

// HighlightPrevLink cycles to the previous link.
func (self *PreviewLinksHelper) HighlightPrevLink() error {
	self.ExtractLinks()
	ns := self.activeCtx().NavState()
	links := ns.Links
	if len(links) == 0 {
		return nil
	}
	cur := ns.RenderedLink
	prev := cur - 1
	if prev < 0 {
		prev = len(links) - 1
	}
	ns.HighlightedLink = prev
	ns.CursorLine = links[prev].Line
	self.c.Helpers().PreviewNav().SyncCardIndexFromCursor()
	self.c.GuiCommon().RenderPreview()
	return nil
}

// OpenLink opens the currently highlighted link.
func (self *PreviewLinksHelper) OpenLink() error {
	ns := self.activeCtx().NavState()
	links := ns.Links
	hl := ns.RenderedLink
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
		opts := self.c.Helpers().Preview().BuildSearchOptions()
		note, err := self.c.RuinCmd().Search.GetByTitle(target, opts)
		if err != nil || note == nil {
			return nil
		}
		nav := self.c.Helpers().PreviewNav()
		nav.PushNavHistory()
		// Wiki-links always open as a card list
		contexts := self.c.GuiCommon().Contexts()
		cl := contexts.CardList
		cl.Cards = []models.Note{*note}
		cl.SelectedCardIdx = 0
		ns := cl.NavState()
		ns.CursorLine = 1
		ns.ScrollOffset = 0
		contexts.ActivePreviewKey = "cardList"
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
