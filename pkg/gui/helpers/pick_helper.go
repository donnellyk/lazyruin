package helpers

import (
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// PickHelper encapsulates the pick popup logic.
type PickHelper struct {
	c *HelperCommon
}

func NewPickHelper(c *HelperCommon) *PickHelper {
	return &PickHelper{c: c}
}

// OpenPick opens the pick popup, resetting state.
func (self *PickHelper) OpenPick() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	ctx := gui.Contexts().Pick
	ctx.Completion = types.NewCompletionState()
	ctx.AnyMode = false
	ctx.SeedHash = true
	ctx.DialogMode = false
	gui.PushContextByKey("pick")
	return nil
}

// OpenPickDialog opens the pick popup in dialog mode (results appear in an overlay).
func (self *PickHelper) OpenPickDialog() error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}

	var scopeTitle string
	switch gui.Contexts().ActivePreviewKey {
	case "compose":
		scopeTitle = gui.Contexts().Compose.ParentTitle
	case "cardList":
		cl := gui.Contexts().CardList
		if cl.SelectedCardIdx < len(cl.Cards) {
			scopeTitle = cl.Cards[cl.SelectedCardIdx].Title
		}
	}

	ctx := gui.Contexts().Pick
	ctx.Completion = types.NewCompletionState()
	ctx.AnyMode = false
	ctx.SeedHash = true
	ctx.DialogMode = true
	ctx.ScopeTitle = scopeTitle
	gui.PushContextByKey("pick")
	return nil
}

// ParsePickQuery splits raw pick input into tags, an optional @date
// (line-level date filter), and remaining filter text.
func ParsePickQuery(raw string) (tags []string, date string, filter string) {
	for _, token := range strings.Fields(raw) {
		if strings.HasPrefix(token, "@") {
			date = token
		} else {
			if !strings.HasPrefix(token, "#") {
				token = "#" + token
			}
			tags = append(tags, token)
		}
	}
	return tags, date, ""
}

// ExecutePick parses the raw input, runs the pick command, and shows results.
func (self *PickHelper) ExecutePick(raw string) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Pick

	if raw == "" {
		return self.CancelPick()
	}

	if ctx.DialogMode {
		ctx.DialogMode = false
		return self.executePickDialog(raw, ctx)
	}

	tags, date, filter := ParsePickQuery(raw)
	results, err := self.c.RuinCmd().Pick.Pick(tags, commands.PickOpts{Any: ctx.AnyMode, Date: date, Filter: filter})

	// Always close the pick dialog
	ctx.Query = raw
	ctx.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)

	if err != nil {
		results = nil
	}

	self.c.Helpers().Preview().ShowPickResults("Pick: "+raw, results)
	gui.ReplaceContextByKey("pickResults")
	return nil
}

// scopedPickOpts builds PickOpts with context-appropriate scoping:
// compose mode scopes to the parent's children, cardList mode scopes to
// the selected note, and all other modes are unscoped.
func (self *PickHelper) scopedPickOpts(date, filter string, anyMode bool) commands.PickOpts {
	opts := commands.PickOpts{Any: anyMode, Date: date, Filter: filter}
	gui := self.c.GuiCommon()
	switch gui.Contexts().ActivePreviewKey {
	case "compose":
		comp := gui.Contexts().Compose
		if comp.ParentUUID != "" {
			opts.Parent = comp.ParentUUID
		}
	case "cardList":
		cl := gui.Contexts().CardList
		if cl.SelectedCardIdx < len(cl.Cards) {
			opts.Notes = []string{cl.Cards[cl.SelectedCardIdx].UUID}
		}
	}
	return opts
}

// executePickDialog runs a pick and shows results in the dialog overlay.
// In compose mode, results are scoped to the parent's children via --parent.
// In cardList mode, results are scoped to the selected note via --notes.
// Otherwise, all results are shown.
func (self *PickHelper) executePickDialog(raw string, ctx *context.PickContext) error {
	gui := self.c.GuiCommon()

	tags, date, filter := ParsePickQuery(raw)

	ctx.Query = raw
	ctx.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)

	opts := self.scopedPickOpts(date, filter, ctx.AnyMode)

	var results []models.PickResult
	res, err := self.c.RuinCmd().Pick.Pick(tags, opts)
	if err == nil {
		results = res
	}

	pd := gui.Contexts().PickDialog
	pd.Results = results
	pd.SelectedCardIdx = 0
	pd.CursorLine = 1
	pd.ScrollOffset = 0
	pd.Query = raw
	pd.ScopeTitle = ctx.ScopeTitle
	gui.ReplaceContextByKey("pickDialog")
	return nil
}

// CancelPick closes the pick popup without executing.
func (self *PickHelper) CancelPick() error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Pick
	ctx.DialogMode = false
	ctx.ScopeTitle = ""
	ctx.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)
	gui.PopContext()
	return nil
}

// TogglePickAny toggles the any-mode flag for pick.
func (self *PickHelper) TogglePickAny() {
	ctx := self.c.GuiCommon().Contexts().Pick
	ctx.AnyMode = !ctx.AnyMode
}

// ReloadPickDialog re-runs the current pick dialog query and updates results.
func (self *PickHelper) ReloadPickDialog() {
	gui := self.c.GuiCommon()
	pd := gui.Contexts().PickDialog
	if pd.Query == "" {
		return
	}

	tags, date, filter := ParsePickQuery(pd.Query)
	opts := self.scopedPickOpts(date, filter, false)

	res, err := self.c.RuinCmd().Pick.Pick(tags, opts)
	if err == nil {
		pd.Results = res
	} else {
		pd.Results = nil
	}

	// Clamp selection
	if pd.SelectedCardIdx >= len(pd.Results) {
		pd.SelectedCardIdx = max(len(pd.Results)-1, 0)
	}

	gui.RenderPickDialog()
}

// ClosePickDialog closes the pick dialog overlay.
func (self *PickHelper) ClosePickDialog() {
	gui := self.c.GuiCommon()
	gui.DeleteView(PickDialogView)
	gui.PopContext()
}

// PickDialogView is the view name constant for the pick dialog overlay.
const PickDialogView = "pickDialog"
