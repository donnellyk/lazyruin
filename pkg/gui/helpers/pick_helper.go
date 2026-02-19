package helpers

import (
	"strings"

	"kvnd/lazyruin/pkg/gui/types"
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
	gui.PushContextByKey("pick")
	return nil
}

// ExecutePick parses the raw input, runs the pick command, and shows results.
func (self *PickHelper) ExecutePick(raw string) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Pick

	if raw == "" {
		return self.CancelPick()
	}

	// Parse tags and @date filters from input
	var tags []string
	var filters []string
	for _, token := range strings.Fields(raw) {
		if strings.HasPrefix(token, "@") {
			filters = append(filters, token)
		} else {
			if !strings.HasPrefix(token, "#") {
				token = "#" + token
			}
			tags = append(tags, token)
		}
	}

	results, err := self.c.RuinCmd().Pick.Pick(tags, ctx.AnyMode, strings.Join(filters, " "))

	// Always close the pick dialog
	ctx.Query = raw
	ctx.Completion = types.NewCompletionState()
	gui.SetCursorEnabled(false)

	if err != nil {
		results = nil
	}

	self.c.Helpers().Preview().ShowPickResults(" Pick: "+raw+" ", results)
	gui.ReplaceContextByKey("preview")
	return nil
}

// CancelPick closes the pick popup without executing.
func (self *PickHelper) CancelPick() error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Pick
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
