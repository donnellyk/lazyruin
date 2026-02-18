package gui

import (
	guictx "kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// cycleQueriesTab cycles through Queries -> Parents tabs
func (gui *Gui) cycleQueriesTab() {
	queriesCtx := gui.contexts.Queries
	idx := (queriesCtx.TabIndex() + 1) % len(guictx.QueriesTabs)
	queriesCtx.CurrentTab = guictx.QueriesTabs[idx]
	queriesCtx.SetSelectedLineIdx(0)
	gui.loadDataForQueriesTab()
}

// switchQueriesTabByIndex switches to a specific tab by index (for tab click)
func (gui *Gui) switchQueriesTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(guictx.QueriesTabs) {
		return nil
	}
	queriesCtx := gui.contexts.Queries
	queriesCtx.CurrentTab = guictx.QueriesTabs[tabIndex]
	queriesCtx.SetSelectedLineIdx(0)
	gui.loadDataForQueriesTab()
	gui.setContext(QueriesContext)
	return nil
}

// loadDataForQueriesTab refreshes data for the active queries tab
func (gui *Gui) loadDataForQueriesTab() {
	gui.updateQueriesTab()
	switch gui.contexts.Queries.CurrentTab {
	case guictx.QueriesTabParents:
		gui.refreshParents(false)
		gui.updatePreviewForParents()
	default:
		gui.refreshQueries(false)
		gui.updatePreviewForQueries()
	}
}

func (gui *Gui) runQuery(g *gocui.Gui, v *gocui.View) error {
	queriesCtx := gui.contexts.Queries
	if queriesCtx.CurrentTab == guictx.QueriesTabParents {
		return gui.viewParent(g, v)
	}
	query := queriesCtx.SelectedQuery()
	if query == nil {
		return nil
	}

	notes, err := gui.ruinCmd.Queries.Run(query.Name, gui.buildSearchOptions())
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.preview.pushNavHistory()
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = notes
	gui.state.Preview.SelectedCardIndex = 0
	gui.views.Preview.Title = " Query: " + query.Name + " "
	gui.renderPreview()
	gui.setContext(PreviewContext)

	return nil
}

func (gui *Gui) deleteQuery(g *gocui.Gui, v *gocui.View) error {
	queriesCtx := gui.contexts.Queries
	if queriesCtx.CurrentTab == guictx.QueriesTabParents {
		return gui.deleteParent(g, v)
	}
	query := queriesCtx.SelectedQuery()
	if query == nil {
		return nil
	}

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
	queriesCtx := gui.contexts.Queries
	if queriesCtx.CurrentTab == guictx.QueriesTabParents {
		gui.updatePreviewForParents()
		return
	}
	query := queriesCtx.SelectedQuery()
	if query == nil {
		return
	}

	gui.preview.updatePreviewCardList(" Query: "+query.Name+" ", func() ([]models.Note, error) {
		return gui.ruinCmd.Queries.Run(query.Name, gui.buildSearchOptions())
	})
}

func (gui *Gui) viewParent(g *gocui.Gui, v *gocui.View) error {
	queriesCtx := gui.contexts.Queries
	parent := queriesCtx.SelectedParent()
	if parent == nil {
		return nil
	}

	composed, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		gui.showError(err)
		return nil
	}

	gui.showComposedNote(composed, parent.Name)
	gui.setContext(PreviewContext)

	return nil
}

func (gui *Gui) deleteParent(g *gocui.Gui, v *gocui.View) error {
	queriesCtx := gui.contexts.Queries
	parent := queriesCtx.SelectedParent()
	if parent == nil {
		return nil
	}

	gui.showConfirm("Delete Parent", "Delete parent bookmark \""+parent.Name+"\"?", func() error {
		err := gui.ruinCmd.Parent.Delete(parent.Name)
		if err != nil {
			gui.showError(err)
			return nil
		}
		gui.refreshParents(false)
		return nil
	})
	return nil
}

func (gui *Gui) updatePreviewForParents() {
	queriesCtx := gui.contexts.Queries
	parent := queriesCtx.SelectedParent()
	if parent == nil {
		return
	}

	composed, err := gui.ruinCmd.Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		return
	}

	gui.showComposedNote(composed, parent.Name)
}

// showComposedNote puts a single composed note into the preview as a one-card card list.
func (gui *Gui) showComposedNote(note models.Note, label string) {
	gui.preview.pushNavHistory()
	gui.state.Preview.Mode = PreviewModeCardList
	gui.state.Preview.Cards = []models.Note{note}
	gui.state.Preview.SelectedCardIndex = 0
	gui.state.Preview.ScrollOffset = 0
	if gui.views.Preview != nil {
		gui.views.Preview.Title = " Parent: " + label + " "
		gui.renderPreview()
	}
}
