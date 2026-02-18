package context

import (
	"slices"

	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// TagsTab represents the sub-tabs within the Tags panel.
type TagsTab string

const (
	TagsTabAll    TagsTab = "all"
	TagsTabGlobal TagsTab = "global"
	TagsTabInline TagsTab = "inline"
)

// TagsTabs maps tab indices to TagsTab values.
var TagsTabs = []TagsTab{TagsTabAll, TagsTabGlobal, TagsTabInline}

// TagsContext owns all Tags panel state: items, cursor, and tab.
type TagsContext struct {
	BaseContext
	*ListContextTrait

	Items      []models.Tag
	CurrentTab TagsTab
	list       *tagsList
}

// tagsList adapts TagsContext for the IList and ICursorList interfaces.
type tagsList struct {
	ctx *TagsContext
}

func (l *tagsList) Len() int {
	return len(l.ctx.FilteredItems())
}

func (l *tagsList) GetSelectedItemId() string {
	item := l.ctx.Selected()
	if item == nil {
		return ""
	}
	return item.Name
}

func (l *tagsList) FindIndexById(id string) int {
	for i, tag := range l.ctx.FilteredItems() {
		if tag.Name == id {
			return i
		}
	}
	return -1
}

// NewTagsContext creates a TagsContext.
// renderFn and previewFn are called after selection changes.
func NewTagsContext(renderFn func(), previewFn func()) *TagsContext {
	ctx := &TagsContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.SIDE_CONTEXT,
			Key:       "tags",
			ViewName:  "tags",
			Focusable: true,
			Title:     "Tags",
		}),
		CurrentTab: TagsTabAll,
	}
	ctx.list = &tagsList{ctx: ctx}
	cursor := NewListCursor(ctx.list)
	ctx.ListContextTrait = NewListContextTrait(cursor, renderFn, previewFn)
	return ctx
}

// FilteredItems returns tags visible under the current tab.
func (self *TagsContext) FilteredItems() []models.Tag {
	switch self.CurrentTab {
	case TagsTabGlobal:
		return filterTagsByScope(self.Items, "global")
	case TagsTabInline:
		return filterTagsByScope(self.Items, "inline")
	default:
		return self.Items
	}
}

// Selected returns the currently selected tag from the filtered list, or nil.
func (self *TagsContext) Selected() *models.Tag {
	items := self.FilteredItems()
	if len(items) == 0 {
		return nil
	}
	idx := self.GetSelectedLineIdx()
	if idx >= len(items) {
		idx = 0
	}
	return &items[idx]
}

// TabIndex returns the current tab index.
func (self *TagsContext) TabIndex() int {
	switch self.CurrentTab {
	case TagsTabGlobal:
		return 1
	case TagsTabInline:
		return 2
	default:
		return 0
	}
}

// GetList returns the IList adapter for this context.
func (self *TagsContext) GetList() types.IList {
	return &tagsListAdapter{ctx: self}
}

// GetSelectedItemId returns the stable ID of the selected item.
func (self *TagsContext) GetSelectedItemId() string {
	return self.list.GetSelectedItemId()
}

// tagsListAdapter wraps TagsContext to implement types.IList.
type tagsListAdapter struct {
	ctx *TagsContext
}

func (a *tagsListAdapter) Len() int                     { return a.ctx.list.Len() }
func (a *tagsListAdapter) GetSelectedItemId() string     { return a.ctx.list.GetSelectedItemId() }
func (a *tagsListAdapter) FindIndexById(id string) int   { return a.ctx.list.FindIndexById(id) }
func (a *tagsListAdapter) GetSelectedLineIdx() int       { return a.ctx.GetSelectedLineIdx() }
func (a *tagsListAdapter) SetSelectedLineIdx(idx int)    { a.ctx.SetSelectedLineIdx(idx) }
func (a *tagsListAdapter) MoveSelectedLine(delta int)    { a.ctx.MoveSelectedLine(delta) }
func (a *tagsListAdapter) ClampSelection()               { a.ctx.ClampSelection() }

func filterTagsByScope(tags []models.Tag, scope string) []models.Tag {
	var out []models.Tag
	for _, t := range tags {
		if slices.Contains(t.Scope, scope) {
			out = append(out, t)
		}
	}
	return out
}

// Verify interface compliance at compile time.
var _ types.IListContext = &TagsContext{}
