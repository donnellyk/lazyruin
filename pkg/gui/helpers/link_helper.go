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
		Title:      "New Link",
		Footer:     " Enter: Resolve URL | <c-s>: Save ",
		DeferClose: true,
		Triggers: func() []types.CompletionTrigger {
			return []types.CompletionTrigger{
				{Prefix: "#", Candidates: self.c.Helpers().Completion().TagCandidates},
			}
		},
		OnAccept: func(raw string, _ *types.CompletionItem) error {
			url, tags := self.parseURLAndTags(raw)
			if url == "" {
				self.c.Helpers().InputPopup().CloseInputPopup()
				return nil
			}
			return self.resolveAndCapture(url, tags)
		},
		OnCtrlS: func(raw string) error {
			url, tags := self.parseURLAndTags(raw)
			if url == "" {
				return nil
			}
			return self.saveImmediate(url, tags)
		},
	})
	return nil
}

// parseURLAndTags splits input like "https://example.com #tag1 #tag2" into URL and tag list.
func (self *LinkHelper) parseURLAndTags(raw string) (url string, tags []string) {
	fields := strings.Fields(raw)
	for _, f := range fields {
		if strings.HasPrefix(f, "#") {
			tag := strings.TrimPrefix(f, "#")
			if tag != "" {
				tags = append(tags, tag)
			}
		} else if url == "" {
			url = f
		}
	}
	return url, tags
}

// saveImmediate creates a link with --no-fetch and optional tags, then closes.
func (self *LinkHelper) saveImmediate(url string, tags []string) error {
	opts := commands.LinkNewOpts{NoFetch: true}
	if len(tags) > 0 {
		opts.Tags = strings.Join(tags, ",")
	}

	_, err := self.c.RuinCmd().Link.New(url, opts)
	if err != nil {
		self.c.GuiCommon().ShowError(err)
		return nil
	}

	self.c.Helpers().InputPopup().CloseInputPopup()
	self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

// resolveAndCapture locks the input, resolves the URL, then opens capture with results.
func (self *LinkHelper) resolveAndCapture(url string, tags []string) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().InputPopup

	cancelled := make(chan struct{})
	done := make(chan struct{})

	// Lock the input popup and wire up Esc to cancel
	if ctx.Config != nil {
		ctx.Config.Locked = true
		ctx.Config.Footer = ""
		ctx.Config.OnCancel = func() { close(cancelled) }
	}

	go self.spinInputTitle(url, done)
	go self.doResolve(url, tags, done, cancelled)

	return nil
}

func (self *LinkHelper) doResolve(url string, tags []string, done, cancelled chan struct{}) {
	result, err := self.c.RuinCmd().Link.Resolve(url)
	close(done)

	gui := self.c.GuiCommon()
	gui.Update(func() error {
		// If Esc was pressed while resolving, discard the result.
		select {
		case <-cancelled:
			return nil
		default:
		}

		ctx := gui.Contexts().InputPopup
		if ctx.Config == nil || !ctx.Config.Locked {
			return nil
		}

		self.c.Helpers().InputPopup().CloseInputPopup()

		if err != nil {
			gui.ShowError(fmt.Errorf("link resolve failed: %w", err))
			return nil
		}

		self.openCaptureWithResolved(url, tags, result)
		return nil
	})
}

func (self *LinkHelper) openCaptureWithResolved(url string, tags []string, result *commands.LinkResolveResult) {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture
	ctx.Parent = nil
	ctx.Completion = types.NewCompletionState()
	ctx.LinkURL = url
	ctx.LinkTitle = " New Link "
	ctx.ResolveState = context.ResolveComplete
	ctx.ResolveResult = &context.LinkResolveResult{
		Title:   result.Title,
		Summary: result.Summary,
	}
	ctx.ResolveDone = nil
	ctx.LinkTags = tags

	gui.PushContextByKey("capture")

	gui.Update(func() error {
		v := gui.GetView("capture")
		if v == nil {
			return nil
		}
		var buf strings.Builder
		if result.Title != "" {
			fmt.Fprintf(&buf, "# %s\n\n", result.Title)
		}
		buf.WriteString(url)
		if result.Summary != "" {
			fmt.Fprintf(&buf, "\n\n%s", result.Summary)
		}
		if len(tags) > 0 {
			buf.WriteString("\n\n")
			for _, t := range tags {
				buf.WriteString("#" + t + " ")
			}
		}
		v.TextArea.TypeString(buf.String())
		return nil
	})
}

func (self *LinkHelper) spinInputTitle(_ string, done chan struct{}) {
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
				ctx := gui.Contexts().InputPopup
				if ctx.Config == nil || !ctx.Config.Locked {
					return nil
				}
				ctx.Config.Title = fmt.Sprintf("Resolving %s", frame)
				return nil
			})
			i++
		case <-done:
			return
		}
	}
}

func (self *LinkHelper) SubmitLinkCapture(content string) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture

	url := ctx.LinkURL
	if url == "" {
		return nil
	}

	title, comment := self.parseLinkContent(content, url)

	opts := commands.LinkNewOpts{
		Title:   title,
		Comment: comment,
		NoFetch: true,
	}
	if len(ctx.LinkTags) > 0 {
		opts.Tags = strings.Join(ctx.LinkTags, ",")
	}

	_, err := self.c.RuinCmd().Link.New(url, opts)
	if err != nil {
		gui.ShowError(err)
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
			if isTagLine(trimmed) {
				continue
			}
			commentLines = append(commentLines, trimmed)
		}
	}
	comment = strings.Join(commentLines, "\n")
	return title, comment
}

// isTagLine returns true if every non-empty token on the line starts with #.
func isTagLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	for _, f := range fields {
		if !strings.HasPrefix(f, "#") {
			return false
		}
	}
	return true
}

func (self *LinkHelper) BrowseLinks() error {
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	opts.Limit = 50
	opts.Link = true
	notes, err := self.c.RuinCmd().Search.Search("", opts)
	if err != nil {
		return nil
	}

	source := context.CardListSource{
		Query: "",
		Requery: func(filterText string) ([]models.Note, error) {
			o := self.c.Helpers().Preview().BuildSearchOptions()
			o.Limit = 50
			o.Link = true
			return self.c.RuinCmd().Search.Search(strings.TrimSpace(filterText), o)
		},
	}

	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList("Links", notes, source)
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
