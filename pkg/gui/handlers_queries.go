package gui

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

func (gui *Gui) queriesDown(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.parentsDown(g, v)
	}
	if listMove(&gui.state.Queries.SelectedIndex, len(gui.state.Queries.Items), 1) {
		gui.renderQueries()
		gui.updatePreviewForQueries()
	}
	return nil
}

func (gui *Gui) queriesUp(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.parentsUp(g, v)
	}
	if listMove(&gui.state.Queries.SelectedIndex, len(gui.state.Queries.Items), -1) {
		gui.renderQueries()
		gui.updatePreviewForQueries()
	}
	return nil
}

func (gui *Gui) queriesClick(g *gocui.Gui, v *gocui.View) error {
	if gui.state.Queries.CurrentTab == QueriesTabParents {
		return gui.parentsClick(g, v)
	}
	_, cy := v.Cursor()
	_, oy := v.Origin()
	idx := (cy + oy) / 2 // 2 lines per query
	if idx >= 0 && idx < len(gui.state.Queries.Items) {
		gui.state.Queries.SelectedIndex = idx
	}
	gui.setContext(QueriesContext)
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
	notes, err := gui.ruinCmd.Queries.Run(query.Name)
	if err != nil {
		return nil
	}

	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Preview: " + query.Name + " "
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
	gui.updatePreviewCardList(" Preview: "+query.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Queries.Run(query.Name)
	})
}
