package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
)

type DatePreviewController struct {
	baseController
	PreviewNavTrait
	c          *ControllerCommon
	getContext func() *context.DatePreviewContext
}

var _ types.IController = &DatePreviewController{}

func NewDatePreviewController(c *ControllerCommon, getContext func() *context.DatePreviewContext) *DatePreviewController {
	return &DatePreviewController{
		PreviewNavTrait: PreviewNavTrait{c: c},
		c:               c,
		getContext:      getContext,
	}
}

func (self *DatePreviewController) Context() types.Context { return self.getContext() }

func (self *DatePreviewController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	bindings := self.NavBindings()
	bindings = append(bindings,
		&types.Binding{
			ID: "datePreview.next_section", Key: ')',
			Handler:     self.nav().NextSection,
			KeyDisplay:  ")/(",
			Description: "Next/prev section",
			Category:    "Navigation",
		},
		&types.Binding{
			ID:      "datePreview.prev_section",
			Key:     '(',
			Handler: self.nav().PrevSection,
		},
	)
	bindings = append(bindings, self.LineOpsBindings("datePreview")...)
	return bindings
}

func (self *DatePreviewController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
