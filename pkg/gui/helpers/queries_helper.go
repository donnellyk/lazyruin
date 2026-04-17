package helpers

import (
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/models"
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

// RunQuery runs the selected query (or views the selected parent) and shows
// results in preview as a committed navigation.
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
	queryCopy := *query

	return self.c.Helpers().Navigator().NavigateTo("cardList", "Query: "+queryCopy.Name, func() error {
		opts := self.c.Helpers().Preview().BuildSearchOptions()
		notes, err := self.c.RuinCmd().Queries.Run(queryCopy.Name, opts)
		if err != nil {
			gui.ShowError(err)
			return err
		}
		source := self.c.Helpers().Preview().NewSearchSource(queryCopy.Query, "")
		self.c.Helpers().Preview().ShowCardList("Query: "+queryCopy.Name, notes, source)
		return nil
	})
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
	parentCopy := *parent

	return self.c.Helpers().Navigator().NavigateTo("compose", "Parent: "+parentCopy.Name, func() error {
		composed, sourceMap, err := self.composeParent(&parentCopy)
		if err != nil {
			gui.ShowError(err)
			return err
		}
		self.c.Helpers().Preview().ShowCompose("Parent: "+parentCopy.Name, composed, sourceMap, parentCopy)
		gui.Contexts().Compose.Requery = self.parentRequery(parentCopy)
		return nil
	})
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

// UpdatePreviewForParents updates the preview for the selected parent as a
// hover preview — no history entry recorded.
func (self *QueriesHelper) UpdatePreviewForParents() {
	gui := self.c.GuiCommon()
	parent := gui.Contexts().Queries.SelectedParent()
	if parent == nil {
		return
	}
	parentCopy := *parent

	_ = self.c.Helpers().Navigator().ShowHover("compose", "Parent: "+parentCopy.Name, func() error {
		composed, sourceMap, err := self.composeParent(&parentCopy)
		if err != nil {
			return err
		}
		self.c.Helpers().Preview().ShowCompose("Parent: "+parentCopy.Name, composed, sourceMap, parentCopy)
		gui.Contexts().Compose.Requery = self.parentRequery(parentCopy)
		return nil
	})
}

// composeParent runs compose for a parent bookmark.
func (self *QueriesHelper) composeParent(parent *models.ParentBookmark) (models.Note, []models.SourceMapEntry, error) {
	return self.c.RuinCmd().Parent.Compose(*parent)
}

// parentRequery returns a closure that re-composes the given parent. Used as
// the ComposeContext.Requery to re-run on history restore.
func (self *QueriesHelper) parentRequery(parent models.ParentBookmark) context.ComposeRequery {
	return func() (models.Note, []models.SourceMapEntry, error) {
		return self.c.RuinCmd().Parent.Compose(parent)
	}
}
