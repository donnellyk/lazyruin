# Regression Risk Refactor Plan

This plan addresses two structural regression risks not covered by `DRY_REFACTOR_PLAN.md`:

1. **Dual Context Stack** — two entry points for context navigation that diverge in behavior
2. **Dual Code Paths** — handler files that duplicate logic already in helpers, with live callers going through both paths

---

## Current State

### The dual context stack

Context navigation enters through two doors:

| Entry point | Used by | Implementation |
|---|---|---|
| `gui.setContext(ContextKey)` / `gui.pushContext` / `gui.popContext` / `gui.replaceContext` | Handler files, `commands.go`, `keybindings.go`, `calendar.go`, `contrib.go`, `handlers_palette.go` | Operates on `GuiState.ContextStack []ContextKey` (string-based). Calls `activateContext()` which sets the gocui view, re-renders all lists, refreshes data, and updates the status bar. |
| `gui.PushContext(types.Context, OnFocusOpts)` / `gui.PopContext()` / `gui.ReplaceContext(types.Context)` / `gui.PushContextByKey(key)` | Controllers, helpers, `gui_common.go` adapters | Adapter methods in `gui_common.go` that extract the `ContextKey` and forward to the string-based methods above. The `OnFocusOpts` parameter is **silently ignored**. `HandleFocus()` / `HandleFocusLost()` lifecycle hooks on `BaseContext` are **never called**. |

Both doors reach the same underlying stack, but the adapter path discards information (`OnFocusOpts`) and skips lifecycle hooks that the `types.Context` interface promises. This means:

- Focus/blur hooks registered via `AddOnFocusFn` / `AddOnFocusLostFn` will never fire.
- If a future change relies on `OnFocusOpts` (e.g., "focus came from search, so don't refresh"), it silently fails.
- Callers can't tell which path they're on — both compile and appear to work.

### The dual code paths

The handler-to-helper migration is half-done. Helpers contain the canonical implementations, but handler files still exist with their own copies. Both paths have live callers:

| Operation | Helper method | Handler method | Live handler callers |
|---|---|---|---|
| Tab switch (notes) | `NotesHelper.SwitchNotesTabByIndex` | `gui.switchNotesTabByIndex` | `commands.go` (×3), `keybindings.go` tab click |
| Tab switch (tags) | `TagsHelper.SwitchTagsTabByIndex` | `gui.switchTagsTabByIndex` | `commands.go` (×3), `keybindings.go` tab click |
| Tab switch (queries) | `QueriesHelper.SwitchQueriesTabByIndex` | `gui.switchQueriesTabByIndex` | `commands.go` (×2), `keybindings.go` tab click |
| Load notes | `NotesHelper.LoadNotesForCurrentTab` | `gui.loadNotesForCurrentTab` | `handlers_notes.go:switchNotesTabByIndex`, `handlers.go:clearSearch` |
| Fetch notes | `NotesHelper.FetchNotesForCurrentTab` | `gui.fetchNotesForCurrentTab` | `handlers_notes.go:loadNotesForCurrentTab` |
| Filter by tag | `TagsHelper.FilterByTag` | `gui.filterByTag` / `filterByTagSearch` / `filterByTagPick` | `handlers_palette.go` quick-open (×2), tests (×6) |
| Run query | `QueriesHelper.RunQuery` | `gui.runQuery` | `handlers_palette.go` quick-open (×1), tests (×3) |
| View parent | `QueriesHelper.ViewParent` | `gui.viewParent` | `handlers_palette.go` quick-open (×1) |
| Delete note | `NotesHelper.DeleteNote` | `gui.deleteNote` | tests (×2) |
| Preview for tags | `TagsHelper.UpdatePreviewForTags` | `gui.updatePreviewForTags` | `setupTagsContext` callback, `activateContext`, `handlers_tags.go:switchTagsTabByIndex` |
| Preview for queries | `QueriesHelper.UpdatePreviewForQueries` | `gui.updatePreviewForQueries` | `activateContext`, `handlers_parents.go:loadDataForQueriesTab` |
| Preview for parents | `QueriesHelper.UpdatePreviewForParents` | `gui.updatePreviewForParents` | `activateContext`, `handlers_parents.go:loadDataForQueriesTab` |
| Search execute | `SearchHelper.ExecuteSearch` | `gui.executeSearch` | `gui.go:setupSearchContext` (completion enter) |
| Search cancel | `SearchHelper.CancelSearch` | `gui.cancelSearch` | `gui.go:setupSearchContext` (completion esc), `handlers.go:executeSearch` (empty input fallback) |
| Clear search | `SearchHelper.ClearSearch` | `gui.clearSearch` | `commands.go`, `keybindings.go` |
| Open search | `SearchHelper.OpenSearch` | `gui.openSearch` | tests (×3) |
| Focus filter | `SearchHelper.FocusSearchFilter` | `gui.focusSearchFilter` | *none outside own definition — dead code* |
| Open editor | `EditorHelper.OpenInEditor` | `gui.openInEditor` | *none outside own definition — dead code* |

