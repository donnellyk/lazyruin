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
	ctx.TodoMode = false
	ctx.AllTagsMode = false
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
	ctx.TodoMode = false
	ctx.AllTagsMode = false
	ctx.SeedHash = true
	ctx.DialogMode = true
	ctx.ScopeTitle = scopeTitle
	gui.PushContextByKey("pick")
	return nil
}

// PickFlags holds boolean flags parsed from the pick input text.
type PickFlags struct {
	Any     bool
	Todo    bool
	AllTags bool
}

// ParsePickQuery splits raw pick input into tags, an optional @date
// (line-level date filter), remaining filter text, and any --flags.
func ParsePickQuery(raw string) (tags []string, date string, filter string, flags PickFlags) {
	for _, token := range strings.Fields(raw) {
		switch token {
		case "--any":
			flags.Any = true
		case "--todo":
			flags.Todo = true
		case "--all-tags":
			flags.AllTags = true
		default:
			if strings.HasPrefix(token, "@") {
				date = token
			} else {
				if !strings.HasPrefix(token, "#") {
					token = "#" + token
				}
				tags = append(tags, token)
			}
		}
	}
	return tags, date, "", flags
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

	tags, date, filter, flags := ParsePickQuery(raw)
	results, err := self.c.RuinCmd().Pick.Pick(tags, commands.PickOpts{
		Any:  ctx.AnyMode || flags.Any,
		Todo: ctx.TodoMode || flags.Todo,
		Date: date, Filter: filter,
	})

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
func (self *PickHelper) scopedPickOpts(date, filter string, anyMode, todoMode bool) commands.PickOpts {
	opts := commands.PickOpts{Any: anyMode, Todo: todoMode, Date: date, Filter: filter}
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

// mergeTagsDedup merges typed and scoped tags, deduplicating case-insensitively.
// Typed tags take precedence (their casing is preserved).
func mergeTagsDedup(typed, scoped []string) []string {
	seen := make(map[string]bool, len(typed))
	for _, t := range typed {
		key := strings.ToLower(strings.TrimPrefix(t, "#"))
		seen[key] = true
	}
	merged := make([]string, len(typed))
	copy(merged, typed)
	for _, s := range scoped {
		key := strings.ToLower(strings.TrimPrefix(s, "#"))
		if !seen[key] {
			seen[key] = true
			if !strings.HasPrefix(s, "#") {
				s = "#" + s
			}
			merged = append(merged, s)
		}
	}
	return merged
}

// executePickDialog runs a pick and shows results in the dialog overlay.
// In compose mode, results are scoped to the parent's children via --parent.
// In cardList mode, results are scoped to the selected note via --notes.
// Otherwise, all results are shown.
func (self *PickHelper) executePickDialog(raw string, ctx *context.PickContext) error {
	gui := self.c.GuiCommon()

	tags, date, filter, flags := ParsePickQuery(raw)

	allTagsActive := ctx.AllTagsMode || flags.AllTags
	anyMode := ctx.AnyMode || flags.Any
	todoMode := ctx.TodoMode || flags.Todo
	if allTagsActive {
		scopedTags := self.c.Helpers().Completion().ScopedInlineTags()
		// Exclude #done — it's a status marker, not a meaningful filter tag.
		filtered := scopedTags[:0:0]
		for _, t := range scopedTags {
			if !strings.EqualFold(strings.TrimPrefix(t, "#"), "done") {
				filtered = append(filtered, t)
			}
		}
		tags = mergeTagsDedup(tags, filtered)
		anyMode = true
	}

	// Build a resolved query that bakes in expanded tags and toggle flags,
	// so ReloadPickDialog can re-execute without needing PickContext state.
	resolvedQuery := buildResolvedQuery(tags, date, anyMode, todoMode)

	ctx.Query = raw
	ctx.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)

	opts := self.scopedPickOpts(date, filter, anyMode, todoMode)

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
	pd.Query = resolvedQuery
	pd.ScopeTitle = ctx.ScopeTitle
	gui.ReplaceContextByKey("pickDialog")
	return nil
}

// buildResolvedQuery assembles a canonical query string from already-resolved
// tags and flags. This is stored in pd.Query so that ReloadPickDialog can
// re-execute without needing the original PickContext toggle state.
func buildResolvedQuery(tags []string, date string, anyMode, todoMode bool) string {
	parts := make([]string, 0, len(tags)+3)
	parts = append(parts, tags...)
	if date != "" {
		parts = append(parts, date)
	}
	if anyMode {
		parts = append(parts, "--any")
	}
	if todoMode {
		parts = append(parts, "--todo")
	}
	return strings.Join(parts, " ")
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

// TogglePickTodo toggles the todo-mode flag for pick.
func (self *PickHelper) TogglePickTodo() {
	ctx := self.c.GuiCommon().Contexts().Pick
	ctx.TodoMode = !ctx.TodoMode
}

// TogglePickAllTags toggles the all-tags mode flag for pick.
func (self *PickHelper) TogglePickAllTags() {
	ctx := self.c.GuiCommon().Contexts().Pick
	ctx.AllTagsMode = !ctx.AllTagsMode
}

// ReloadPickDialog re-runs the current pick dialog query and updates results.
// pd.Query is already a resolved query (tags expanded, flags baked in), so
// this just parses and executes — no PickContext state needed.
func (self *PickHelper) ReloadPickDialog() {
	gui := self.c.GuiCommon()
	pd := gui.Contexts().PickDialog
	if pd.Query == "" {
		return
	}

	tags, date, filter, flags := ParsePickQuery(pd.Query)
	opts := self.scopedPickOpts(date, filter, flags.Any, flags.Todo)

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
