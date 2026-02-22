package controllers

import (
	"github.com/jesseduffield/gocui"
	"kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
)

// PreviewNavTrait provides shared navigation bindings for all preview modes
// (CardList, PickResults, Compose). Concrete preview controllers embed this
// and append their own mode-specific bindings.
type PreviewNavTrait struct {
	c *ControllerCommon
}

func (t *PreviewNavTrait) nav() *helpers.PreviewNavHelper     { return t.c.Helpers().PreviewNav() }
func (t *PreviewNavTrait) links() *helpers.PreviewLinksHelper { return t.c.Helpers().PreviewLinks() }
func (t *PreviewNavTrait) preview() *helpers.PreviewHelper    { return t.c.Helpers().Preview() }

// NavBindings returns the shared navigation keybindings.
func (t *PreviewNavTrait) NavBindings() []*types.Binding {
	return []*types.Binding{
		// Line navigation (no Description -> excluded from palette)
		{Key: 'j', Handler: t.nav().MoveDown},
		{Key: gocui.KeyArrowDown, Handler: t.nav().MoveDown},
		{Key: 'k', Handler: t.nav().MoveUp},
		{Key: gocui.KeyArrowUp, Handler: t.nav().MoveUp},
		// Card navigation
		{Key: 'J', Handler: t.nav().CardDown},
		{Key: 'K', Handler: t.nav().CardUp},
		// Header navigation
		{Key: '}', Handler: t.nav().NextHeader},
		{Key: '{', Handler: t.nav().PrevHeader},
		// Link navigation
		{Key: 'l', Handler: t.links().HighlightNextLink},
		{Key: 'L', Handler: t.links().HighlightPrevLink},
		// Link actions
		{
			ID:          "preview.open_link",
			Key:         'o',
			Handler:     t.links().OpenLink,
			Description: "Open Link",
			Category:    "Preview",
		},
		// Info
		{
			ID:          "preview.show_info",
			Key:         's',
			Handler:     t.c.Helpers().PreviewInfo().ShowInfoDialog,
			Description: "Show Info",
			Category:    "Note Actions",
		},
		// Display toggles
		{
			ID:          "preview.toggle_frontmatter",
			Key:         'f',
			Handler:     t.preview().ToggleFrontmatter,
			Description: "Toggle Frontmatter",
			Category:    "Preview",
		},
		{
			ID:          "preview.view_options",
			Key:         'v',
			Handler:     t.preview().ViewOptionsDialog,
			Description: "View Options",
			Category:    "Preview",
		},
		// Enter (dispatches per active preview mode)
		{
			ID:              "preview.enter",
			Key:             gocui.KeyEnter,
			Handler:         t.nav().PreviewEnter,
			Description:     "Enter",
			Category:        "Preview",
			DisplayOnScreen: true,
		},
		// Back / forward / history
		{
			ID:      "preview.back",
			Key:     gocui.KeyEsc,
			Handler: t.nav().Back,
		},
		{
			ID:          "preview.nav_back",
			Key:         '[',
			Handler:     t.nav().NavBack,
			Description: "Go Back",
			Category:    "Preview",
		},
		{
			ID:          "preview.nav_forward",
			Key:         ']',
			Handler:     t.nav().NavForward,
			Description: "Go Forward",
			Category:    "Preview",
		},
		// Palette-only display toggles
		{
			ID:          "preview.toggle_title",
			Handler:     t.preview().ToggleTitle,
			Description: "Toggle Title",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_global_tags",
			Handler:     t.preview().ToggleGlobalTags,
			Description: "Toggle Global Tags",
			Category:    "Preview",
		},
		{
			ID:          "preview.toggle_markdown",
			Handler:     t.preview().ToggleMarkdown,
			Description: "Toggle Markdown",
			Category:    "Preview",
		},
		{
			ID:          "preview.show_nav_history",
			Handler:     t.nav().ShowNavHistory,
			Description: "View History",
			Category:    "Preview",
		},
		// Pick dialog (inline pick without leaving preview)
		{
			ID:          "preview.pick_dialog",
			Key:         gocui.KeyCtrlP,
			Handler:     func() error { return t.c.Helpers().Pick().OpenPickDialog() },
			Description: "Pick (Dialog)",
			Category:    "Preview",
		},
	}
}

// LineOpsBindings returns the 4 standard line-level operation bindings
// (toggle todo, toggle #done, toggle inline tag, toggle inline date),
// parameterized by prefix for unique binding IDs.
func (t *PreviewNavTrait) LineOpsBindings(prefix string) []*types.Binding {
	lo := t.c.Helpers().PreviewLineOps()
	return []*types.Binding{
		{
			ID: prefix + ".toggle_todo", Key: 'x',
			Handler: lo.ToggleTodo, Description: "Toggle Todo", Category: "Preview",
		},
		{
			ID: prefix + ".append_done", Key: 'D',
			Handler: lo.AppendDone, Description: "Toggle #done", Category: "Preview",
		},
		{
			ID: prefix + ".toggle_inline_tag", Key: gocui.KeyCtrlT,
			Handler: lo.ToggleInlineTag, Description: "Toggle Inline Tag", Category: "Preview",
		},
		{
			ID: prefix + ".toggle_inline_date", Key: gocui.KeyCtrlD,
			Handler: lo.ToggleInlineDate, Description: "Toggle Inline Date", Category: "Preview",
		},
	}
}

// NavMouseBindings returns the shared mouse bindings.
func (t *PreviewNavTrait) NavMouseBindings() []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: "preview",
			Key:      gocui.MouseLeft,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return t.nav().Click()
			},
		},
		{
			ViewName: "",
			Key:      gocui.MouseWheelDown,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return t.nav().ScrollDown()
			},
		},
		{
			ViewName: "",
			Key:      gocui.MouseWheelUp,
			Handler: func(mopts gocui.ViewMouseBindingOpts) error {
				return t.nav().ScrollUp()
			},
		},
	}
}