If a bug is fixed in one copy but not the other, the fix only applies to callers going through that path.

---

## Step 1: Wire Remaining Handler Callers to Helpers

**Goal:** Eliminate all live callers of handler methods so the handler methods become dead code.

### 1A. Rewire `commands.go` palette-only commands

`paletteOnlyCommands()` calls `gui.switchNotesTabByIndex(n)`, `gui.switchQueriesTabByIndex(n)`, `gui.switchTagsTabByIndex(n)`, and `gui.clearSearch(...)`.

Replace with helper calls:

```go
// Before:
{Name: "Notes: All", Category: "Tabs", OnRun: func() error { return gui.switchNotesTabByIndex(0) }},

// After:
{Name: "Notes: All", Category: "Tabs", OnRun: func() error { return gui.helpers.Notes().SwitchNotesTabByIndex(0) }},
```

Same for tags, queries, and clear search:

```go
// Before:
{Name: "Clear Search", ..., OnRun: func() error { return gui.clearSearch(gui.g, nil) }},

// After:
{Name: "Clear Search", ..., OnRun: func() error { gui.helpers.Search().ClearSearch(); return nil }},
```

### 1B. Rewire `keybindings.go` tab click bindings

Tab click bindings call `gui.switchNotesTabByIndex`, `gui.switchQueriesTabByIndex`, `gui.switchTagsTabByIndex`.

```go
// Before:
gui.g.SetTabClickBinding(NotesView, gui.suppressTabClickDuringDialog(gui.switchNotesTabByIndex))

// After:
gui.g.SetTabClickBinding(NotesView, gui.suppressTabClickDuringDialog(
    func(idx int) error { return gui.helpers.Notes().SwitchNotesTabByIndex(idx) },
))
```

### 1C. Rewire `keybindings.go` clear search binding

```go
// Before:
gui.g.SetKeybinding(SearchFilterView, 'x', gocui.ModNone, gui.clearSearch)

// After:
gui.g.SetKeybinding(SearchFilterView, 'x', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
    gui.helpers.Search().ClearSearch()
    return nil
})
```

### 1D. Rewire `handlers_palette.go` quick-open items

Quick-open calls `gui.runQuery`, `gui.viewParent`, `gui.filterByTagSearch`, `gui.filterByTagPick`. Replace:

```go
// Before:
return gui.runQuery(nil, nil)

// After:
return gui.helpers.Queries().RunQuery()
```

```go
// Before:
return gui.filterByTagPick(&tag)

// After:
return gui.helpers.Tags().FilterByTagPick(&tag)
```

```go
// Before:
return gui.filterByTagSearch(&tag)

// After:
return gui.helpers.Tags().FilterByTagSearch(&tag)
```

```go
// Before:
return gui.viewParent(nil, nil)

// After:
return gui.helpers.Queries().ViewParent()
```

Quick-open items also call `gui.setContext(QueriesContext)` / `gui.setContext(TagsContext)` before the helper call. The helpers already call `PushContextByKey` internally where needed, but quick-open needs to set context *before* the action (to set the active panel). Check each site — if the helper already navigates to the right context, drop the `gui.setContext` call. If not, replace with `gui.helpers.[X].PushContextByKey(...)` or keep the `gui.setContext` call (it's in the `gui` package, so it's fine to call it directly until Step 2 consolidates it).

