# DRY & Boilerplate Reduction Refactor Plan

This plan targets three axes: **boilerplate reduction**, **regression minimization**, and **DRY**. Each step is independently deployable and keeps `go build ./...` and `scripts/smoke-test.sh` green.

This plan is complementary to `REFACTOR_DETAILED_PLAN.md` (which focuses on moving handler logic into helpers/controllers). The steps here can be interleaved with that plan or executed independently.

---

## Overview

| # | Step | LOC Saved (est.) | Risk |
|---|------|-----------------|------|
| 1 | Generic `PopupController[C]` | ~150 | Low |
| 2 | `ListMouseTrait` for mouse bindings | ~80 | Low |
| 3 | Generic `ListAdapter[T]` for contexts | ~70 | Low |
| 4 | `tabIndex` utility + `TabContext` interface | ~30 | Low |
| 5 | Move `h()` to `ControllerCommon` | ~15 | Low |
| 6 | "Push results to Preview" helper methods | ~60 | Medium |
| 7 | "Confirm-then-delete" helper | ~40 | Low |
| 8 | Unify `IGuiCommon` into a single source of truth | ~0 (maintenance) | Medium |
| 9 | Derive `contextToView` from context metadata | ~25 | Low |
| 10 | Consolidate `gui_common.go` passthrough methods | ~30 | Low |

---

## Step 1: Generic `PopupController[C]`

### Problem

`SearchController`, `InputPopupController`, `CaptureController`, and `PickController` are near-identical. Each defines:
- A struct with `baseController` + `getContext` + `onEnter`/`onSubmit` + `onEsc` + `onTab` (+ optional extra)
- A matching `*Opts` struct
- A `New*Controller` constructor
- A `Context()` method
- A `GetKeybindingsFn()` that returns 3-4 bindings

The four files total ~200 LOC with almost no behavioral difference.

### Solution

Create `controllers/popup_controller.go`:

```go
// PopupController is a generic controller for popup contexts
// that need Enter/Esc/Tab plus optional extra bindings.
type PopupController[C types.Context] struct {
    baseController
    getContext func() C
    bindings   []*types.Binding
}

func NewPopupController[C types.Context](
    getContext func() C,
    bindings []*types.Binding,
) *PopupController[C] {
    return &PopupController[C]{
        getContext: getContext,
        bindings:   bindings,
    }
}

func (self *PopupController[C]) Context() types.Context {
    return self.getContext()
}

func (self *PopupController[C]) GetKeybindingsFn() types.KeybindingsFn {
    return func(opts types.KeybindingsOpts) []*types.Binding {
        return self.bindings
    }
}
```

### Migration

Replace each concrete controller with a `NewPopupController` call at the wiring site. For example, `setupSearchContext` becomes:

```go
func (gui *Gui) setupSearchContext() {
    gui.contexts.Search = context.NewSearchContext()
    searchState := func() *CompletionState { return gui.state.SearchCompletion }
    ctrl := controllers.NewPopupController(
        func() *context.SearchContext { return gui.contexts.Search },
        []*types.Binding{
            {Key: gocui.KeyEnter, Handler: func() error {
                return gui.completionEnter(searchState, gui.searchTriggers, gui.executeSearch)(gui.g, gui.views.Search)
            }},
            {Key: gocui.KeyEsc, Handler: func() error {
                return gui.completionEsc(searchState, gui.cancelSearch)(gui.g, gui.views.Search)
            }},
            {Key: gocui.KeyTab, Handler: func() error {
                return gui.completionTab(searchState, gui.searchTriggers)(gui.g, gui.views.Search)
            }},
        },
    )
    controllers.AttachController(ctrl)
}
```

### Files to delete after migration

- `controllers/search_controller.go`
- `controllers/input_popup_controller.go`
- `controllers/capture_controller.go`
- `controllers/pick_controller.go`

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (search, capture, pick, and input popup flows)
- `./lazyruin --debug-bindings` diff (binding IDs unchanged)

---

## Step 2: `ListMouseTrait` for Mouse Bindings

### Problem

`NotesController`, `TagsController`, and `QueriesController` each copy ~40 lines of `GetMouseKeybindingsFn` implementing click-to-select, wheel-up, and wheel-down. They differ only in:
- View name (`"notes"`, `"tags"`, `"queries"`)
- Click margin (3, 1, 2)
- Item count source (`ctx.Items`, `ctx.FilteredItems()`, or tab-dependent)
- Selection target (`ctx.SetSelectedLineIdx` or `ctx.QueriesTrait().SetSelectedLineIdx`)

