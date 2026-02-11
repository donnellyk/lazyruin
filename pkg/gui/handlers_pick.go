package gui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) openPick(g *gocui.Gui, v *gocui.View) error {
	gui.state.PickMode = true
	gui.state.PickCompletion = NewCompletionState()
	gui.state.PickAnyMode = false
	gui.state.PickSeedHash = true
	gui.setContext(PickContext)
	return nil
}

func (gui *Gui) pickEnter(g *gocui.Gui, v *gocui.View) error {
	if gui.state.PickCompletion.Active {
		gui.acceptCompletion(v, gui.state.PickCompletion, gui.pickTriggers())
		return nil
	}
	return gui.executePick(g, v)
}

func (gui *Gui) pickEsc(g *gocui.Gui, v *gocui.View) error {
	if gui.state.PickCompletion.Active {
		gui.state.PickCompletion.Active = false
		gui.state.PickCompletion.Items = nil
		gui.state.PickCompletion.SelectedIndex = 0
		return nil
	}
	return gui.cancelPick(g, v)
}

func (gui *Gui) pickTab(g *gocui.Gui, v *gocui.View) error {
	if gui.state.PickCompletion.Active {
		gui.acceptCompletion(v, gui.state.PickCompletion, gui.pickTriggers())
	}
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

	// Parse tags from input
	var tags []string
	for _, token := range strings.Fields(raw) {
		if !strings.HasPrefix(token, "#") {
			token = "#" + token
		}
		tags = append(tags, token)
	}

	results, err := gui.ruinCmd.Pick.Pick(tags, gui.state.PickAnyMode)
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.state.PickQuery = raw
	gui.state.PickMode = false
	gui.state.PickCompletion = NewCompletionState()
	g.Cursor = false

	gui.state.Preview.Mode = PreviewModePickResults
	gui.state.Preview.PickResults = results
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Pick: " + raw + " "
	}
	gui.renderPreview()

	gui.setContext(PreviewContext)
	return nil
}

func (gui *Gui) cancelPick(g *gocui.Gui, v *gocui.View) error {
	gui.state.PickMode = false
	gui.state.PickCompletion = NewCompletionState()
	g.Cursor = false
	gui.setContext(gui.state.PreviousContext)
	return nil
}