### 1E. Rewire `gui.go:setupSearchContext` completion callbacks

The search controller's Enter/Esc/Tab callbacks call `gui.executeSearch`, `gui.cancelSearch` via the completion wrappers. Replace:

```go
// Before:
OnEnter: func() error {
    return gui.completionEnter(searchState, gui.searchTriggers, gui.executeSearch)(gui.g, gui.views.Search)
},

// After:
OnEnter: func() error {
    return gui.completionEnter(searchState, gui.searchTriggers, func(g *gocui.Gui, v *gocui.View) error {
        raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
        if !gui.helpers.Search().ExecuteSearch(raw) {
            gui.helpers.Search().CancelSearch()
        }
        return nil
    })(gui.g, gui.views.Search)
},
OnEsc: func() error {
    return gui.completionEsc(searchState, func(g *gocui.Gui, v *gocui.View) error {
        gui.helpers.Search().CancelSearch()
        return nil
    })(gui.g, gui.views.Search)
},
```

### 1F. Rewire `activateContext` to use helpers

`activateContext` calls `gui.updatePreviewForTags()`, `gui.updatePreviewForQueries()`, `gui.updatePreviewForParents()`. Replace with helper calls:

```go
// Before:
case TagsContext:
    gui.refreshTags(true)
    gui.updatePreviewForTags()

// After:
case TagsContext:
    gui.refreshTags(true)
    gui.helpers.Tags().UpdatePreviewForTags()
```

Same for queries/parents. `gui.refreshNotes(true)` already delegates to `gui.helpers.Notes().FetchNotesForCurrentTab(true)` so that's fine.

### 1G. Rewire `setupTagsContext` callback

```go
// Before:
tagsCtx := context.NewTagsContext(gui.renderTags, gui.updatePreviewForTags)

// After:
tagsCtx := context.NewTagsContext(gui.renderTags, func() { gui.helpers.Tags().UpdatePreviewForTags() })
```

### 1H. Rewire tests

Tests call handler methods directly (`tg.gui.filterByTag(...)`, `tg.gui.runQuery(...)`, etc.). Update tests to call helpers instead:

```go
// Before:
tg.gui.filterByTag(tg.g, tg.gui.views.Tags)

// After:
tg.gui.helpers.Tags().FilterByTag(tg.gui.contexts.Tags.Selected())
```

This is the largest sub-step by line count but is mechanical.

### Verification

