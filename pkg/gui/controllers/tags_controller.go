package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	helpers "kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// TagsController handles all Tags panel keybindings and behavior.
type TagsController struct {
	baseController
	*ListControllerTrait[models.Tag]
	c          *ControllerCommon
	getContext func() *context.TagsContext

	// Callbacks injected by gui wiring — these call back into gui methods
	// that handle cross-cutting concerns (preview updates, dialogs, etc.)
	// during the hybrid migration period.
	onFilterByTag func(tag *models.Tag) error
	onRenameTag   func(tag *models.Tag) error
	onDeleteTag   func(tag *models.Tag) error
}

var _ types.IController = &TagsController{}

// TagsControllerOpts holds the callbacks injected during wiring.
type TagsControllerOpts struct {
	Common     *ControllerCommon
	GetContext func() *context.TagsContext
	// Action callbacks — these delegate back to gui methods during hybrid period
	OnFilterByTag func(tag *models.Tag) error
	OnRenameTag   func(tag *models.Tag) error
	OnDeleteTag   func(tag *models.Tag) error
}

// NewTagsController creates a new TagsController.
func NewTagsController(opts TagsControllerOpts) *TagsController {
	ctrl := &TagsController{
		c:             opts.Common,
		getContext:    opts.GetContext,
		onFilterByTag: opts.OnFilterByTag,
		onRenameTag:   opts.OnRenameTag,
		onDeleteTag:   opts.OnDeleteTag,
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
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "tags",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("tags")
					if v == nil {
						return nil
					}
					ctx := self.getContext()
					items := ctx.FilteredItems()
					idx := helpers.ListClickIndex(v, 1)
					if idx >= 0 && idx < len(items) {
						ctx.SetSelectedLineIdx(idx)
					}
					self.c.GuiCommon().PushContext(ctx, types.OnFocusOpts{})
					return nil
				},
			},
			{
				ViewName: "tags",
				Key:      gocui.MouseWheelDown,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("tags")
					if v != nil {
						helpers.ScrollViewport(v, 3)
					}
					return nil
				},
			},
			{
				ViewName: "tags",
				Key:      gocui.MouseWheelUp,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					v := self.c.GuiCommon().GetView("tags")
					if v != nil {
						helpers.ScrollViewport(v, -3)
					}
					return nil
				},
			},
		}
	}
}

// Action handlers — delegate to injected callbacks.

func (self *TagsController) filterByTag(tag models.Tag) error {
	return self.onFilterByTag(&tag)
}

func (self *TagsController) renameTag(tag models.Tag) error {
	return self.onRenameTag(&tag)
}

func (self *TagsController) deleteTag(tag models.Tag) error {
	return self.onDeleteTag(&tag)
}
