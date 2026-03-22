package controllers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

type InboxBrowserController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.InboxBrowserContext
}

var _ types.IController = &InboxBrowserController{}

func NewInboxBrowserController(
	c *ControllerCommon,
	getContext func() *context.InboxBrowserContext,
) *InboxBrowserController {
	return &InboxBrowserController{
		c:          c,
		getContext: getContext,
	}
}

func (self *InboxBrowserController) Context() types.Context {
	return self.getContext()
}

func (self *InboxBrowserController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{Key: 'j', Handler: self.nextItem},
		{Key: 'k', Handler: self.prevItem},
		{Key: gocui.KeyArrowDown, Handler: self.nextItem},
		{Key: gocui.KeyArrowUp, Handler: self.prevItem},
		{Key: gocui.KeyEnter, Description: "Promote", Handler: self.promoteItem},
		{Key: 'd', Description: "Delete", Handler: self.deleteItem},
		{Key: gocui.KeyEsc, Description: "Close", Handler: self.close},
	}
}

func (self *InboxBrowserController) nextItem() error {
	ctx := self.getContext()
	if ctx.SelectedIdx < len(ctx.Items)-1 {
		ctx.SelectedIdx++
	}
	return nil
}

func (self *InboxBrowserController) prevItem() error {
	ctx := self.getContext()
	if ctx.SelectedIdx > 0 {
		ctx.SelectedIdx--
	}
	return nil
}

func (self *InboxBrowserController) promoteItem() error {
	ctx := self.getContext()
	if len(ctx.Items) == 0 {
		return nil
	}
	item := ctx.Items[ctx.SelectedIdx]
	if ctx.OnSelect != nil {
		self.c.GuiCommon().PopContext()
		return ctx.OnSelect(item)
	}
	self.c.Helpers().Inbox().DeleteItem(item.ID)
	self.c.GuiCommon().PopContext()
	return self.c.Helpers().Capture().OpenCaptureWithContent(item.Text)
}

func (self *InboxBrowserController) deleteItem() error {
	ctx := self.getContext()
	if len(ctx.Items) == 0 {
		return nil
	}
	item := ctx.Items[ctx.SelectedIdx]
	self.c.GuiCommon().ShowConfirm("Delete inbox item?", item.Text, func() error {
		self.c.Helpers().Inbox().DeleteItem(item.ID)
		self.c.Helpers().Inbox().RefreshBrowser()
		return nil
	})
	return nil
}

func (self *InboxBrowserController) close() error {
	self.c.GuiCommon().PopContext()
	return nil
}
