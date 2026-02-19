package helpers

import (
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

// CaptureHelper encapsulates the capture popup logic.
type CaptureHelper struct {
	c *HelperCommon
}

func NewCaptureHelper(c *HelperCommon) *CaptureHelper {
	return &CaptureHelper{c: c}
}

// OpenCapture opens the capture popup, resetting state.
func (self *CaptureHelper) OpenCapture() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	ctx := self.c.GuiCommon().Contexts().Capture
	ctx.Parent = nil
	ctx.Completion = types.NewCompletionState()
	gui.PushContextByKey("capture")
	return nil
}

// SubmitCapture submits the capture content and closes the popup.
func (self *CaptureHelper) SubmitCapture(content string, quickCapture bool) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture

	if content == "" {
		if quickCapture {
			return gocui.ErrQuit
		}
		return self.CloseCapture()
	}

	args := []string{"log", content}
	if ctx.Parent != nil {
		args = append(args, "--parent", ctx.Parent.UUID)
	}
	_, err := self.c.RuinCmd().Execute(args...)
	if err != nil {
		if quickCapture {
			return gocui.ErrQuit
		}
		return self.CloseCapture()
	}

	if quickCapture {
		return gocui.ErrQuit
	}

	self.CloseCapture()
	self.c.Helpers().Notes().FetchNotesForCurrentTab(false)
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

// CancelCapture cancels the capture, dismissing completion first if active.
func (self *CaptureHelper) CancelCapture(quickCapture bool) error {
	ctx := self.c.GuiCommon().Contexts().Capture
	if ctx.Completion.Active {
		ctx.Completion.Dismiss()
		return nil
	}
	if quickCapture {
		return gocui.ErrQuit
	}
	return self.CloseCapture()
}

// CloseCapture resets capture state and pops the context.
func (self *CaptureHelper) CloseCapture() error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture
	ctx.Parent = nil
	ctx.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)
	gui.PopContext()
	return nil
}