### Solution

Create `controllers/list_mouse_trait.go`:

```go
// ListMouseOpts configures mouse behavior for a list panel.
type ListMouseOpts struct {
    ViewName     string
    ClickMargin  int
    ItemCount    func() int
    SetSelection func(idx int)
    GetContext   func() types.Context
    GuiCommon    func() IGuiCommon
}

// ListMouseBindings returns the standard mouse bindings for a list panel.
func ListMouseBindings(opts ListMouseOpts) types.MouseKeybindingsFn {
    return func(_ types.KeybindingsOpts) []*gocui.ViewMouseBinding {
        return []*gocui.ViewMouseBinding{
            {
                ViewName: opts.ViewName,
                Key:      gocui.MouseLeft,
                Handler: func(_ gocui.ViewMouseBindingOpts) error {
                    v := opts.GuiCommon().GetView(opts.ViewName)
                    if v == nil {
                        return nil
                    }
                    idx := helpers.ListClickIndex(v, opts.ClickMargin)
                    if idx >= 0 && idx < opts.ItemCount() {
                        opts.SetSelection(idx)
                    }
                    opts.GuiCommon().PushContext(opts.GetContext(), types.OnFocusOpts{})
                    return nil
                },
            },
            {
                ViewName: opts.ViewName,
                Key:      gocui.MouseWheelDown,
                Handler: func(_ gocui.ViewMouseBindingOpts) error {
                    if v := opts.GuiCommon().GetView(opts.ViewName); v != nil {
                        helpers.ScrollViewport(v, 3)
                    }
                    return nil
                },
            },
            {
                ViewName: opts.ViewName,
                Key:      gocui.MouseWheelUp,
                Handler: func(_ gocui.ViewMouseBindingOpts) error {
                    if v := opts.GuiCommon().GetView(opts.ViewName); v != nil {
                        helpers.ScrollViewport(v, -3)
                    }
                    return nil
                },
            },
        }
    }
}
```

### Migration

In each controller, replace the `GetMouseKeybindingsFn` method body:

```go
func (self *NotesController) GetMouseKeybindingsFn() types.MouseKeybindingsFn {
    return ListMouseBindings(ListMouseOpts{
        ViewName:     "notes",
        ClickMargin:  3,
        ItemCount:    func() int { return len(self.getContext().Items) },
        SetSelection: func(idx int) { self.getContext().SetSelectedLineIdx(idx) },
        GetContext:   func() types.Context { return self.getContext() },
        GuiCommon:    func() IGuiCommon { return self.c.GuiCommon() },
    })
}
```

`QueriesController` uses a slightly more complex `SetSelection` that branches on `CurrentTab`, but this still fits in the closure.

### Verification

- `go build ./...`
- `scripts/smoke-test.sh` (click and scroll in all three panels)

---

## Step 3: Generic `ListAdapter[T]` for Contexts

### Problem

`notesListAdapter`, `tagsListAdapter`, `queriesListAdapter`, and `parentsListAdapter` each implement the same 7 methods of `types.IList` by mechanically forwarding to a `*list` struct and a `*ListContextTrait`. This totals ~70 lines of pure boilerplate across 3 files.

### Solution

Create `context/list_adapter.go`:

```go
// ListAdapter implements types.IList by composing an ID source and a cursor source.
type ListAdapter struct {
    lenFn          func() int
    selectedIdFn   func() string
    findByIdFn     func(string) int
    cursor         func() *ListContextTrait
}

func NewListAdapter(
    lenFn func() int,
    selectedIdFn func() string,
    findByIdFn func(string) int,
    cursor func() *ListContextTrait,
) *ListAdapter {
    return &ListAdapter{
        lenFn:        lenFn,
        selectedIdFn: selectedIdFn,
        findByIdFn:   findByIdFn,
        cursor:       cursor,
    }
}

func (a *ListAdapter) Len() int                    { return a.lenFn() }
func (a *ListAdapter) GetSelectedItemId() string   { return a.selectedIdFn() }
func (a *ListAdapter) FindIndexById(id string) int { return a.findByIdFn(id) }
func (a *ListAdapter) GetSelectedLineIdx() int     { return a.cursor().GetSelectedLineIdx() }
func (a *ListAdapter) SetSelectedLineIdx(idx int)  { a.cursor().SetSelectedLineIdx(idx) }
func (a *ListAdapter) MoveSelectedLine(delta int)  { a.cursor().MoveSelectedLine(delta) }
func (a *ListAdapter) ClampSelection()             { a.cursor().ClampSelection() }

var _ types.IList = &ListAdapter{}
```