After each sub-step:
- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh`

After all sub-steps: every handler method in `handlers_notes.go`, `handlers_tags.go`, `handlers_parents.go`, and the search/clear/focusFilter methods in `handlers.go` should have **zero callers** outside their own file.

---

## Step 2: Delete Dead Handler Methods

**Goal:** Remove the now-uncalled handler methods and their files.

### 2A. Delete handler methods

After Step 1, these methods have no callers:

**`handlers_notes.go`** (entire file):
- `switchNotesTabByIndex` — replaced by `NotesHelper.SwitchNotesTabByIndex`
- `loadNotesForCurrentTab` — replaced by `NotesHelper.LoadNotesForCurrentTab`
- `fetchNotesForCurrentTab` — replaced by `NotesHelper.FetchNotesForCurrentTab`
- `newNote` — one-liner delegating to `openCapture` (called nowhere after controller wiring)
- `deleteNote` — replaced by `NotesHelper.DeleteNote`

**`handlers_tags.go`** (entire file):
- `selectedFilteredTag` — inlined; `contexts.Tags.Selected()` is called directly
- `switchTagsTabByIndex` — replaced by `TagsHelper.SwitchTagsTabByIndex`
- `filterByTag` / `filterByTagSearch` / `filterByTagPick` — replaced by `TagsHelper.FilterByTag*`
- `renameTag` / `deleteTag` — replaced by `TagsHelper.RenameTag` / `DeleteTag`
- `updatePreviewForTags` / `updatePreviewPickResults` — replaced by `TagsHelper.UpdatePreviewForTags` / `UpdatePreviewPickResults`

**`handlers_parents.go`** (entire file):
- `switchQueriesTabByIndex` — replaced by `QueriesHelper.SwitchQueriesTabByIndex`
- `loadDataForQueriesTab` — replaced by `QueriesHelper.LoadDataForQueriesTab`
- `runQuery` / `deleteQuery` — replaced by `QueriesHelper.RunQuery` / `DeleteQuery`
- `viewParent` / `deleteParent` — replaced by `QueriesHelper.ViewParent` / `DeleteParent`
- `updatePreviewForQueries` / `updatePreviewForParents` — replaced by `QueriesHelper.*`
- `showComposedNote` — replaced by `QueriesHelper.ShowComposedNote`

**`handlers.go`** (partial — keep `quit`, `buildSearchOptions`, `showHelpHandler`, `refresh`):
- `openSearch` — replaced by `SearchHelper.OpenSearch`
- `executeSearch` — replaced by `SearchHelper.ExecuteSearch`
- `cancelSearch` — replaced by `SearchHelper.CancelSearch`
- `clearSearch` — replaced by `SearchHelper.ClearSearch`
- `focusSearchFilter` — replaced by `SearchHelper.FocusSearchFilter`
- `openInEditor` — replaced by `EditorHelper.OpenInEditor`
- `listMove` — unused after `ListControllerTrait` migration

### 2B. Delete remaining forwarding methods in `handlers.go`

`nextPanel`, `prevPanel`, `focusNotes`, `focusQueries`, `focusTags`, `focusPreview` are one-liners that delegate to `gui.globalController.*`. If they have no callers outside `handlers.go` (the bindings were migrated to `GlobalController`), delete them.

### 2C. Clean up `handlers_test.go`

After 1H rewired tests to use helpers, verify no test references the deleted handler methods. If `handlers_test.go` only tested handler methods that are now deleted, move the test logic to test helpers directly or delete the tests if coverage is already provided by helper tests + smoke tests.

### Verification

- `go build ./...` (confirms no dangling references)
- `go test ./...`
- `scripts/smoke-test.sh`
- `./lazyruin --debug-bindings` diff (unchanged)

---

## Step 3: Unify the Context Stack

**Goal:** Make `types.Context`-based navigation the single entry point, with lifecycle hooks honored.

### 3A. Move stack operations to accept `types.Context`

Rewrite `pushContext`, `popContext`, `replaceContext` to work with `types.Context` objects:

```go
// gui.go

func (gui *Gui) pushContext(ctx types.Context) {
    gui.state.ContextStack = append(gui.state.ContextStack, ctx.GetKey())

    // Call focus-lost on the outgoing context
    if prev := gui.currentContextObject(); prev != nil {
        prev.HandleFocusLost(types.OnFocusLostOpts{})
    }

    gui.activateContext(ctx.GetKey())
    ctx.HandleFocus(types.OnFocusOpts{})
}

func (gui *Gui) popContext() {
    if len(gui.state.ContextStack) <= 1 {
        return
    }
    // Call focus-lost on outgoing
    if cur := gui.currentContextObject(); cur != nil {
        cur.HandleFocusLost(types.OnFocusLostOpts{})
    }
    gui.state.ContextStack = gui.state.ContextStack[:len(gui.state.ContextStack)-1]
    gui.activateContext(gui.state.currentContext())
    // Call focus on incoming
    if next := gui.currentContextObject(); next != nil {
        next.HandleFocus(types.OnFocusOpts{})
    }
}

func (gui *Gui) replaceContext(ctx types.Context) {
    if cur := gui.currentContextObject(); cur != nil {
        cur.HandleFocusLost(types.OnFocusLostOpts{})
    }
    if len(gui.state.ContextStack) > 0 {
        gui.state.ContextStack[len(gui.state.ContextStack)-1] = ctx.GetKey()
    } else {
        gui.state.ContextStack = []types.ContextKey{ctx.GetKey()}
    }
    gui.activateContext(ctx.GetKey())
    ctx.HandleFocus(types.OnFocusOpts{})
}

