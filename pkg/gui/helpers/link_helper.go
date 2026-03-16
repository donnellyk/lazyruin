package helpers

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

type LinkHelper struct {
	c *HelperCommon
}

func NewLinkHelper(c *HelperCommon) *LinkHelper {
	return &LinkHelper{c: c}
}

func (self *LinkHelper) CreateLink() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}

	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title: "New Link",
		OnAccept: func(raw string, _ *types.CompletionItem) error {
			url := strings.TrimSpace(raw)
			if url == "" {
				return nil
			}
			return self.openLinkCapture(url)
		},
	})
	return nil
}

func (self *LinkHelper) openLinkCapture(url string) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture
	ctx.Parent = nil
	ctx.Completion = types.NewCompletionState()
	ctx.LinkURL = url
	ctx.LinkTitle = " New Link "
	ctx.ResolveState = context.ResolveInFlight
	ctx.ResolveResult = nil
	ctx.ResolveDone = make(chan struct{})

	gui.PushContextByKey("capture")

	// Seed the capture view with the URL after the view is created.
	gui.Update(func() error {
		v := gui.GetView("capture")
		if v == nil {
			return nil
		}
		v.TextArea.TypeString(url)
		return nil
	})

	done := ctx.ResolveDone

	go self.resolveURL(url, done)
	go self.spinResolve(url, done)

	return nil
}

func (self *LinkHelper) resolveURL(url string, done chan struct{}) {
	result, err := self.c.RuinCmd().Link.Resolve(url)
	gui := self.c.GuiCommon()
	gui.Update(func() error {
		defer close(done)
		ctx := gui.Contexts().Capture
		if ctx.LinkURL != url {
			return nil
		}
		if err != nil {
			ctx.ResolveState = context.ResolveFailed
		} else {
			ctx.ResolveState = context.ResolveComplete
			ctx.ResolveResult = &context.LinkResolveResult{
				Title:   result.Title,
				Summary: result.Summary,
			}
			self.populateResolvedContent(url, result.Title, result.Summary)
		}
		return nil
	})
}

func (self *LinkHelper) spinResolve(url string, done chan struct{}) {
	frames := []string{"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f"}
	tick := time.NewTicker(80 * time.Millisecond)
	defer tick.Stop()
	i := 0
	gui := self.c.GuiCommon()
	for {
		select {
		case <-tick.C:
			frame := frames[i%len(frames)]
			gui.Update(func() error {
				ctx := gui.Contexts().Capture
				if ctx.LinkURL != url {
					return nil
				}
				ctx.LinkTitle = fmt.Sprintf(" New Link %s ", frame)
				return nil
			})
			i++
		case <-done:
			gui.Update(func() error {
				ctx := gui.Contexts().Capture
				if ctx.LinkURL != url {
					return nil
				}
				if ctx.ResolveState == context.ResolveFailed {
					ctx.LinkTitle = " New Link (fetch failed) "
				} else {
					ctx.LinkTitle = " New Link "
				}
				return nil
			})
			return
		}
	}
}

func (self *LinkHelper) populateResolvedContent(url, title, summary string) {
	gui := self.c.GuiCommon()
	v := gui.GetView("capture")
	if v == nil {
		return
	}
	current := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	if current != url {
		return
	}
	// Clear existing content and replace with resolved template
	v.TextArea.Clear()
	v.Clear()
	var buf strings.Builder
	if title != "" {
		fmt.Fprintf(&buf, "# %s\n\n", title)
	}
	buf.WriteString(url)
	if summary != "" {
		fmt.Fprintf(&buf, "\n\n%s", summary)
	}
	v.TextArea.TypeString(buf.String())
}

func (self *LinkHelper) SubmitLinkCapture(content string) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture

	url := ctx.LinkURL
	if url == "" {
		return nil
	}

	title, comment := self.parseLinkContent(content, url)

	var opts commands.LinkNewOpts
	if ctx.ResolveState == context.ResolveComplete {
		opts = commands.LinkNewOpts{
			Title:   title,
			Comment: comment,
			NoFetch: true,
		}
	}
	// For InFlight/Failed, use empty opts so CLI fetches itself.

	_, err := self.c.RuinCmd().Link.New(url, opts)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.c.Helpers().Capture().CloseCapture()
	self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

func (self *LinkHelper) parseLinkContent(content, url string) (title, comment string) {
	lines := strings.Split(content, "\n")
	var commentLines []string
	pastURL := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && title == "" {
			title = strings.TrimPrefix(trimmed, "# ")
			continue
		}
		if trimmed == url {
			pastURL = true
			continue
		}
		if pastURL && trimmed != "" {
			commentLines = append(commentLines, trimmed)
		}
	}
	comment = strings.Join(commentLines, "\n")
	return title, comment
}

func (self *LinkHelper) BrowseLinks() error {
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	opts.Limit = 50
	notes, err := self.c.RuinCmd().Search.Search("#link", opts)
	if err != nil {
		return nil
	}
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList("Links", notes)
	self.c.GuiCommon().PushContextByKey("cardList")
	return nil
}

func (self *LinkHelper) OpenLinkURL(note *models.Note) error {
	url := note.URL()
	if url == "" {
		return nil
	}
	return exec.Command("open", url).Start()
}
