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

	err := RefreshList(
		func() ([]models.Query, error) { return self.c.RuinCmd().Queries.List() },
		func(queries []models.Query) { queriesCtx.Queries = queries },
		queriesCtx.GetQueriesList(),
		preserve,
	)
	if err != nil {
		gui.ShowError(err)
	}
	gui.RenderQueries()
}

// RefreshParents fetches all parent bookmarks and re-renders the list.
// If preserve is true, the current selection is preserved by ID.
func (self *QueriesHelper) RefreshParents(preserve bool) {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries

	err := RefreshList(
		func() ([]models.ParentBookmark, error) { return self.c.RuinCmd().Parent.List() },
		func(parents []models.ParentBookmark) { queriesCtx.Parents = parents },
		queriesCtx.GetParentsList(),
		preserve,
	)
	if err != nil {
		gui.ShowError(err)
	}
	gui.RenderQueries()
}

// CycleQueriesTab cycles through Queries -> Parents tabs.
func (self *QueriesHelper) CycleQueriesTab() {
	queriesCtx := self.c.GuiCommon().Contexts().Queries
	CycleTab(context.QueriesTabs, queriesCtx.TabIndex(), func(tab context.QueriesTab) {
		queriesCtx.CurrentTab = tab
		queriesCtx.SetSelectedLineIdx(0)
	}, self.LoadDataForQueriesTab)
}

// SwitchQueriesTabByIndex switches to a specific tab by index (for tab click).
func (self *QueriesHelper) SwitchQueriesTabByIndex(tabIndex int) error {
	gui := self.c.GuiCommon()
	queriesCtx := gui.Contexts().Queries
	SwitchTab(context.QueriesTabs, tabIndex, func(tab context.QueriesTab) {
		queriesCtx.CurrentTab = tab
		queriesCtx.SetSelectedLineIdx(0)
	}, func() {
		self.LoadDataForQueriesTab()
		gui.PushContextByKey("queries")
	})
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

	opts := self.c.Helpers().Preview().BuildSearchOptions()
	notes, err := self.c.RuinCmd().Queries.Run(query.Name, opts)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	source := self.c.Helpers().Preview().NewSearchSource(query.Query, "")

	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCardList("Query: "+query.Name, notes, source)
	gui.PushContextByKey("cardList")
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

	composed, sourceMap, err := self.composeParent(parent)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	self.ShowComposedNote(composed, sourceMap, parent)
	gui.PushContextByKey("compose")
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

	self.c.Helpers().Preview().UpdatePreviewCardList("Query: "+query.Name, func() ([]models.Note, error) {
		return self.c.RuinCmd().Queries.Run(query.Name, self.c.Helpers().Preview().BuildSearchOptions())
	})
}

// UpdatePreviewForParents updates the preview for the selected parent.
func (self *QueriesHelper) UpdatePreviewForParents() {
	gui := self.c.GuiCommon()
	parent := gui.Contexts().Queries.SelectedParent()
	if parent == nil {
		return
	}

	composed, sourceMap, err := self.composeParent(parent)
	if err != nil {
		return
	}

	self.ShowComposedNote(composed, sourceMap, parent)
}

// composeParent runs compose for a parent bookmark.
func (self *QueriesHelper) composeParent(parent *models.ParentBookmark) (models.Note, []models.SourceMapEntry, error) {
	return self.c.RuinCmd().Parent.Compose(*parent)
}

// ShowComposedNote puts a single composed note into the preview as a compose view.
func (self *QueriesHelper) ShowComposedNote(note models.Note, sourceMap []models.SourceMapEntry, parent *models.ParentBookmark) {
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.c.Helpers().Preview().ShowCompose("Parent: "+parent.Name, note, sourceMap, *parent)
}