// currentContextObject looks up the types.Context for the top of stack.
func (gui *Gui) currentContextObject() types.Context {
    return gui.contextByKey(gui.state.currentContext())
}
```

### 3B. Provide key-based convenience methods

The stack still stores `ContextKey` strings (changing that to `types.Context` objects would be a larger refactor and isn't necessary — keys are stable identifiers). But callers that only have a key need a lookup:

```go
func (gui *Gui) pushContextByKey(key types.ContextKey) {
    ctx := gui.contextByKey(key)
    if ctx != nil {
        gui.pushContext(ctx)
    }
}

func (gui *Gui) replaceContextByKey(key types.ContextKey) {
    ctx := gui.contextByKey(key)
    if ctx != nil {
        gui.replaceContext(ctx)
    }
}
```

### 3C. Delete `setContext`

`setContext` was an alias for `pushContext` from the legacy era. After Step 1, all callers inside the `gui` package use the helpers (which call `PushContextByKey`). Delete `setContext` entirely.

If any remaining call sites exist (e.g. in `handlers_palette.go` quick-open before Step 1D is done), replace them with `gui.pushContextByKey(key)`.

### 3D. Simplify `gui_common.go` adapter methods

The adapters now just forward to the unified methods:

```go
func (gui *Gui) PushContext(ctx types.Context, opts types.OnFocusOpts) {
    if ctx != nil {
        gui.pushContext(ctx)
    }
}

func (gui *Gui) PopContext()                            { gui.popContext() }
func (gui *Gui) ReplaceContext(ctx types.Context)       { if ctx != nil { gui.replaceContext(ctx) } }
func (gui *Gui) PushContextByKey(key types.ContextKey)  { gui.pushContextByKey(key) }
func (gui *Gui) ReplaceContextByKey(key types.ContextKey) { gui.replaceContextByKey(key) }
```

`OnFocusOpts` is still accepted by `PushContext` for interface compatibility. If needed in the future, pass it through to `HandleFocus`. For now, it's an empty struct everywhere so there's no behavioral change.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (exercise all panel transitions, popups, back-navigation)

---

## Execution Order

```
Step 1A-1G  (rewire callers)       ← Can be done in sub-step commits, each independently safe
Step 1H     (rewire tests)         ← After 1A-1G, so handler methods are only called from tests
Step 2A-2C  (delete dead code)     ← After 1H confirms zero callers
Step 3A-3D  (unify context stack)  ← After Step 2, so there's only one set of navigation code to modify
```

Steps 1A-1G are independent of each other and can be done in any order. Each sub-step should be a single commit.

---

## Risk Notes

1. **Test churn (Step 1H):** Tests that call handler methods directly are integration-style tests that create a `testGui` with a mock executor. After rewiring, they call helpers instead, but the test setup and assertions remain the same. The risk is low — the behavioral path is identical.

2. **Quick-open context sequencing (Step 1D):** Some quick-open items call `gui.setContext(X)` then `gui.runQuery(...)`. The handler's `runQuery` internally also calls `gui.setContext(PreviewContext)`. The helper's `RunQuery()` calls `gui.PushContextByKey("preview")`. Verify that the resulting context stack is the same after rewiring. The key question is whether the helper already pushes to the right panel — check each site.

3. **`activateContext` re-renders (Step 3A):** Adding `HandleFocus`/`HandleFocusLost` calls changes behavior — previously these hooks were dead code. Verify that no context has registered focus hooks before enabling them. Currently no controller sets `GetOnFocus()` or `GetOnFocusLost()` to non-nil (they all inherit the `baseController` no-op), and no setup code calls `AddOnFocusFn`, so the hooks fire but do nothing. This is safe.

4. **Completion wrapper signatures (Step 1E):** `completionEnter` expects a `func(*gocui.Gui, *gocui.View) error` callback. The helper's `ExecuteSearch` takes a `string` and returns `bool`. The wrapper must extract the text from the view and translate the bool to call `CancelSearch` on empty input.
