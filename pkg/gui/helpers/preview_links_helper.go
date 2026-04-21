package helpers

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/models"
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

// ExtractLinks parses the preview content for wiki-links and URLs.
//
// Source lines may be broken across several consecutive ns.Lines entries
// by our internal wordwrap (for example at hyphens inside a URL). The
// simple "regex each visual line" approach would capture only the prefix
// before the break. Here we instead reconstruct each source line from its
// contiguous ns.Lines group, match against the reconstruction, then emit
// one PreviewLink per match — possibly with multiple Segments, one per
// visual line it spans.
func (self *PreviewLinksHelper) ExtractLinks() {
	ns := self.activeCtx().NavState()
	ns.Links = nil

	wikiRe := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	urlRe := regexp.MustCompile(`https?://[^\s)\]>]+`)

	// Walk consecutive groups of ns.Lines that share the same source
	// identity (UUID + LineNum). Lines with LineNum == 0 are synthetic
	// (separators, blanks) and don't carry links.
	i := 0
	for i < len(ns.Lines) {
		if ns.Lines[i].LineNum == 0 {
			i++
			continue
		}
		start := i
		uuid := ns.Lines[i].UUID
		ln := ns.Lines[i].LineNum
		for i < len(ns.Lines) && ns.Lines[i].UUID == uuid && ns.Lines[i].LineNum == ln {
			i++
		}
		end := i

		// Per-segment plain texts. The renderer prefixes each visual line
		// with a single leading space; strip it so reconstructed source
		// columns align with what the regex expects.
		segs := make([]string, end-start)
		for j := start; j < end; j++ {
			plain := stripAnsi(ns.Lines[j].Text)
			segs[j-start] = strings.TrimPrefix(plain, " ")
		}
		source := strings.Join(segs, "")

		for _, match := range wikiRe.FindAllStringIndex(source, -1) {
			ns.Links = append(ns.Links,
				buildLink(source[match[0]:match[1]], start, match[0], match[1], segs))
		}
		for _, match := range urlRe.FindAllStringIndex(source, -1) {
			ns.Links = append(ns.Links,
				buildLink(source[match[0]:match[1]], start, match[0], match[1], segs))
		}
	}
}

// buildLink maps a [matchStart, matchEnd) byte range in the reconstructed
// source string to one or more on-screen segments. Each segment carries
// the ns.Lines index (start + segment offset), the visible column within
// that line (+1 for the leading space the renderer adds), and the visible
// length it covers.
func buildLink(text string, groupStart, matchStart, matchEnd int, segs []string) context.PreviewLink {
	var segments []context.PreviewLinkSegment
	offset := 0
	for j, s := range segs {
		segLen := len(s)
		segStart := offset
		segEnd := offset + segLen
		if matchStart < segEnd && matchEnd > segStart {
			col := matchStart - segStart
			if col < 0 {
				col = 0
			}
			end := matchEnd - segStart
			if end > segLen {
				end = segLen
			}
			segments = append(segments, context.PreviewLinkSegment{
				Line: groupStart + j,
				Col:  col + 1, // +1 for the renderer's leading space
				Len:  end - col,
			})
		}
		offset = segEnd
	}
	return context.PreviewLink{Text: text, Segments: segments}
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
	if len(links[next].Segments) > 0 {
		ns.CursorLine = links[next].Segments[0].Line
	}
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
	if len(links[prev].Segments) > 0 {
		ns.CursorLine = links[prev].Segments[0].Line
	}
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
		noteCopy := *note
		title := displayTitleForNote(noteCopy.Title)
		return self.c.Helpers().Navigator().NavigateTo("cardList", title, func() error {
			source := self.c.Helpers().Preview().NewSingleNoteSource(noteCopy.UUID)
			self.c.Helpers().Preview().ShowCardList(title, []models.Note{noteCopy}, source)
			return nil
		})
	}

	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		exec.Command("open", text).Start()
	}
	return nil
}
