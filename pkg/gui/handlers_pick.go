package gui

import (
	"kvnd/lazyruin/pkg/gui/types"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) openPick(g *gocui.Gui, v *gocui.View) error {
	if gui.popupActive() {
		return nil
	}
	gui.state.PickCompletion = types.NewCompletionState()
	gui.state.PickAnyMode = false
	gui.state.PickSeedHash = true
	gui.pushContextByKey("pick")
	return nil
}

func (gui *Gui) togglePickAny(g *gocui.Gui, v *gocui.View) error {
	gui.state.PickAnyMode = !gui.state.PickAnyMode
	if gui.views.Pick != nil {
		gui.views.Pick.Footer = gui.pickFooter()
	}
	return nil
}

func (gui *Gui) executePick(g *gocui.Gui, v *gocui.View) error {
	raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
	if raw == "" {
		return gui.cancelPick(g, v)
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

	results, err := gui.ruinCmd.Pick.Pick(tags, gui.state.PickAnyMode, strings.Join(filters, " "))

	// Always close the pick dialog
	gui.state.PickQuery = raw
	gui.state.PickCompletion = types.NewCompletionState()
	g.Cursor = false

	if err != nil {
		results = nil
	}

	gui.helpers.Preview().ShowPickResults(" Pick: "+raw+" ", results)

	gui.replaceContextByKey("preview")
	return nil
}

func (gui *Gui) cancelPick(g *gocui.Gui, v *gocui.View) error {
	gui.state.PickCompletion = types.NewCompletionState()
	g.Cursor = false
	gui.popContext()
	return nil
}