### Migration

Replace each concrete adapter with a `NewListAdapter` call in the context's `GetList()`:

```go
func (self *NotesContext) GetList() types.IList {
    return NewListAdapter(
        self.list.Len,
        self.list.GetSelectedItemId,
        self.list.FindIndexById,
        func() *ListContextTrait { return self.ListContextTrait },
    )
}
```

Delete `notesListAdapter`, `tagsListAdapter`, `queriesListAdapter`, `parentsListAdapter` structs and their method sets.

### Verification

- `go build ./...`
- `go test ./pkg/gui/context/...`

---

## Step 4: `tabIndex` Utility

### Problem

`NotesContext.TabIndex()`, `TagsContext.TabIndex()`, and `QueriesContext.TabIndex()` each implement a linear scan or switch to find the index of the current tab in a tab slice. Notes uses a `for` loop, Tags uses a `switch`, and Queries uses an `if`.

### Solution

Add a utility function to `context/` (e.g. in `list_context_trait.go` or a new `context/util.go`):

```go
// TabIndexOf returns the index of current in tabs, or 0 if not found.
func TabIndexOf[T comparable](tabs []T, current T) int {
    for i, tab := range tabs {
        if tab == current {
            return i
        }
    }
    return 0
}
```

### Migration

```go
func (self *NotesContext) TabIndex() int   { return TabIndexOf(NotesTabs, self.CurrentTab) }
func (self *TagsContext) TabIndex() int    { return TabIndexOf(TagsTabs, self.CurrentTab) }
func (self *QueriesContext) TabIndex() int { return TabIndexOf(QueriesTabs, self.CurrentTab) }
```

### Verification

- `go build ./...`
- `go test ./pkg/gui/context/...`

---

## Step 5: Move `h()` to `ControllerCommon`

### Problem

Four controllers (`NotesController`, `TagsController`, `QueriesController`, `PreviewController`) define an identical method:

```go
func (self *XController) h() *helpers.Helpers {
    return self.c.Helpers().(*helpers.Helpers)
}
```

### Solution

Add a typed accessor to `ControllerCommon`:

```go
func (self *ControllerCommon) H() *helpers.Helpers {
    return self.helpers.(*helpers.Helpers)
}
```

Then replace `self.h().X()` with `self.c.H().X()` in all four controllers and delete the per-controller `h()` methods.

**Note:** The `IHelpers` interface returns typed helpers already (`Refresh() *helpers.RefreshHelper`), so an alternative is to just call `self.c.Helpers().Notes()` directly. The `h()` accessor only exists because some call sites need the concrete `*Helpers` struct for chaining (e.g., `h().Preview().UpdatePreviewForNotes()`). Adding `H()` to `ControllerCommon` eliminates the repetition while preserving this convenience.

### Verification

- `go build ./...`

---

## Step 6: "Push Results to Preview" Helper Methods

### Problem

This sequence appears 5+ times across handler files:

```go
gui.helpers.Preview().PushNavHistory()
gui.contexts.Preview.Mode = PreviewModeCardList
gui.contexts.Preview.Cards = notes
gui.contexts.Preview.SelectedCardIndex = 0
gui.contexts.Preview.ScrollOffset = 0
gui.views.Preview.Title = title
gui.renderPreview()
gui.setContext(PreviewContext)
```

Variants exist for `PreviewModePickResults` as well. Sites include:
- `handlers_tags.go:filterByTagSearch`
- `handlers_tags.go:filterByTagPick`
- `handlers_parents.go:runQuery`
- `handlers.go:executeSearch`
- `handlers_parents.go:showComposedNote`
- `gui_common.go:SetPreviewCards`
- `gui_common.go:SetPreviewPickResults`

### Solution

Add two methods to `PreviewHelper`:

