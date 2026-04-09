package helpers

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

type LinkHelper struct {
	c *HelperCommon
}

func NewLinkHelper(c *HelperCommon) *LinkHelper {
	return &LinkHelper{c: c}
}

func (self *LinkHelper) CreateLink(quickExit bool) error {
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
				if quickExit {
					return gocui.ErrQuit
				}
				return nil
			}
			return self.resolveAndCapture(url, tags, quickExit)
		},
		OnCtrlS: func(raw string) error {
			url, tags := self.parseURLAndTags(raw)
			if url == "" {
				return nil
			}
			return self.saveImmediate(url, tags, quickExit)
		},
	})
	return nil
}

// CreateLinkFromURL skips the input popup and resolves the given URL directly,
// opening the capture popup with the resolved title/summary on completion.
// The input popup is opened in a locked state to act as a spinner UI.
func (self *LinkHelper) CreateLinkFromURL(url string, quickExit bool) error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}

	cancelled := make(chan struct{})
	done := make(chan struct{})

	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title:      "Resolving",
		Locked:     true,
		DeferClose: true,
		OnCancel:   func() { close(cancelled) },
	})

	opts := linkResolveOpts{
		url:       url,
		quickExit: quickExit,
		done:      done,
		cancelled: cancelled,
	}

	go self.spinInputTitle(url, done)
	go self.doResolve(opts)

	return nil
}

// parseURLAndTags splits input like "https://example.com #tag1 #tag2" into URL and tag list.
func (self *LinkHelper) parseURLAndTags(raw string) (url string, tags []string) {
	fields := strings.FieldsSeq(raw)
	for f := range fields {
		if after, ok := strings.CutPrefix(f, "#"); ok {
			tag := after
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
// When quickExit is true, returns gocui.ErrQuit on success to terminate the app.
func (self *LinkHelper) saveImmediate(url string, tags []string, quickExit bool) error {
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
	if quickExit {
		return gocui.ErrQuit
	}
	self.c.Helpers().Preview().ReloadActivePreview()
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

// ReResolveLink re-resolves an existing link note's URL, preserving global tags and parent.
func (self *LinkHelper) ReResolveLink(note *models.Note) error {
	url := note.URL()
	if url == "" {
		return nil
	}

	// Preserve global tags without # prefix, excluding link (added automatically by ruin)
	var tags []string
	for _, t := range note.Tags {
		t = strings.TrimPrefix(t, "#")
		if t != "" && t != "link" {
			tags = append(tags, t)
		}
	}

	cancelled := make(chan struct{})
	done := make(chan struct{})

	self.c.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
		Title:      "Re-resolve Link",
		DeferClose: true,
		Locked:     true,
		OnCancel:   func() { close(cancelled) },
	})

	opts := linkResolveOpts{
		url:          url,
		tags:         tags,
		existingUUID: note.UUID,
		parent:       note.Parent,
		done:         done,
		cancelled:    cancelled,
	}

	go self.spinInputTitle(url, done)
	go self.doResolve(opts)

	return nil
}

// resolveAndCapture locks the input, resolves the URL, then opens capture with results.
func (self *LinkHelper) resolveAndCapture(url string, tags []string, quickExit bool) error {
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

	opts := linkResolveOpts{
		url:       url,
		tags:      tags,
		quickExit: quickExit,
		done:      done,
		cancelled: cancelled,
	}

	go self.spinInputTitle(url, done)
	go self.doResolve(opts)

	return nil
}

// linkResolveOpts holds parameters for the async resolve → capture flow.
type linkResolveOpts struct {
	url          string
	tags         []string
	existingUUID string // non-empty when re-resolving
	parent       string // parent UUID to preserve
	quickExit    bool   // when true, downstream actions should terminate the app
	done         chan struct{}
	cancelled    chan struct{}
}

func (self *LinkHelper) doResolve(opts linkResolveOpts) {
	result, err := self.c.RuinCmd().Link.Resolve(opts.url)
	close(opts.done)

	gui := self.c.GuiCommon()
	gui.Update(func() error {
		// If Esc was pressed while resolving, discard the result.
		select {
		case <-opts.cancelled:
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

		self.openCaptureWithResolved(opts, result)
		return nil
	})
}

func (self *LinkHelper) openCaptureWithResolved(opts linkResolveOpts, result *commands.LinkResolveResult) {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture
	ctx.Parent = nil
	ctx.Completion = types.NewCompletionState()
	ctx.LinkURL = opts.url
	ctx.LinkExistingUUID = opts.existingUUID
	ctx.LinkParent = opts.parent
	ctx.ResolveState = context.ResolveComplete
	ctx.ResolveResult = &context.LinkResolveResult{
		Title:   result.Title,
		Summary: result.Summary,
	}
	ctx.ResolveDone = nil
	ctx.LinkTags = opts.tags

	if opts.existingUUID != "" {
		ctx.LinkTitle = " Re-resolve Link "
	} else {
		ctx.LinkTitle = " New Link "
	}

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
		buf.WriteString(opts.url)
		if result.Summary != "" {
			fmt.Fprintf(&buf, "\n\n%s", result.Summary)
		}
		if len(opts.tags) > 0 {
			buf.WriteString("\n\n")
			for _, t := range opts.tags {
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

// SubmitLinkCapture finalizes the link capture and creates the note.
// When quickExit is true, returns gocui.ErrQuit on success to terminate the app.
func (self *LinkHelper) SubmitLinkCapture(content string, quickExit bool) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture

	url := ctx.LinkURL
	if url == "" {
		return nil
	}

	title, comment, contentTags := self.parseLinkContent(content, url)

	opts := commands.LinkNewOpts{
		Title:   title,
		Comment: comment,
		NoFetch: true,
	}
	// Content tags are authoritative — they reflect the user's edits in the
	// capture popup, which may differ from the original ctx.LinkTags.
	if len(contentTags) > 0 {
		opts.Tags = strings.Join(contentTags, ",")
	}
	if ctx.LinkParent != "" {
		opts.Parent = ctx.LinkParent
	}

	// Re-resolve: delete the old note before creating the replacement.
	if ctx.LinkExistingUUID != "" {
		if err := self.c.RuinCmd().Note.Delete(ctx.LinkExistingUUID); err != nil {
			gui.ShowError(fmt.Errorf("failed to delete old link note: %w", err))
			return nil
		}
	}

	_, err := self.c.RuinCmd().Link.New(url, opts)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	self.c.Helpers().Capture().CloseCapture()
	if quickExit {
		return gocui.ErrQuit
	}
	self.c.Helpers().Preview().ReloadActivePreview()
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

func (self *LinkHelper) parseLinkContent(content, url string) (title, comment string, tags []string) {
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
				for _, f := range strings.Fields(trimmed) {
					if after, ok := strings.CutPrefix(f, "#"); ok && after != "" {
						tags = append(tags, after)
					}
				}
				continue
			}
			commentLines = append(commentLines, trimmed)
		}
	}
	comment = strings.Join(commentLines, "\n")
	return title, comment, tags
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
