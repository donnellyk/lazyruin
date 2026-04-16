package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/jesseduffield/gocui"
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
	return self.BuildPreviewBindings("datePreview",
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
		&types.Binding{
			ID: "datePreview.open_editor", Key: 'E',
			Handler: self.nav().OpenInEditor, Description: "Open in Editor", Category: "Preview",
			DisplayOnScreen: true, StatusBarLabel: "Editor",
		},
		&types.Binding{
			ID: "datePreview.edit_inline", Key: 'e',
			Handler: self.editInline, Description: "Edit in Popup", Category: "Preview",
		},
	)
}

func (self *DatePreviewController) editInline() error {
	card := self.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return nil
	}
	target := self.c.Helpers().PreviewLineOps().ResolveTarget()
	if target != nil && target.UUID == card.UUID {
		return self.c.Helpers().Capture().OpenCaptureForEditAtLine(card, target.LineNum)
	}
	return self.c.Helpers().Capture().OpenCaptureForEdit(card)
}

func (self *DatePreviewController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return self.NavMouseBindings()
}
