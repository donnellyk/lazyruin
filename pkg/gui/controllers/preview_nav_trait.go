package controllers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/jesseduffield/gocui"
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
		{Key: 'j', Handler: t.nav().MoveDown, KeyDisplay: "j/k", Description: "Scroll line-by-line", Category: "Navigation"},
		{Key: gocui.KeyArrowDown, Handler: t.nav().MoveDown},
		{Key: 'k', Handler: t.nav().MoveUp},
		{Key: gocui.KeyArrowUp, Handler: t.nav().MoveUp},
		{Key: 'J', Handler: t.nav().CardDown, KeyDisplay: "J/K", Description: "Jump between cards", Category: "Navigation"},
		{Key: 'K', Handler: t.nav().CardUp},
		{Key: '}', Handler: t.nav().NextHeader, KeyDisplay: "]/[", Description: "Next/prev header", Category: "Navigation"},
		{Key: '{', Handler: t.nav().PrevHeader},
		{Key: 'l', Handler: t.links().HighlightNextLink, KeyDisplay: "l/L", Description: "Next/prev link", Category: "Navigation"},
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
			ID:              "preview.view_options",
			Key:             'v',
			Handler:         t.preview().ViewOptionsDialog,
			Description:     "View Options",
			Category:        "Preview",
			DisplayOnScreen: true,
			StatusBarLabel:  "View",
		},
		// Enter (dispatches per active preview mode)
		{
			ID:              "preview.enter",
			Key:             gocui.KeyEnter,
			Handler:         t.nav().PreviewEnter,
			Description:     "Enter",
			Category:        "Preview",
			DisplayOnScreen: true,
			StatusBarLabel:  "Focus",
		},
		// Back / forward / history
		{
			ID:              "preview.back",
			Key:             gocui.KeyEsc,
			Handler:         t.nav().Back,
			Description:     "Back",
			Category:        "Preview",
			DisplayOnScreen: true,
			StatusBarLabel:  "Back",
		},
		// Palette-only display toggles
		{
			ID:          "preview.toggle_dim_done",
			Handler:     t.preview().ToggleDimDone,
			Description: "Toggle Dim Done",
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

// BuildPreviewBindings assembles the standard preview binding set: shared nav
// bindings, line-ops (parameterized by prefix), and any custom bindings.
func (t *PreviewNavTrait) BuildPreviewBindings(lineOpsPrefix string, custom ...*types.Binding) []*types.Binding {
	bindings := t.NavBindings()
	bindings = append(bindings, t.LineOpsBindings(lineOpsPrefix)...)
	bindings = append(bindings, custom...)
	return bindings
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
			DisplayOnScreen: true, StatusBarLabel: "Todo",
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
