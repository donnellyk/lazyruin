package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
)

// PreviewController handles all Preview panel keybindings.
// Action methods delegate to PreviewHelper.
type PreviewController struct {
	baseController
	c          *ControllerCommon
	getContext func() *context.PreviewContext
}

var _ types.IController = &PreviewController{}

// NewPreviewController creates a new PreviewController.
func NewPreviewController(c *ControllerCommon, getContext func() *context.PreviewContext) *PreviewController {
	return &PreviewController{
		c:          c,
		getContext: getContext,
	}
}

// Context returns the context this controller is attached to.
func (self *PreviewController) Context() types.Context {
	return self.getContext()
}

func (self *PreviewController) h() *helpers.Helpers {
	return self.c.Helpers().(*helpers.Helpers)
}

func (self *PreviewController) p() *helpers.PreviewHelper {
	return self.h().Preview()
}

// Note action wrappers — call helpers directly.
func (self *PreviewController) addTag() error         { return self.h().NoteActions().AddGlobalTag() }
func (self *PreviewController) removeTag() error      { return self.h().NoteActions().RemoveTag() }
func (self *PreviewController) setParent() error      { return self.h().NoteActions().SetParentDialog() }
func (self *PreviewController) removeParent() error   { return self.h().NoteActions().RemoveParent() }
func (self *PreviewController) toggleBookmark() error { return self.h().NoteActions().ToggleBookmark() }

// GetKeybindingsFn returns the keybinding producer for preview.
// Bindings with Key == nil are palette-only (no keybinding registered).
func (self *PreviewController) GetKeybindingsFn() types.KeybindingsFn {
	return func(opts types.KeybindingsOpts) []*types.Binding {
		return []*types.Binding{
			// Navigation (no Description -> excluded from palette)
			{Key: 'j', Handler: self.p().MoveDown},
			{Key: gocui.KeyArrowDown, Handler: self.p().MoveDown},
			{Key: 'k', Handler: self.p().MoveUp},
			{Key: gocui.KeyArrowUp, Handler: self.p().MoveUp},
			{Key: 'J', Handler: self.p().CardDown},
			{Key: 'K', Handler: self.p().CardUp},
			{Key: '}', Handler: self.p().NextHeader},
			{Key: '{', Handler: self.p().PrevHeader},
			{Key: 'l', Handler: self.p().HighlightNextLink},
			{Key: 'L', Handler: self.p().HighlightPrevLink},

			// Card actions (Description -> shown in palette)
			{
				ID:          "preview.delete",
				Key:         'd',
				Handler:     self.p().DeleteCard,
				Description: "Delete Card",
				Category:    "Preview",
			},
			{
				ID:              "preview.open_editor",
				Key:             'E',
				Handler:         self.p().OpenInEditor,
				Description:     "Open in Editor",
				Category:        "Preview",
				DisplayOnScreen: true,
			},
			{
				ID:          "preview.append_done",
				Key:         'D',
				Handler:     self.p().AppendDone,
				Description: "Toggle #done",
				Category:    "Preview",
			},
			{
				ID:          "preview.move_card",
				Key:         'm',
				Handler:     self.p().MoveCardDialog,
				Description: "Move Card",
				Category:    "Preview",
			},
			{
				ID:          "preview.merge_card",
				Key:         'M',
				Handler:     self.p().MergeCardDialog,
				Description: "Merge Notes",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_frontmatter",
				Key:         'f',
				Handler:     self.p().ToggleFrontmatter,
				Description: "Toggle Frontmatter",
				Category:    "Preview",
			},
			{
				ID:          "preview.view_options",
				Key:         'v',
				Handler:     self.p().ViewOptionsDialog,
				Description: "View Options",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_inline_tag",
				Key:         gocui.KeyCtrlT,
				Handler:     self.p().ToggleInlineTag,
				Description: "Toggle Inline Tag",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_inline_date",
				Key:         gocui.KeyCtrlD,
				Handler:     self.p().ToggleInlineDate,
				Description: "Toggle Inline Date",
				Category:    "Preview",
			},
			{
				ID:          "preview.open_link",
				Key:         'o',
				Handler:     self.p().OpenLink,
				Description: "Open Link",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_todo",
				Key:         'x',
				Handler:     self.p().ToggleTodo,
				Description: "Toggle Todo",
				Category:    "Preview",
			},
			{
				ID:              "preview.focus_note",
				Key:             gocui.KeyEnter,
				Handler:         self.p().FocusNote,
				Description:     "Focus Note from Preview",
				Category:        "Preview",
				DisplayOnScreen: true,
			},
			// Back (no Description -> excluded from palette)
			{
				ID:      "preview.back",
				Key:     gocui.KeyEsc,
				Handler: self.p().Back,
			},
			{
				ID:          "preview.nav_back",
				Key:         '[',
				Handler:     self.p().NavBack,
				Description: "Go Back",
				Category:    "Preview",
			},
			{
				ID:          "preview.nav_forward",
				Key:         ']',
				Handler:     self.p().NavForward,
				Description: "Go Forward",
				Category:    "Preview",
			},

			// Note actions (shared with Notes panel) -- call helpers directly
			{
				ID:              "preview.add_tag",
				Key:             't',
				Handler:         self.addTag,
				Description:     "Add Tag",
				Category:        "Note Actions",
				DisplayOnScreen: true,
			},
			{
				ID:          "preview.remove_tag",
				Key:         'T',
				Handler:     self.removeTag,
				Description: "Remove Tag",
				Category:    "Note Actions",
			},
			{
				ID:          "preview.set_parent",
				Key:         '>',
				Handler:     self.setParent,
				Description: "Set Parent",
				Category:    "Note Actions",
			},
			{
				ID:          "preview.remove_parent",
				Key:         'P',
				Handler:     self.removeParent,
				Description: "Remove Parent",
				Category:    "Note Actions",
			},
			{
				ID:          "preview.toggle_bookmark",
				Key:         'b',
				Handler:     self.toggleBookmark,
				Description: "Toggle Bookmark",
				Category:    "Note Actions",
			},
			{
				ID:          "preview.show_info",
				Key:         's',
				Handler:     self.p().ShowInfoDialog,
				Description: "Show Info",
				Category:    "Note Actions",
			},

			// Palette-only commands (Key == nil -> no keybinding, palette entry only)
			{
				ID:          "preview.toggle_title",
				Handler:     self.p().ToggleTitle,
				Description: "Toggle Title",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_global_tags",
				Handler:     self.p().ToggleGlobalTags,
				Description: "Toggle Global Tags",
				Category:    "Preview",
			},
			{
				ID:          "preview.toggle_markdown",
				Handler:     self.p().ToggleMarkdown,
				Description: "Toggle Markdown",
				Category:    "Preview",
			},
			{
				ID:          "preview.order_cards",
				Handler:     self.p().OrderCards,
				Description: "Order Cards",
				Category:    "Preview",
			},
			{
				ID:          "preview.show_nav_history",
				Handler:     self.p().ShowNavHistory,
				Description: "View History",
				Category:    "Preview",
			},
		}
	}
}

// GetMouseKeybindingsFn returns mouse bindings for the preview panel.
func (self *PreviewController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
	return func(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
		return []*gocui.ViewMouseBinding{
			{
				ViewName: "preview",
				Key:      gocui.MouseLeft,
				Handler: func(mopts gocui.ViewMouseBindingOpts) error {
					return self.p().Click()
				},
			},
		}
	}
}
