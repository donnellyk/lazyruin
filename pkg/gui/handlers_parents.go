package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// cycleQueriesTab cycles through Queries -> Parents tabs
func (gui *Gui) cycleQueriesTab() {
	idx := (gui.queriesTabIndex() + 1) % len(queriesTabs)
	gui.state.Queries.CurrentTab = queriesTabs[idx]
	gui.loadDataForQueriesTab()
}

// switchQueriesTabByIndex switches to a specific tab by index (for tab click)
func (gui *Gui) switchQueriesTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(queriesTabs) {
		return nil
	}
	gui.state.Queries.CurrentTab = queriesTabs[tabIndex]
	gui.loadDataForQueriesTab()
	gui.setContext(QueriesContext)
	return nil
}

// loadDataForQueriesTab refreshes data for the active queries tab
func (gui *Gui) loadDataForQueriesTab() {
	gui.updateQueriesTab()
	switch gui.state.Queries.CurrentTab {
	case QueriesTabParents:
		gui.refreshParents(false)
		gui.updatePreviewForParents()
	default:
		gui.refreshQueries(false)
		gui.updatePreviewForQueries()
	}
}

func (gui *Gui) refreshParents(preserve bool) {
	idx := gui.state.Parents.SelectedIndex
	parents, err := gui.ruinCmd.Parent.List()
	if err != nil {
		return
	}
	gui.state.Parents.Items = parents
	if preserve && idx < len(parents) {
		gui.state.Parents.SelectedIndex = idx
	} else {
		gui.state.Parents.SelectedIndex = 0
	}
	gui.renderQueries()
}

func (gui *Gui) parentsDown(g *gocui.Gui, v *gocui.View) error {
	if listMove(&gui.state.Parents.SelectedIndex, len(gui.state.Parents.Items), 1) {
		gui.renderQueries()
		gui.updatePreviewForParents()
	}
	return nil
}

func (gui *Gui) parentsUp(g *gocui.Gui, v *gocui.View) error {
	if listMove(&gui.state.Parents.SelectedIndex, len(gui.state.Parents.Items), -1) {
		gui.renderQueries()
		gui.updatePreviewForParents()
	}
	return nil
}

func (gui *Gui) parentsClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	idx := (cy + oy) / 2 // 2 lines per parent
	if idx >= 0 && idx < len(gui.state.Parents.Items) {
		gui.state.Parents.SelectedIndex = idx
	}
	gui.setContext(QueriesContext)
	return nil
}

func (gui *Gui) viewParent(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Parents.Items) == 0 {
		return nil
	}

	parent := gui.state.Parents.Items[gui.state.Parents.SelectedIndex]
	composed, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		return nil
	}

	gui.showComposedNote(composed, parent.Name)
	gui.setContext(PreviewContext)

	return nil
}

func (gui *Gui) deleteParent(g *gocui.Gui, v *gocui.View) error {
	if len(gui.state.Parents.Items) == 0 {
		return nil
	}

	parent := gui.state.Parents.Items[gui.state.Parents.SelectedIndex]

	gui.showConfirm("Delete Parent", "Delete parent bookmark \""+parent.Name+"\"?", func() error {
		err := gui.ruinCmd.Parent.Delete(parent.Name)
		if err != nil {
			return nil
		}
		gui.refreshParents(false)
		return nil
	})
	return nil
}

func (gui *Gui) updatePreviewForParents() {
	if len(gui.state.Parents.Items) == 0 {
		return
	}

	parent := gui.state.Parents.Items[gui.state.Parents.SelectedIndex]
	composed, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		return
	}

	gui.showComposedNote(composed, parent.Name)
}

// showComposedNote puts a single composed note into the preview as a one-card card list.
func (gui *Gui) showComposedNote(note models.Note, label string) {
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = []models.Note{note}
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.ScrollOffset = 0
	gui.state.Preview.EditMode = false
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Preview: " + label + " "
		gui.renderPreview()
	}
}