```go
// ShowCardList pushes nav history and displays a card list in the preview pane.
func (self *PreviewHelper) ShowCardList(title string, cards []models.Note) {
    self.PushNavHistory()
    pc := self.ctx()
    pc.Mode = context.PreviewModeCardList
    pc.Cards = cards
    pc.SelectedCardIndex = 0
    pc.ScrollOffset = 0
    v := self.view()
    if v != nil {
        v.Title = title
    }
    self.c.GuiCommon().RenderPreview()
}

// ShowPickResults pushes nav history and displays pick results in the preview pane.
func (self *PreviewHelper) ShowPickResults(title string, results []models.PickResult) {
    self.PushNavHistory()
    pc := self.ctx()
    pc.Mode = context.PreviewModePickResults
    pc.PickResults = results
    pc.SelectedCardIndex = 0
    pc.CursorLine = 1
    pc.ScrollOffset = 0
    v := self.view()
    if v != nil {
        v.Title = title
    }
    self.c.GuiCommon().RenderPreview()
}
```

### Migration

Replace each inline sequence with a one-liner. For example, `filterByTagSearch`:

```go
// Before:
gui.helpers.Preview().PushNavHistory()
gui.contexts.Preview.Mode = PreviewModeCardList
gui.contexts.Preview.Cards = notes
gui.contexts.Preview.SelectedCardIndex = 0
gui.views.Preview.Title = " Tag: #" + tag.Name + " "
gui.renderPreview()
gui.setContext(PreviewContext)

// After:
gui.helpers.Preview().ShowCardList(" Tag: #"+tag.Name+" ", notes)
gui.setContext(PreviewContext)
```

Also simplify `gui_common.go:SetPreviewCards` and `SetPreviewPickResults` to delegate to these helpers.

### Verification

- `go build ./...`
- `scripts/smoke-test.sh` (search, tag filter, query run, parent view)

---

## Step 7: "Confirm-then-Delete" Helper

### Problem

Four handler functions follow the same pattern:
1. Get selected item
2. Truncate/format display name
3. `showConfirm(title, message, func() error { delete; refresh })`

Sites: `deleteNote`, `deleteTag`, `deleteQuery`, `deleteParent`.

### Solution

Add to `ConfirmationHelper`:

```go
// ConfirmDelete shows a confirmation dialog and executes deleteFn on confirm.
func (self *ConfirmationHelper) ConfirmDelete(
    entityType string,
    displayName string,
    deleteFn func() error,
    refreshFn func(),
) {
    name := displayName
    if len(name) > 30 {
        name = name[:30] + "..."
    }
    self.c.GuiCommon().ShowConfirm(
        "Delete "+entityType,
        "Delete \""+name+"\"?",
        func() error {
            if err := deleteFn(); err != nil {
                self.c.GuiCommon().ShowError(err)
                return nil
            }
            refreshFn()
            return nil
        },
    )
}
```

### Migration

```go
// Before (handlers_notes.go):
title := note.Title
if title == "" { title = note.Path }
if len(title) > 30 { title = title[:30] + "..." }
gui.showConfirm("Delete Note", "Delete \""+title+"\"?", func() error {
    err := gui.ruinCmd.Note.Delete(note.UUID)
    ...
})

// After (notes_helper.go or note_actions_helper.go):
self.c.Helpers().Confirmation().ConfirmDelete("Note", note.DisplayName(), 
    func() error { return self.c.RuinCmd().Note.Delete(note.UUID) },
    func() { self.c.GuiCommon().RefreshNotes(false) },
)
```

### Verification

- `go build ./...`
- `scripts/smoke-test.sh` (delete note, delete tag, delete query, delete parent)

---

## Step 8: Unify `IGuiCommon` into Single Source of Truth

### Problem

Three separate `IGuiCommon` interfaces exist:
1. `gui_common.go` — 15 methods (used internally by gui package)
2. `controllers/controller_common.go` — 12 methods (used by controllers)
3. `helpers/helper_common.go` — ~45 methods (used by helpers)

When a new GUI capability is needed, it must be added to the right interface(s) and implemented on `*Gui`. Forgetting one causes silent breakage or forces workarounds.

### Solution

Move the authoritative `IGuiCommon` definition to `types/gui_common.go` as a superset:

```go
// types/gui_common.go
package types

type IGuiCommon interface {
    // Rendering
    Render()
    RenderAll()
    RenderNotes()
    RenderTags()
    // ... full list from helpers/helper_common.go
}
```

