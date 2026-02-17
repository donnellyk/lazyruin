package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) queriesPanel() *listPanel {
	return &listPanel{
		selectedIndex: &gui.state.Queries.SelectedIndex,
		itemCount:     func() int { return len(gui.state.Queries.Items) },
		render:        gui.renderQueries,
		updatePreview: gui.updatePreviewForQueries,
		context:       QueriesContext,
	}
}

func (gui *Gui) queriesDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.parentsDown(g, v)
	}
	return gui.queriesPanel().listDown(g, v)
}

func (gui *Gui) queriesUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.parentsUp(g, v)
	}
	return gui.queriesPanel().listUp(g, v)
}

func (gui *Gui) queriesClick(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.parentsClick(g, v)
	}
	idx := listClickIndex(v, 2)
	if idx >= 0 && idx < len(gui.state.Queries.Items) {
		gui.state.Queries.SelectedIndex = idx
	}
	gui.setContext(QueriesContext)
	return nil
}

func (gui *Gui) queriesWheelDown(g *gocui.Gui, v *gocui.View) error {
	scrollViewport(v, 3)
	return nil
}

func (gui *Gui) queriesWheelUp(g *gocui.Gui, v *gocui.View) error {
	scrollViewport(v, -3)
	return nil
}

func (gui *Gui) runQuery(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.viewParent(g, v)
	}
	if len(gui.state.Queries.Items) == 0 {
		return nil
	}

	query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
	notes, err := gui.ruinCmd.Queries.Run(query.Name, gui.buildSearchOptions())
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.pushNavHistory()
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Query: " + query.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)

	return nil
}

func (gui *Gui) deleteQuery(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.deleteParent(g, v)
	}
	if len(gui.state.Queries.Items) == 0 {
		return nil
	}

	query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]

	gui.showConfirm("Delete Query", "Delete query \""+query.Name+"\"?", func() error {
		err := gui.ruinCmd.Queries.Delete(query.Name)
		if err != nil {
			gui.showError(err)
			return nil
		}
		gui.refreshQueries(false)
		return nil
	})
	return nil
}

func (gui *Gui) updatePreviewForQueries() {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		gui.updatePreviewForParents()
		return
	}
	if len(gui.state.Queries.Items) == 0 {
		return
	}

	query := gui.state.Queries.Items[gui.state.Queries.SelectedIndex]
	gui.updatePreviewCardList(" Query: "+query.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Queries.Run(query.Name, gui.buildSearchOptions())
	})
}
