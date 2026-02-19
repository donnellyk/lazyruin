package helpers

import (
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/models"
)

// QueriesHelper handles query and parent bookmark domain operations.
type QueriesHelper struct {
	c *HelperCommon
}

// NewQueriesHelper creates a new QueriesHelper.
func NewQueriesHelper(c *HelperCommon) *QueriesHelper {
	return &QueriesHelper{c: c}
}

// RefreshQueries fetches all queries and re-renders the list.
// If preserve is true, the current selection is preserved by ID.
func (self *QueriesHelper) RefreshQueries(preserve bool) {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	prevID := queriesCtx.GetQueriesList().GetSelectedItemId()

	queries, err := self.c.RuinCmd().Queries.List()
	if err != nil {
		return
	}
	queriesCtx.Queries = queries

	if preserve && prevID != "" {
		if newIdx := queriesCtx.GetQueriesList().FindIndexById(prevID); newIdx >= 0 {
			queriesCtx.QueriesTrait().SetSelectedLineIdx(newIdx)
		}
	} else {
		queriesCtx.QueriesTrait().SetSelectedLineIdx(0)
	}
	queriesCtx.QueriesTrait().ClampSelection()

	gui.RenderQueries()
}

// RefreshParents fetches all parent bookmarks and re-renders the list.
// If preserve is true, the current selection is preserved by ID.
func (self *QueriesHelper) RefreshParents(preserve bool) {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	prevID := queriesCtx.GetParentsList().GetSelectedItemId()

	parents, err := self.c.RuinCmd().Parent.List()
	if err != nil {
		return
	}
	queriesCtx.Parents = parents

	if preserve && prevID != "" {
		if newIdx := queriesCtx.GetParentsList().FindIndexById(prevID); newIdx >= 0 {
			queriesCtx.ParentsTrait().SetSelectedLineIdx(newIdx)
		}
	} else {
		queriesCtx.ParentsTrait().SetSelectedLineIdx(0)
	}
	queriesCtx.ParentsTrait().ClampSelection()

	gui.RenderQueries()
}

// CycleQueriesTab cycles through Queries -> Parents tabs.
func (self *QueriesHelper) CycleQueriesTab() {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	idx := (queriesCtx.TabIndex() + 1) % len(context.QueriesTabs)
	queriesCtx.CurrentTab = context.QueriesTabs[idx]
	queriesCtx.SetSelectedLineIdx(0)
	self.LoadDataForQueriesTab()
}

// SwitchQueriesTabByIndex switches to a specific tab by index (for tab click).
func (self *QueriesHelper) SwitchQueriesTabByIndex(tabIndex int) error {
	if tabIndex < 0 || tabIndex >= len(context.QueriesTabs) {
		return nil
	}
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	queriesCtx.CurrentTab = context.QueriesTabs[tabIndex]
	queriesCtx.SetSelectedLineIdx(0)
	self.LoadDataForQueriesTab()
	gui.PushContextByKey("queries")
	return nil
}

// LoadDataForQueriesTab refreshes data for the active queries tab.
func (self *QueriesHelper) LoadDataForQueriesTab() {
	gui := self.c.GuiCommon()
	gui.UpdateQueriesTab()
	switch gui.Contexts().Queries.CurrentTab {
	case context.QueriesTabParents:
		self.RefreshParents(false)
		self.UpdatePreviewForParents()
	default:
		self.RefreshQueries(false)
		self.UpdatePreviewForQueries()
	}
}

// RunQuery runs the selected query (or views the selected parent) and shows results in preview.
func (self *QueriesHelper) RunQuery() error {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	if queriesCtx.CurrentTab == context.QueriesTabParents {
		return self.ViewParent()
	}
	query := queriesCtx.SelectedQuery()
	if query == nil {
		return nil
	}

	notes, err := self.c.RuinCmd().Queries.Run(query.Name, gui.BuildSearchOptions())
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(" Query: "+query.Name+" ", notes)
	gui.PushContextByKey("preview")
	return nil
}

// DeleteQuery shows confirmation and deletes the selected query (or parent).
func (self *QueriesHelper) DeleteQuery() error {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	if queriesCtx.CurrentTab == context.QueriesTabParents {
		return self.DeleteParent()
	}
	query := queriesCtx.SelectedQuery()
	if query == nil {
		return nil
	}

	self.c.Helpers().Confirmation().ConfirmDelete("Query", query.Name,
		func() error { return self.c.RuinCmd().Queries.Delete(query.Name) },
		func() { self.RefreshQueries(false) },
	)
	return nil
}

// ViewParent composes and shows the selected parent bookmark in preview.
func (self *QueriesHelper) ViewParent() error {
	gui := self.c.GuiCommon()
	parent := gui.Contexts().Queries.SelectedParent()
	if parent == nil {
		return nil
	}

	composed, err := self.c.RuinCmd().Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	self.ShowComposedNote(composed, parent.Name)
	gui.PushContextByKey("preview")
	return nil
}

// DeleteParent shows confirmation and deletes the selected parent bookmark.
func (self *QueriesHelper) DeleteParent() error {
	gui := self.c.GuiCommon()
	parent := gui.Contexts().Queries.SelectedParent()
	if parent == nil {
		return nil
	}

	self.c.Helpers().Confirmation().ConfirmDelete("Parent", parent.Name,
		func() error { return self.c.RuinCmd().Parent.Delete(parent.Name) },
		func() { self.RefreshParents(false) },
	)
	return nil
}

// UpdatePreviewForQueries updates the preview for the current queries tab.
func (self *QueriesHelper) UpdatePreviewForQueries() {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	if queriesCtx.CurrentTab == context.QueriesTabParents {
		self.UpdatePreviewForParents()
		return
	}
	query := queriesCtx.SelectedQuery()
	if query == nil {
		return
	}

	self.c.Helpers().Preview().UpdatePreviewCardList(" Query: "+query.Name+" ", func() ([]models.Note, error) {
		return self.c.RuinCmd().Queries.Run(query.Name, gui.BuildSearchOptions())
	})
}

// UpdatePreviewForParents updates the preview for the selected parent.
func (self *QueriesHelper) UpdatePreviewForParents() {
	gui := self.c.GuiCommon()
	parent := gui.Contexts().Queries.SelectedParent()
	if parent == nil {
		return
	}

	composed, err := self.c.RuinCmd().Parent.ComposeFlat(parent.UUID, parent.Title)
	if err != nil {
		return
	}

	self.ShowComposedNote(composed, parent.Name)
}

// ShowComposedNote puts a single composed note into the preview as a one-card card list.
func (self *QueriesHelper) ShowComposedNote(note models.Note, label string) {
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList(" Parent: "+label+" ", []models.Note{note})
}