Then:
- `helpers/helper_common.go`: `type IGuiCommon = types.IGuiCommon`
- `controllers/controller_common.go`: Define `IGuiCommon` as a subset interface that embeds only the methods controllers need, OR use the full `types.IGuiCommon` directly.
- `gui_common.go`: Delete the local `IGuiCommon` definition. The `var _ IGuiCommon = &Gui{}` assertion uses `types.IGuiCommon`.

This is a **zero-behavioral-change** refactor — it just consolidates type definitions.

### Risk

Medium — touches many import paths. Do it in one commit with careful `go build ./...` verification. The circular import concern (helpers/controllers importing gui) is already solved by having `*Gui` implement the interface; moving the interface to `types/` doesn't change import directions.

### Verification

- `go build ./...`
- `go vet ./...`

---

## Step 9: Derive `contextToView` from Context Metadata

### Problem

`gui.go:contextToView()` is a 15-case switch that maps `ContextKey → view name`. This duplicates information already stored in each context's `BaseContext.GetPrimaryViewName()`. Adding a new context requires updating both places.

### Solution

Replace the switch with a dynamic lookup:

```go
func (gui *Gui) contextToView(key ContextKey) string {
    for _, ctx := range gui.contexts.All() {
        if ctx.GetKey() == key {
            return ctx.GetPrimaryViewName()
        }
    }
    return NotesView // fallback
}
```

Or, for O(1) lookup, add a map to `ContextTree`:

```go
// context/context_tree.go
func (self *ContextTree) ViewNameForKey(key types.ContextKey) string {
    for _, ctx := range self.All() {
        if ctx.GetKey() == key {
            return ctx.GetPrimaryViewName()
        }
    }
    return "notes"
}
```

### Verification

- `go build ./...`
- `scripts/smoke-test.sh` (exercise all panel transitions)

---

## Step 10: Consolidate `gui_common.go` Passthrough Methods

### Problem

`gui_common.go` contains ~40 one-line methods that pass through to private methods or helpers:

```go
func (gui *Gui) RenderNotes()      { gui.renderNotes() }
func (gui *Gui) RefreshNotes(p bool) { gui.helpers.Notes().FetchNotesForCurrentTab(p) }
```

Meanwhile `gui.go` has its own set of thin wrappers:

```go
func (gui *Gui) refreshNotes(preserve bool) { gui.helpers.Notes().FetchNotesForCurrentTab(preserve) }
```

This creates two layers of indirection for the same operation.

### Solution

After Step 8 (unified `IGuiCommon`), audit all passthrough methods:

1. **Direct helper calls**: Where `gui_common.go` delegates to a helper, and `gui.go` also has a private wrapper that does the same, delete the private wrapper and have all internal call sites use the public method.

2. **Embed logic in helpers**: `SetPreviewCards` and `SetPreviewPickResults` contain real logic (setting mode, cards, scroll, title, rendering). After Step 6 these delegate to `PreviewHelper.ShowCardList` / `ShowPickResults` and become true passthroughs.

3. **Remove redundant wrappers**: `gui.go:renderAll()` calls `gui.helpers.Refresh().RenderAll()`, and `gui_common.go:Render()` also calls `gui.helpers.Refresh().RenderAll()`. Keep only the public `Render()` / `RenderAll()` methods.

### Verification

- `go build ./...`
- `go test ./...`

---

## Execution Order

```
Step 4  (tabIndex utility)              ← Smallest, zero-risk warmup
Step 5  (h() → ControllerCommon)        ← Tiny, mechanical
Step 3  (generic ListAdapter)           ← Context-layer only
Step 1  (generic PopupController)       ← Controller-layer only
Step 2  (ListMouseTrait)                ← Controller-layer only
Step 7  (ConfirmDelete helper)          ← Helper-layer only
Step 6  (ShowCardList/ShowPickResults)  ← Touches handlers + helpers
Step 9  (contextToView from metadata)   ← Touches gui.go
Step 8  (unify IGuiCommon)              ← Cross-cutting, do after above stabilize
Step 10 (consolidate passthroughs)      ← Final cleanup, depends on 6 + 8
```

Steps 1-5 can be done in any order or in parallel since they touch different layers. Steps 6-7 can also be parallelized. Steps 8-10 are sequential.

Each step should be a single commit. Run `go build ./...`, `go test ./...`, and `scripts/smoke-test.sh` after each.
