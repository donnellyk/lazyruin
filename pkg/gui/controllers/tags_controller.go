package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// TagsController handles all Tags panel keybindings and behavior.
type TagsController struct {
	baseController
	*ListControllerTrait[models.Tag]
	c          *ControllerCommon
	getContext func() *context.TagsContext
}

var _ types.IController = &TagsController{}

// TagsControllerOpts holds dependencies for construction.
type TagsControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.TagsContext
}

// NewTagsController creates a new TagsController.
func NewTagsController(opts TagsControllerOpts) *TagsController {
	ctrl := &TagsController{
		c:          opts.Common,
		getContext: opts.GetContext,
	}

	ctrl.ListControllerTrait = NewListControllerTrait[models.Tag](
		opts.Common,
		func() types.IListContext { return opts.GetContext() },
		func() []models.Tag { return opts.GetContext().FilteredItems() },
		func() *context.ListContextTrait { return opts.GetContext().ListContextTrait },
	)

	return ctrl
}

// Context returns the context this controller is attached to.
func (self *TagsController) Context() types.Context {
	return self.getContext()
}

// GetKeybindingsFn returns the keybinding producer for tags.
func (self *TagsController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		bindings := []*types.Binding{
			// Actions (have Description → shown in palette)
			{
				ID:                "tags.filter",
				Key:               gocui.KeyEnter,
				Handler:           self.withItem(self.filterByTag),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Filter by Tag",
				Category:          "Tags",
				DisplayOnScreen:   true,
			},
			{
				ID:                "tags.rename",
				Key:               'r',
				Handler:           self.withItem(self.renameTag),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Rename Tag",
				Category:          "Tags",
				DisplayOnScreen:   true,
			},
			{
				ID:                "tags.delete",
				Key:               'd',
				Handler:           self.withItem(self.deleteTag),
				GetDisabledReason: self.require(self.singleItemSelected()),
				Description:       "Delete Tag",
				Category:          "Tags",
				DisplayOnScreen:   true,
			},
		}
		// Navigation bindings (no Description → excluded from palette)
		bindings = append(bindings, self.NavBindings()...)
		return bindings
	}
}

// GetMouseKeybindingsFn returns mouse bindings for the tags panel.
func (self *TagsController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
	return ListMouseBindings(ListMouseOpts{
		ViewName:     "tags",
		ClickMargin:  1,
		ItemCount:    func() int { return len(self.getContext().FilteredItems()) },
		SetSelection: func(idx int) { self.getContext().SetSelectedLineIdx(idx) },
		GetContext:   func() types.Context { return self.getContext() },
		GuiCommon:    func() IGuiCommon { return self.c.GuiCommon() },
	})
}

// Action handlers — call helpers directly.

func (self *TagsController) filterByTag(tag models.Tag) error {
	return self.c.Helpers().Tags().FilterByTag(&tag)
}

func (self *TagsController) renameTag(tag models.Tag) error {
	return self.c.Helpers().Tags().RenameTag(&tag)
}

func (self *TagsController) deleteTag(tag models.Tag) error {
	return self.c.Helpers().Tags().DeleteTag(&tag)
}
