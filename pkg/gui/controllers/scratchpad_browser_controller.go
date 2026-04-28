package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"

	"github.com/jesseduffield/gocui"
)

type ScratchpadBrowserController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.ScratchpadBrowserContext
}

var _ types.IController = &ScratchpadBrowserController{}

func NewScratchpadBrowserController(
	c *ControllerCommon,
	getContext func() *context.ScratchpadBrowserContext,
) *ScratchpadBrowserController {
	return &ScratchpadBrowserController{
		c:          c,
		getContext: getContext,
	}
}

func (self *ScratchpadBrowserController) Context() types.Context {
	return self.getContext()
}

func (self *ScratchpadBrowserController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return WheelScrollBindings("scratchpadBrowser", func() IGuiCommon { return self.c.GuiCommon() })
}

func (self *ScratchpadBrowserController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
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

func (self *ScratchpadBrowserController) nextItem() error {
	ctx := self.getContext()
	if ctx.SelectedIdx < len(ctx.Items)-1 {
		ctx.SelectedIdx++
	}
	return nil
}

func (self *ScratchpadBrowserController) prevItem() error {
	ctx := self.getContext()
	if ctx.SelectedIdx > 0 {
		ctx.SelectedIdx--
	}
	return nil
}

func (self *ScratchpadBrowserController) promoteItem() error {
	ctx := self.getContext()
	if len(ctx.Items) == 0 {
		return nil
	}
	item := ctx.Items[ctx.SelectedIdx]
	if ctx.OnSelect != nil {
		self.c.GuiCommon().PopContext()
		return ctx.OnSelect(item)
	}
	self.c.Helpers().Scratchpad().DeleteItem(item.ID)
	self.c.GuiCommon().PopContext()
	return self.c.Helpers().Capture().OpenCaptureWithContent(item.Text)
}

func (self *ScratchpadBrowserController) deleteItem() error {
	ctx := self.getContext()
	if len(ctx.Items) == 0 {
		return nil
	}
	item := ctx.Items[ctx.SelectedIdx]
	self.c.GuiCommon().ShowConfirm("Delete scratchpad item?", item.Text, func() error {
		self.c.Helpers().Scratchpad().DeleteItem(item.ID)
		self.c.Helpers().Scratchpad().RefreshBrowser()
		return nil
	})
	return nil
}

func (self *ScratchpadBrowserController) close() error {
	self.c.GuiCommon().PopContext()
	return nil
}
