package context

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// QueriesTab represents the sub-tabs within the Queries panel.
type QueriesTab string

const (
	QueriesTabQueries QueriesTab = "queries"
	QueriesTabParents QueriesTab = "parents"
)

// QueriesTabs maps tab indices to QueriesTab values.
var QueriesTabs = []QueriesTab{QueriesTabQueries, QueriesTabParents}

// QueriesContext owns all Queries panel state: queries, parents, cursors, and tab.
// The panel shows either the Queries list or the Parents list depending on CurrentTab.
type QueriesContext struct {
	BaseContext

	Queries    []models.Query
	Parents    []models.ParentBookmark
	CurrentTab QueriesTab

	queriesTrait *ListContextTrait
	parentsTrait *ListContextTrait
	queriesList  *queriesList
	parentsList  *parentsList
}

// queriesList adapts QueriesContext to IList for the queries tab.
type queriesList struct {
	ctx *QueriesContext
}

func (l *queriesList) Len() int { return len(l.ctx.Queries) }

func (l *queriesList) GetSelectedItemId() string {
	if len(l.ctx.Queries) == 0 {
		return ""
	}
	idx := l.ctx.queriesTrait.GetSelectedLineIdx()
	if idx >= len(l.ctx.Queries) {
		return ""
	}
	return l.ctx.Queries[idx].Name
}

func (l *queriesList) FindIndexById(id string) int {
	for i, q := range l.ctx.Queries {
		if q.Name == id {
			return i
		}
	}
	return -1
}

// parentsList adapts QueriesContext to IList for the parents tab.
type parentsList struct {
	ctx *QueriesContext
}

func (l *parentsList) Len() int { return len(l.ctx.Parents) }

func (l *parentsList) GetSelectedItemId() string {
	if len(l.ctx.Parents) == 0 {
		return ""
	}
	idx := l.ctx.parentsTrait.GetSelectedLineIdx()
	if idx >= len(l.ctx.Parents) {
		return ""
	}
	return l.ctx.Parents[idx].UUID
}

func (l *parentsList) FindIndexById(id string) int {
	for i, p := range l.ctx.Parents {
		if p.UUID == id {
			return i
		}
	}
	return -1
}

// NewQueriesContext creates a QueriesContext.
// queriesRenderFn/queriesPreviewFn and parentsRenderFn/parentsPreviewFn
// are called when selection changes in each respective tab.
func NewQueriesContext(
	queriesRenderFn func(), queriesPreviewFn func(),
	parentsRenderFn func(), parentsPreviewFn func(),
) *QueriesContext {
	ctx := &QueriesContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.SIDE_CONTEXT,
			Key:       "queries",
			ViewName:  "queries",
			Focusable: true,
			Title:     "Queries",
		}),
		CurrentTab: QueriesTabQueries,
	}

	ctx.queriesList = &queriesList{ctx: ctx}
	ctx.parentsList = &parentsList{ctx: ctx}

	queriesCursor := NewListCursor(ctx.queriesList)
	parentsCursor := NewListCursor(ctx.parentsList)

	ctx.queriesTrait = NewListContextTrait(queriesCursor, queriesRenderFn, queriesPreviewFn)
	ctx.parentsTrait = NewListContextTrait(parentsCursor, parentsRenderFn, parentsPreviewFn)

	return ctx
}

// ActiveTrait returns the ListContextTrait for the currently active tab.
func (self *QueriesContext) ActiveTrait() *ListContextTrait {
	if self.CurrentTab == QueriesTabParents {
		return self.parentsTrait
	}
	return self.queriesTrait
}

// QueriesTrait returns the queries tab trait (for direct access).
func (self *QueriesContext) QueriesTrait() *ListContextTrait {
	return self.queriesTrait
}

// ParentsTrait returns the parents tab trait (for direct access).
func (self *QueriesContext) ParentsTrait() *ListContextTrait {
	return self.parentsTrait
}

// SelectedQuery returns the selected query or nil.
func (self *QueriesContext) SelectedQuery() *models.Query {
	if len(self.Queries) == 0 {
		return nil
	}
	idx := self.queriesTrait.GetSelectedLineIdx()
	if idx >= len(self.Queries) {
		return nil
	}
	return &self.Queries[idx]
}

// SelectedParent returns the selected parent or nil.
func (self *QueriesContext) SelectedParent() *models.ParentBookmark {
	if len(self.Parents) == 0 {
		return nil
	}
	idx := self.parentsTrait.GetSelectedLineIdx()
	if idx >= len(self.Parents) {
		return nil
	}
	return &self.Parents[idx]
}

// TabIndex returns the current tab index.
func (self *QueriesContext) TabIndex() int { return TabIndexOf(QueriesTabs, self.CurrentTab) }

// ActiveItemCount returns the number of items in the active tab.
func (self *QueriesContext) ActiveItemCount() int {
	if self.CurrentTab == QueriesTabParents {
		return len(self.Parents)
	}
	return len(self.Queries)
}

// GetList returns the IList adapter for the active tab.
func (self *QueriesContext) GetList() types.IList {
	if self.CurrentTab == QueriesTabParents {
		return self.GetParentsList()
	}
	return self.GetQueriesList()
}

// GetQueriesList returns the IList adapter for the queries tab (regardless of active tab).
func (self *QueriesContext) GetQueriesList() types.IList {
	return NewListAdapter(
		self.queriesList.Len,
		self.queriesList.GetSelectedItemId,
		self.queriesList.FindIndexById,
		func() *ListContextTrait { return self.queriesTrait },
	)
}

// GetParentsList returns the IList adapter for the parents tab (regardless of active tab).
func (self *QueriesContext) GetParentsList() types.IList {
	return NewListAdapter(
		self.parentsList.Len,
		self.parentsList.GetSelectedItemId,
		self.parentsList.FindIndexById,
		func() *ListContextTrait { return self.parentsTrait },
	)
}

// GetSelectedItemId returns the stable ID for the currently selected item.
func (self *QueriesContext) GetSelectedItemId() string {
	if self.CurrentTab == QueriesTabParents {
		return self.parentsList.GetSelectedItemId()
	}
	return self.queriesList.GetSelectedItemId()
}

// IListContext delegation — routes to the active tab's trait.

func (self *QueriesContext) GetSelectedLineIdx() int    { return self.ActiveTrait().GetSelectedLineIdx() }
func (self *QueriesContext) SetSelectedLineIdx(idx int) { self.ActiveTrait().SetSelectedLineIdx(idx) }
func (self *QueriesContext) MoveSelectedLine(delta int) { self.ActiveTrait().MoveSelectedLine(delta) }
func (self *QueriesContext) ClampSelection()            { self.ActiveTrait().ClampSelection() }

// Verify interface compliance at compile time.
var _ types.IListContext = &QueriesContext{}
