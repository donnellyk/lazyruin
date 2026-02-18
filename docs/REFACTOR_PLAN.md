# LazyRuin Controller Refactor Plan

This document outlines a comprehensive refactoring plan to adopt lazygit's proven
controller/context architecture. The goal is to replace the current "handlers on
`*Gui`" pattern with a proper separation of concerns using controllers, contexts,
and helpers.

The driving motivation is **reducing regressions** — updates not happening at the
right time, inconsistent UX, keyboard shortcuts failing — caused by the current
mix of business logic, state management, and UI handling in a monolithic `*Gui`.

> **Companion document**: [REFACTOR_REVIEW.md](REFACTOR_REVIEW.md) contains the
> full oracle review that informed several design decisions.

---

## Current State Analysis

### Problems with Current Architecture

1. **Monolithic Gui struct**: All handler methods live directly on `*Gui`, making it a god object
2. **Context as strings**: `ContextKey` is just a string constant, not a rich object with behavior
3. **No controller interface**: No formal contract for what a controller provides
4. **Mixed concerns**: Handlers mix input handling, business logic, state updates, and rendering
5. **Scattered keybindings**: Navigation bindings and command bindings are defined separately in two different systems
6. **Duplicated list behavior**: Each panel reimplements j/k navigation, selection, click handling
7. **No helper abstraction**: Domain operations are scattered across handler files
8. **Fragile overlay system**: `OverlayType` enum + `overlayActive()` is separate from the context system, causing suppression bugs

---

## Design Decisions

These were settled through discussion and oracle review:

| Decision | Rationale |
|----------|-----------|
| **Contexts own state directly** | Wrapping `GuiState` sub-structs would be a leaky abstraction. Contexts owning their items, cursor, and tab state ensures one source of truth. `GuiState` shrinks to cross-cutting concerns only. |
| **Controllers are the single source of truth for keybindings** | All bindings — navigation, actions, palette-worthy — are defined in controller `GetKeybindingsFn()`. `commands.go` becomes an aggregator that collects bindings for palette/hints/help. No parallel `Command` table. |
| **`paletteOnlyCommands()` for orphaned commands** | Tab switching, snippet management, and similar commands that lack a natural controller home live in a small separate list. |
| **Popups as contexts on the stack** | Replace `OverlayType` + `overlayActive()` with proper popup contexts (`PERSISTENT_POPUP`, `TEMPORARY_POPUP`). The context stack becomes the single mechanism for focus and suppression. |
| **Mouse bindings in controllers** | Mouse click, scroll, and wheel handlers are part of controller `GetMouseKeybindingsFn()`, not scattered in `keybindings.go`. |
| **Port existing `listPanel` logic into traits** | Don't redesign list behavior from scratch — the current `listPanel`, `listDown`, `listUp`, `listTop`, `listBottom`, `scrollListView` logic is proven. |
| **Multi-view contexts** | Contexts declare view names (plural), not a single view. Calendar/SnippetEditor use multiple views. A primary view name determines which view receives focus. |
| **Queries/Parents share one context** | They share one gocui view (`QueriesView`). Keep a single `QueriesContext` with `CurrentTab` that switches between query and parent data, rather than two contexts fighting over one view. |
| **Selection preservation by ID** | Background refresh can reorder items. `RefreshHelper` preserves selection using `GetSelectedItemId()` + `FindIndexById()` instead of raw index. |
| **Concurrency rule** | Helpers that touch context state must run on the UI thread (inside `gui.g.Update()`). Helpers doing I/O return results that are applied inside `Update`. |

---

## Target Architecture

```
App
 └── Gui
      ├── Contexts (own state + view + identity)
      │    ├── NotesContext (IListContext)       — owns Items, ListCursor, CurrentTab
      │    ├── TagsContext (IListContext)        — owns Items, ListCursor, CurrentTab
      │    ├── QueriesContext (IListContext)     — owns Queries+Parents, ListCursor, CurrentTab
      │    ├── PreviewContext                    — owns Cards, ScrollOffset, CursorLine, Mode
      │    ├── SearchContext (PERSISTENT_POPUP)  — owns query, completion state
      │    ├── CaptureContext (PERSISTENT_POPUP) — owns capture state, completion
      │    ├── PickContext (TEMPORARY_POPUP)     — owns pick query, completion
      │    ├── PaletteContext (TEMPORARY_POPUP)  — owns filtered commands, selection
      │    └── CalendarContext (PERSISTENT_POPUP)— owns year, month, day, notes
      │
      ├── Controllers (own keybindings + handlers)
      │    ├── GlobalController        — quit, search, pick, new note, refresh, focus shortcuts
      │    ├── NotesController          — list nav + enter/edit/delete/copy/tag/parent/bookmark
      │    ├── TagsController           — list nav + filter/rename/delete
      │    ├── QueriesController        — list nav + run/delete (queries + parents tabs)
      │    ├── PreviewController        — scroll/card nav + edit/delete/done/move/merge/links
      │    ├── SearchController         — enter/esc/tab (completion)
      │    ├── CaptureController        — ctrl+s/esc/tab (completion)
      │    ├── PickController           — enter/esc/tab/ctrl+a
      │    ├── PaletteController        — enter/esc/click
      │    └── CalendarController       — grid nav/enter/esc/tab
      │
      └── Helpers (reusable domain operations)
           ├── RefreshHelper            — refresh notes/tags/queries/parents/all/background
           ├── NotesHelper              — edit/delete/addTag/removeTag/setParent/bookmark
           ├── EditorHelper             — suspend & edit, editor command
           ├── ConfirmationHelper       — confirm/menu/prompt dialogs
           ├── SearchHelper             — execute search, save query
           └── ClipboardHelper          — copy to clipboard
```

---

## Interface Definitions

### 1. IController Interface

Controllers supply behavior to a context. They do **not** implement `HasKeybindings`
directly — they provide producer functions that the context aggregates.

```go
// pkg/gui/types/controller.go

type IController interface {
    // Context returns the context this controller is attached to
    Context() Context

    // Binding producers (return nil if not applicable)
    GetKeybindingsFn() KeybindingsFn
    GetMouseKeybindingsFn() MouseKeybindingsFn

    // Lifecycle hooks (return nil if not applicable)
    GetOnRenderToMain() func()
    GetOnFocus() func(OnFocusOpts)
    GetOnFocusLost() func(OnFocusLostOpts)
}
```

### 2. Context Interface

Contexts own state, view identity, and aggregate keybindings from attached controllers.
They are the single source of truth for "what is this panel and what can it do."

```go
// pkg/gui/types/context.go

type ContextKind int

const (
    SIDE_CONTEXT     ContextKind = iota // Notes, Tags, Queries panels
    MAIN_CONTEXT                        // Preview panel
    PERSISTENT_POPUP                    // Search, Capture, Calendar — can return to
    TEMPORARY_POPUP                     // Pick, Palette, Menus — cannot return to
    GLOBAL_CONTEXT                      // Global keybindings only
)

type ContextKey string

type Context interface {
    IBaseContext

    HandleFocus(opts OnFocusOpts)
    HandleFocusLost(opts OnFocusLostOpts)
    HandleRender()
}

type IBaseContext interface {
    GetKind() ContextKind
    GetKey() ContextKey
    IsFocusable() bool
    Title() string

    // View identity — contexts store view *names*, not *gocui.View pointers,
    // because views may be nil before layout or during resize. Views are
    // looked up at render time via the gui.
    //
    // Multi-view contexts (e.g., Calendar with grid/input/notes views)
    // return multiple names. Keybindings are registered for all of them.
    // The primary view receives focus when the context is activated.
    GetViewNames() []string
    GetPrimaryViewName() string

    // Aggregated keybindings (collected from attached controllers)
    GetKeybindings(opts KeybindingsOpts) []*Binding
    GetMouseKeybindings(opts KeybindingsOpts) []*gocui.ViewMouseBinding
    GetOnClick() func() error

    // Tab click bindings (for tabbed panels like Notes, Queries, Tags)
    GetTabClickBindingFn() func(int) error

    // Controller attachment points
    AddKeybindingsFn(KeybindingsFn)
    AddMouseKeybindingsFn(MouseKeybindingsFn)
    AddOnFocusFn(func(OnFocusOpts))
    AddOnFocusLostFn(func(OnFocusLostOpts))
    AddOnRenderToMainFn(func())
}

type IListContext interface {
    Context

    GetList() IList
    GetSelectedItemId() string
}
```

### 3. Binding Structure

```go
// pkg/gui/types/binding.go

type Binding struct {
    ID                string             // stable identity for auditing (e.g. "tags.rename")
    Key               any
    Mod               gocui.Modifier     // default: gocui.ModNone
    Handler           func() error
    Description       string             // shown in palette & help; empty = nav-only
    Tooltip           string
    Category          string             // palette grouping
    GetDisabledReason func() *DisabledReason
    DisplayOnScreen   bool               // show in status bar hints
}

type DisabledReason struct {
    Text string
}

type KeybindingsOpts struct {
    GetKey func(string) any // config lookup (future: user-configurable keys)
}

type KeybindingsFn func(opts KeybindingsOpts) []*Binding
type MouseKeybindingsFn func(opts KeybindingsOpts) []*gocui.ViewMouseBinding
```

### 4. List Interface

```go
// pkg/gui/types/list.go

type IList interface {
    IListCursor
    Len() int
    // GetSelectedItemId returns a stable ID for the selected item (e.g., UUID).
    // Used by RefreshHelper to preserve selection across data refreshes.
    GetSelectedItemId() string
    // FindIndexById locates an item by stable ID, returning -1 if not found.
    FindIndexById(id string) int
}

type IListCursor interface {
    GetSelectedLineIdx() int
    SetSelectedLineIdx(int)
    MoveSelectedLine(delta int)
    ClampSelection()
}
```

### 5. Focus & Context Stack

The context stack replaces both `ContextStack []ContextKey` and `OverlayType`:

```go
// pkg/gui/types/common.go

type OnFocusOpts struct {
    ClickedViewLineIdx      int
    ScrollSelectionIntoView bool
}

type OnFocusLostOpts struct {
    NewContextKey ContextKey
}
```

`overlayActive()` becomes a stack query:

```go
func (mgr *ContextMgr) PopupActive() bool {
    top := mgr.Current()
    kind := top.GetKind()
    return kind == PERSISTENT_POPUP || kind == TEMPORARY_POPUP
}
```

---

## Package Structure

```
pkg/gui/
├── types/                      # Interface definitions (no implementation)
│   ├── context.go              # Context, IBaseContext, IListContext, ContextKey, ContextKind
│   ├── controller.go           # IController
│   ├── binding.go              # Binding, DisabledReason, KeybindingsOpts, KeybindingsFn
│   ├── list.go                 # IList, IListCursor
│   └── common.go               # OnFocusOpts, OnFocusLostOpts
│
├── context/                    # Context implementations
│   ├── base_context.go         # BaseContext struct (aggregates keybindings, focus hooks)
│   ├── context_common.go       # ContextCommon for shared dependencies
│   ├── list_cursor.go          # ListCursor implementing IListCursor
│   ├── list_context_trait.go   # ListContextTrait (selection, render, scroll, footer)
│   ├── notes_context.go        # Owns Items []models.Note, ListCursor, CurrentTab
│   ├── tags_context.go         # Owns Items []models.Tag, ListCursor, CurrentTab
│   ├── queries_context.go      # Owns Queries + Parents, ListCursor, CurrentTab
│   ├── preview_context.go      # Owns Cards, ScrollOffset, CursorLine, Mode, Links
│   ├── search_context.go       # PERSISTENT_POPUP — query, CompletionState
│   ├── capture_context.go      # PERSISTENT_POPUP — capture state, CompletionState
│   ├── pick_context.go         # TEMPORARY_POPUP — query, AnyMode, CompletionState
│   ├── palette_context.go      # TEMPORARY_POPUP — commands, filtered, selection
│   ├── calendar_context.go     # PERSISTENT_POPUP — year, month, day, notes
│   ├── contrib_context.go      # PERSISTENT_POPUP — day counts, notes
│   └── context_tree.go         # ContextTree: typed accessors + All()
│
├── controllers/                # Controller implementations
│   ├── base_controller.go      # Null object pattern (all methods return nil)
│   ├── controller_common.go    # ControllerCommon (helpers access, context access)
│   ├── list_controller_trait.go # Generic ListControllerTrait[T] (withItem, require, nav)
│   ├── global_controller.go    # Quit, search, pick, new note, refresh, focus 1/2/3, tab/backtab
│   ├── notes_controller.go     # List nav + enter/edit/delete/copy/tag/parent/bookmark
│   ├── tags_controller.go      # List nav + filter/rename/delete
│   ├── queries_controller.go   # List nav + run/delete (queries + parents tabs)
│   ├── preview_controller.go   # Scroll/card nav + edit/delete/done/move/merge/links/history
│   ├── search_controller.go    # Enter/esc/tab
│   ├── capture_controller.go   # Ctrl+s/esc/tab
│   ├── pick_controller.go      # Enter/esc/tab/ctrl+a
│   ├── palette_controller.go   # Enter/esc/click
│   ├── calendar_controller.go  # Grid nav/enter/esc/tab/click
│   ├── contrib_controller.go   # Grid nav/enter/esc/tab
│   └── controllers.go          # SetupControllers + attachControllers wiring
│
├── helpers/                    # Domain operation helpers
│   ├── helper_common.go        # HelperCommon struct (ruin commands, gui common)
│   ├── helpers.go              # Helpers aggregator struct
│   ├── refresh_helper.go       # RefreshNotes/Tags/Queries/Parents/All/Background
│   ├── notes_helper.go         # Edit/Delete/AddTag/RemoveTag/SetParent/Bookmark
│   ├── editor_helper.go        # SuspendAndEdit, GetEditorCommand
│   ├── confirmation_helper.go  # Confirm/Menu/Prompt
│   ├── search_helper.go        # ExecuteSearch, SaveQuery
│   └── clipboard_helper.go     # CopyToClipboard
│
├── gui.go                      # Gui struct (slimmed): Run, layout, lifecycle
├── gui_common.go               # IGuiCommon interface
├── layout.go                   # View creation and positioning
├── keybindings.go              # registerContextBindings (iterates contexts)
├── commands.go                 # paletteOnlyCommands() + allBindings() aggregator
├── render.go                   # Shared rendering utilities (renderList, wrapLine, etc.)
├── views.go                    # View name constants
├── colors.go                   # Color/style constants
└── state.go                    # Slimmed GuiState (cross-cutting only)
```

---

## Slimmed GuiState

After migration, `GuiState` retains only cross-cutting concerns:

```go
type GuiState struct {
    // Context stack replaces both ContextStack and OverlayType
    // (managed by ContextMgr, not directly)

    // Navigation history (cross-cutting, spans contexts)
    NavHistory []NavEntry
    NavIndex   int

    // Dialog state (confirmation/menu popups — managed by ConfirmationHelper)
    Dialog *DialogState

    // Layout tracking
    Initialized bool
    lastWidth   int
    lastHeight  int
}
```

Everything else moves into its respective context:
- `NotesState` → `NotesContext`
- `TagsState` → `TagsContext`
- `QueriesState` + `ParentsState` → `QueriesContext` (one context, tab-switched)
- `PreviewState` → `PreviewContext`
- `SearchCompletion`, `SearchQuery` → `SearchContext`
- `CaptureCompletion`, `CaptureParent` → `CaptureContext`
- `PickCompletion`, `PickQuery`, `PickAnyMode` → `PickContext`
- `Palette` → `PaletteContext`
- `Calendar` → `CalendarContext`
- `Contrib` → `ContribContext`
- `ActiveOverlay` → replaced by context stack `PopupActive()`
- `InputPopup*` → `InputPopupContext` or folded into `ConfirmationHelper`
- `SnippetEditor*` → `SnippetEditorContext`

---

## Key Patterns

### 1. Null Object Controller

```go
type baseController struct{}

func (self *baseController) GetKeybindingsFn() types.KeybindingsFn           { return nil }
func (self *baseController) GetMouseKeybindingsFn() types.MouseKeybindingsFn { return nil }
func (self *baseController) GetOnRenderToMain() func()                       { return nil }
func (self *baseController) GetOnFocus() func(types.OnFocusOpts)             { return nil }
func (self *baseController) GetOnFocusLost() func(types.OnFocusLostOpts)     { return nil }
```

### 2. Trait Composition

```go
type NotesController struct {
    baseController
    *ListControllerTrait[*models.Note]
    c *ControllerCommon
}

var _ types.IController = &NotesController{}
```

### 3. Single Source of Truth for Bindings

```go
func (self *NotesController) GetKeybindingsFn() types.KeybindingsFn {
    return func(opts types.KeybindingsOpts) []*types.Binding {
        return []*types.Binding{
            {
                Key:               'E',
                Handler:           self.withItem(self.edit),
                GetDisabledReason: self.require(self.singleItemSelected()),
                Description:       "Open in Editor",
                Category:          "Notes",
                DisplayOnScreen:   true,
            },
            {
                Key:               'd',
                Handler:           self.withItem(self.delete),
                GetDisabledReason: self.require(self.singleItemSelected()),
                Description:       "Delete Note",
                Category:          "Notes",
            },
            // Navigation (no Description = excluded from palette)
            {Key: 'j', Handler: self.nextItem},
            {Key: 'k', Handler: self.prevItem},
            {Key: 'g', Handler: self.goTop},
            {Key: 'G', Handler: self.goBottom},
        }
    }
}

func (self *NotesController) edit(note *models.Note) error {
    return self.c.Helpers().Editor.SuspendAndEdit(note.Path)
}
```

### 4. Controller Attachment

```go
// controllers/controllers.go

func SetupControllers(common *ControllerCommon) {
    attachControllers(
        NewGlobalController(common),
        NewNotesController(common),
        NewTagsController(common),
        NewQueriesController(common),
        NewPreviewController(common),
        NewSearchController(common),
        NewCaptureController(common),
        NewPickController(common),
        NewPaletteController(common),
        NewCalendarController(common),
        NewContribController(common),
    )
}

func attachControllers(controllers ...types.IController) {
    for _, ctrl := range controllers {
        ctx := ctrl.Context()
        if f := ctrl.GetKeybindingsFn(); f != nil {
            ctx.AddKeybindingsFn(f)
        }
        if f := ctrl.GetMouseKeybindingsFn(); f != nil {
            ctx.AddMouseKeybindingsFn(f)
        }
        if f := ctrl.GetOnFocus(); f != nil {
            ctx.AddOnFocusFn(f)
        }
        if f := ctrl.GetOnFocusLost(); f != nil {
            ctx.AddOnFocusLostFn(f)
        }
        if f := ctrl.GetOnRenderToMain(); f != nil {
            ctx.AddOnRenderToMainFn(f)
        }
    }
}
```

### 5. Binding Registration (gocui bridge)

```go
// keybindings.go

func (gui *Gui) registerContextBindings() error {
    opts := types.KeybindingsOpts{}
    registered := map[string]string{} // "viewName|key" → binding.ID (conflict detection)

    for _, ctx := range gui.contexts.All() {
        // Multi-view: register bindings for ALL views the context owns
        viewNames := ctx.GetViewNames()

        for _, b := range ctx.GetKeybindings(opts) {
            binding := b // capture for closure
            handler := func(g *gocui.Gui, v *gocui.View) error {
                if gui.contexts.PopupActive() && ctx.GetKind() == types.SIDE_CONTEXT {
                    return nil // suppress during popups
                }
                if binding.GetDisabledReason != nil {
                    if reason := binding.GetDisabledReason(); reason != nil {
                        return nil
                    }
                }
                return binding.Handler()
            }

            for _, viewName := range viewNames {
                // Conflict detection: fail fast on duplicate (view, key) registration
                conflictKey := fmt.Sprintf("%s|%v", viewName, binding.Key)
                if prev, exists := registered[conflictKey]; exists {
                    return fmt.Errorf("binding conflict: %s and %s both bind key %v on view %s",
                        prev, binding.ID, binding.Key, viewName)
                }
                registered[conflictKey] = binding.ID

                if err := gui.g.SetKeybinding(viewName, binding.Key, binding.Mod, handler); err != nil {
                    return err
                }
            }
        }

        for _, viewName := range viewNames {
            for _, mb := range ctx.GetMouseKeybindings(opts) {
                if err := gui.g.SetViewMouseBinding(viewName, mb); err != nil {
                    return err
                }
            }
        }

        // Tab click bindings
        if tabFn := ctx.GetTabClickBindingFn(); tabFn != nil {
            if err := gui.g.SetTabClickBinding(ctx.GetPrimaryViewName(), tabFn); err != nil {
                return err
            }
        }
    }
    return nil
}

// DumpBindings returns a stable sorted list of all registered bindings
// for debugging and regression diffing. Use with --debug-bindings flag.
func (gui *Gui) DumpBindings() []string {
    opts := types.KeybindingsOpts{}
    var entries []string
    for _, ctx := range gui.contexts.All() {
        for _, b := range ctx.GetKeybindings(opts) {
            for _, viewName := range ctx.GetViewNames() {
                entries = append(entries, fmt.Sprintf("%-12s %-16s %-8v %s",
                    string(ctx.GetKey()), viewName, keyDisplayString(b.Key), b.ID))
            }
        }
    }
    sort.Strings(entries)
    return entries
}
```

### 6. Palette as Consumer

```go
// commands.go

func (gui *Gui) paletteCommands() []PaletteCommand {
    opts := types.KeybindingsOpts{}
    var commands []PaletteCommand

    // Collect from all controller bindings
    for _, ctx := range gui.contexts.All() {
        for _, b := range ctx.GetKeybindings(opts) {
            if b.Description == "" {
                continue // nav-only, skip
            }
            commands = append(commands, PaletteCommand{
                Name:     b.Description,
                Category: b.Category,
                Key:      keyDisplayString(b.Key),
                OnRun:    b.Handler,
            })
        }
    }

    // Add palette-only commands (no controller home)
    commands = append(commands, gui.paletteOnlyCommands()...)

    return commands
}

func (gui *Gui) paletteOnlyCommands() []PaletteCommand {
    return []PaletteCommand{
        {Name: "Notes: All", Category: "Tabs", OnRun: func() error { /* ... */ }},
        {Name: "Notes: Today", Category: "Tabs", OnRun: func() error { /* ... */ }},
        {Name: "Notes: Recent", Category: "Tabs", OnRun: func() error { /* ... */ }},
        {Name: "List Snippets", Category: "Snippets", OnRun: gui.helpers.Snippets.ListSnippets},
        {Name: "Create Snippet", Category: "Snippets", OnRun: gui.helpers.Snippets.CreateSnippet},
        // ...
    }
}
```

---

## Implementation Phases

### Migration Strategy: Foundation → Vertical Slice → Batch

To minimize the hybrid period where old and new systems coexist:

1. Build the complete foundation (types, base structs, traits, helpers)
2. Fully migrate **Tags** end-to-end as a proof-of-concept
3. Validate the architecture works (smoke tests pass for Tags)
4. Batch-migrate all remaining panels in a focused push
5. Remove all old code

---

### Phase 1: Foundation

**Goal**: All new packages compile. No existing behavior changes.

#### Tasks:

1. **`pkg/gui/types/`** — Pure interfaces, no implementation
   - [ ] `context.go` — Context, IBaseContext, IListContext, ContextKey, ContextKind
   - [ ] `controller.go` — IController
   - [ ] `binding.go` — Binding, DisabledReason, KeybindingsOpts, KeybindingsFn, MouseKeybindingsFn
   - [ ] `list.go` — IList, IListCursor
   - [ ] `common.go` — OnFocusOpts, OnFocusLostOpts

2. **`pkg/gui/context/`** — Base implementations
   - [ ] `base_context.go` — BaseContext (aggregates keybinding fns, focus hooks, view names — not pointers)
   - [ ] `list_cursor.go` — ListCursor struct implementing IListCursor
   - [ ] `list_context_trait.go` — Ported from existing `listPanel` logic: selection, render, scroll, footer

3. **`pkg/gui/controllers/`** — Base implementations
   - [ ] `base_controller.go` — Null object (all methods return nil)
   - [ ] `controller_common.go` — ControllerCommon (helpers access, contexts access)
   - [ ] `list_controller_trait.go` — Generic `ListControllerTrait[T]` with `withItem()`, `require()`, `singleItemSelected()`, nav handlers. Ported from existing `listDown`/`listUp`/`listTop`/`listBottom`.

4. **`pkg/gui/helpers/`** — Domain operation extraction
   - [ ] `helper_common.go` — HelperCommon (ruin commands, gui common ref)
   - [ ] `helpers.go` — Helpers aggregator struct
   - [ ] `refresh_helper.go` — `RefreshNotes/Tags/Queries/Parents/All/BackgroundRefresh`; uses `GetSelectedItemId()` + `FindIndexById()` to preserve selection by ID, not raw index
   - [ ] `notes_helper.go` — `EditNote/DeleteNote/AddTag/RemoveTag/SetParent/ToggleBookmark`
   - [ ] `editor_helper.go` — `SuspendAndEdit/GetEditorCommand`
   - [ ] `confirmation_helper.go` — `Confirm/Menu/Prompt`
   - [ ] `search_helper.go` — `ExecuteSearch/SaveQuery`
   - [ ] `clipboard_helper.go` — `CopyToClipboard`

5. **`pkg/gui/gui_common.go`** — IGuiCommon interface
   ```go
   type IGuiCommon interface {
       Contexts() *context.ContextTree
       PushContext(ctx types.Context, opts types.OnFocusOpts)
       PopContext()
       ReplaceContext(ctx types.Context)  // for flows like search→preview
       CurrentContext() types.Context
       PopupActive() bool

       GetView(name string) *gocui.View  // lookup views by name at render time
       Render()
       Update(func() error)
       Suspend() error
       Resume() error
   }
   ```

**Deliverable**: All new packages compile. Old code untouched.

---

### Phase 2: Tags Vertical Slice

**Goal**: Fully migrate Tags panel end-to-end. Old Tags handlers removed. Smoke test passes.

This validates the entire architecture before committing to the rest.

#### Tasks:

1. **TagsContext** (`context/tags_context.go`)
   - [ ] Owns `Items []models.Tag`, `*ListCursor`, `CurrentTab TagsTab`
   - [ ] Embeds `BaseContext` + `ListContextTrait`
   - [ ] Implements `GetSelected() *models.Tag`
   - [ ] Implements `HandleRender()` — ported from `renderTags()`
   - [ ] Implements `HandleFocus()` / `HandleFocusLost()`
   - [ ] Handles tab filtering (All/Global/Inline)

2. **TagsController** (`controllers/tags_controller.go`)
   - [ ] Embeds `baseController` + `ListControllerTrait[*models.Tag]`
   - [ ] `GetKeybindingsFn()` returns all bindings:
     - j/k/g/G/arrows (nav, no Description)
     - Enter (filter by tag)
     - r (rename tag)
     - d (delete tag)
   - [ ] `GetMouseKeybindingsFn()` returns click + wheel bindings
   - [ ] `GetOnFocus()` — refreshes tags, updates preview
   - [ ] `GetOnRenderToMain()` — updates preview for selected tag
   - [ ] Handlers call helpers (e.g., `self.c.Helpers().Refresh.RefreshTags()`)

3. **Wire up**
   - [ ] Add `TagsContext` to `ContextTree`
   - [ ] Add `TagsController` to `SetupControllers`
   - [ ] Update `registerContextBindings()` to include tags context bindings
   - [ ] Remove tags entries from old `commands()` table
   - [ ] Remove `tagsNavBindings()` from old `keybindings.go`

4. **Remove old code**
   - [ ] Remove `handlers_tags.go`
   - [ ] Remove `TagsState` from `GuiState`
   - [ ] Remove `tagsPanel()` from `handlers_notes.go` (if referenced)

5. **Validate**
   - [ ] `go build ./...`
   - [ ] `go test ./...`
   - [ ] Smoke test passes (tags section)
   - [ ] Manual verification: j/k nav, Enter filter, r rename, d delete, tab switch, mouse click
   - [ ] `DumpBindings()` output includes all old tags bindings (diff against old system)
   - [ ] Palette shows Tags commands (filter, rename, delete) — validates aggregator pipeline
   - [ ] Binding conflict detection fires correctly on deliberate duplicate (then remove)

**Deliverable**: Tags panel fully migrated. Architecture proven. Palette integration validated.

---

### Phase 3: Batch Migration

**Goal**: Migrate all remaining panels. Remove all old handler code. Full cutover.

This is mechanical — the patterns are proven from Phase 2. Split into sub-phases
by risk level, smoke testing between each.

> **Transitional palette**: During batch migration, `paletteCommands()` must
> consume both new controller bindings AND not-yet-migrated legacy `commands()`
> entries. Add a transitional aggregator that merges both sources, removing
> legacy entries as each panel migrates. This prevents the palette from breaking
> mid-migration.

#### 3a. List Panels (Notes, Queries)

For each, follow the Tags pattern:

**NotesContext + NotesController**
- [ ] Context owns `Items []models.Note`, `*ListCursor`, `CurrentTab NotesTab`
- [ ] Controller bindings: j/k/g/G, Enter (view), E (edit), d (delete), y (copy), t/T (tag), > (parent), P (remove parent), b (bookmark), s (info)
- [ ] `GetOnRenderToMain()` updates preview for selected note
- [ ] Remove `handlers_notes.go`, `handlers_note_actions.go`

**QueriesContext + QueriesController** (includes Parents tab)
- [ ] Context owns `Queries []models.Query`, `Parents []models.ParentBookmark`, `*ListCursor`, `CurrentTab QueriesTab`
- [ ] Tab switching changes which data slice the `ListContextTrait` indexes into
- [ ] Controller bindings: j/k/g/G, Enter (run query or parent), d (delete query)
- [ ] Remove `handlers_queries.go`, `handlers_parents.go`

**Checkpoint**: Smoke test + `DumpBindings()` diff after 3a.

#### 3b. Preview Panel

**PreviewContext + PreviewController**
- [ ] Context owns all `PreviewState` fields (Cards, ScrollOffset, CursorLine, Mode, Links, etc.)
- [ ] Controller bindings: j/k (scroll), J/K (card nav), }/{  (headers), l/L (links), d/E/D/m/M/f/v/o/x (actions), Enter, Esc, [/]
- [ ] Ported from existing `preview_controller.go` and `handlers_preview.go` (already partially extracted)
- [ ] Remove old `PreviewState` from `GuiState`

**Checkpoint**: Smoke test + `DumpBindings()` diff after 3b.

#### 3c. Popup Contexts (replace OverlayType) — highest integration risk

For each popup, create a context + controller pair:

**SearchContext (PERSISTENT_POPUP) + SearchController**
- [ ] Context owns: query string, `*CompletionState`
- [ ] Controller bindings: Enter (execute), Esc (cancel), Tab (completion)
- [ ] Remove `handlers.go` search methods, `searchBindings()`

**CaptureContext (PERSISTENT_POPUP) + CaptureController**
- [ ] Context owns: capture text, `*CompletionState`, `*CaptureParentInfo`
- [ ] Controller bindings: Ctrl+S (submit), Esc (cancel), Tab (completion)
- [ ] Remove `handlers_capture.go`, `captureBindings()`

**PickContext (TEMPORARY_POPUP) + PickController**
- [ ] Context owns: query, AnyMode, SeedHash, `*CompletionState`
- [ ] Controller bindings: Enter, Esc, Tab, Ctrl+A
- [ ] Remove `handlers_pick.go`, `pickBindings()`

**PaletteContext (TEMPORARY_POPUP) + PaletteController**
- [ ] Context owns: commands, filtered, selection, filter text
- [ ] Controller bindings: Enter, Esc, click
- [ ] Remove `handlers_palette.go`, `paletteBindings()`

**InputPopupContext (TEMPORARY_POPUP) + InputPopupController**
- [ ] Context owns: config, `*CompletionState`
- [ ] Remove `handlers_input_popup.go`, `inputPopupBindings()`

**SnippetEditorContext (TEMPORARY_POPUP) + SnippetEditorController**
- [ ] **Multi-view**: `GetViewNames()` returns `[SnippetNameView, SnippetExpansionView]`
- [ ] Context owns: focus (which sub-view), `*CompletionState`
- [ ] Remove `handlers_snippets.go`, `snippetEditorBindings()`

**CalendarContext (PERSISTENT_POPUP) + CalendarController**
- [ ] **Multi-view**: `GetViewNames()` returns `[CalendarGridView, CalendarInputView, CalendarNotesView]`
- [ ] `GetPrimaryViewName()` returns `CalendarGridView` (default focus)
- [ ] Context owns: year, month, day, focus, notes, note index
- [ ] Remove `calendarBindings()`

**ContribContext (PERSISTENT_POPUP) + ContribController**
- [ ] **Multi-view**: `GetViewNames()` returns `[ContribGridView, ContribNotesView]`
- [ ] Context owns: day counts, selected date, focus, notes
- [ ] Remove `contribBindings()`

**Checkpoint**: Smoke test all popups. Verify suppression: global keys (q, /, n) must NOT fire while any popup is open.

#### 3d. Replace Overlay System

- [ ] Remove `OverlayType` enum and all `Overlay*` constants from `state.go`
- [ ] Remove `openOverlay()`, `closeOverlay()`, `overlayActive()` from `gui.go`
- [ ] Replace with `ContextMgr.PopupActive()` (checks if top context is a popup kind)
- [ ] Update `suppressDuringDialog` to use `PopupActive()`

#### 3e. Global Controller

**GlobalController**
- [ ] Bindings: q/Ctrl+C (quit), / (search), p/\\ (pick), n (new note), Ctrl+R (refresh), ? (help), : (palette), c (calendar), C (contrib), 1/2/3 (focus), 0 (search filter), Tab/Backtab (next/prev panel)
- [ ] Remove `handlers.go` global methods

#### 3f. Final Wiring & Cleanup

- [ ] Update `gui.go`: add `contexts *context.ContextTree`, `helpers *helpers.Helpers`
- [ ] Replace `setupKeybindings()` with `registerContextBindings()`
- [ ] Replace `commands()` with `paletteCommands()` + `paletteOnlyCommands()`
- [ ] Remove transitional palette aggregator (all panels migrated, no more legacy commands)
- [ ] Slim `GuiState` to cross-cutting concerns only
- [ ] Delete all `handlers_*.go` files
- [ ] Delete old `keybindings.go` nav binding functions
- [ ] Update `activateContext()` / `pushContext()` / `popContext()` to use `ContextMgr`
- [ ] Ensure `backgroundRefresh()` wraps all context state mutations inside `gui.g.Update()` — helpers doing I/O return results, applied on UI thread
- [ ] Add `--debug-bindings` CLI flag that prints `DumpBindings()` output and exits
- [ ] Final `DumpBindings()` diff against complete old binding set — every old binding must be accounted for

**Deliverable**: Full cutover. No old handler code remains.

---

### Phase 4: Testing & Polish

**Goal**: Comprehensive validation, documentation update.

#### Tasks:

1. **Smoke test** (`scripts/smoke-test.sh`)
   - [ ] Verify all keybindings across all panels
   - [ ] Verify all popups open/close correctly
   - [ ] Verify popup suppression (global keys don't fire during popups)
   - [ ] Verify tab switching on all tabbed panels
   - [ ] Verify mouse click and wheel on all panels

2. **Unit tests**
   - [ ] `ListCursor` — selection, clamp, move
   - [ ] `ListControllerTrait` — `withItem`, `singleItemSelected`, `require`
   - [ ] `ListContextTrait` — render, scroll, footer
   - [ ] Helpers — with mocked `commands.RuinCommand`
   - [ ] `ContextMgr` — push, pop, `PopupActive()`

3. **Documentation**
   - [ ] Update `docs/ARCHITECTURE.md` to reflect new structure
   - [ ] Update `docs/KEYBINDINGS.md` if any bindings changed
   - [ ] Update `CLAUDE.md` with new package structure

4. **Remove dead code**
   - [ ] Verify no orphaned exports
   - [ ] Run `golangci-lint run`

**Deliverable**: Clean, tested, documented codebase.

---

## Risks and Guardrails

| Risk | Guardrail |
|------|-----------|
| Phase 3 batch is large | Patterns are proven from Phase 2. Each sub-phase (3a→3b→3c→3d→3e→3f) has a checkpoint. Smoke test between each. |
| Popup context stack breaks overlay suppression | Test popup suppression explicitly in smoke test. Verify global keys (q, /, n) don't fire during popups. Compare behavior before/after for each popup type. |
| Binding registration misses a keybinding | `DumpBindings()` produces a diffable list. Compare old vs new after each sub-phase. Every old binding must be accounted for. |
| Duplicate binding registration | Conflict detection in `registerContextBindings()` fails fast with a clear error identifying both conflicting bindings by ID. |
| State migration loses data | Each context's constructor should initialize the same defaults as `NewGuiState()`. Compare field-by-field. |
| gocui handler signature mismatch | The bridge in `registerContextBindings()` wraps `func() error` → `func(*gocui.Gui, *gocui.View) error`. Mouse bindings keep gocui's native signature. |
| Multi-view popup regressions | Calendar/SnippetEditor contexts declare all their view names. Bindings registered for each. Test focus switching between sub-views (grid↔input↔notes). |
| Selection drift on refresh | `RefreshHelper` preserves selection by stable ID (`GetSelectedItemId()` + `FindIndexById()`), not raw index. Background refresh can reorder items. |
| Nil view access during startup/resize | Contexts store view *names*, not `*gocui.View` pointers. Views are looked up at render time via `IGuiCommon.GetView(name)`. |
| Concurrency: background refresh mutates context state | All context state mutations must happen inside `gui.g.Update()`. Helpers doing I/O return results that are applied on the UI thread. |
| Palette breaks during partial migration | Transitional palette aggregator merges new controller bindings + legacy `commands()` entries. Legacy entries removed as each panel migrates. |

---

## Success Criteria

1. ✅ All keybindings work as before (verified by smoke test + `DumpBindings()` diff)
2. ✅ No handler methods directly on `*Gui` (except lifecycle)
3. ✅ Each panel has its own Context (owns state) + Controller (owns bindings)
4. ✅ Domain operations live in Helpers (called by both controllers and palette)
5. ✅ List navigation uses shared traits (zero duplication)
6. ✅ `OverlayType` removed — popups are contexts on the stack
7. ✅ `commands.go` is a consumer of controller bindings, not a parallel source of truth
8. ✅ Adding a new panel requires only: Context, Controller, wire up
9. ✅ Smoke tests pass
10. ✅ Code is more testable (helpers can be unit tested with mocked commands)
11. ✅ No duplicate binding conflicts (enforced by conflict detection at startup)
12. ✅ Multi-view popups (Calendar, SnippetEditor, Contrib) work correctly
13. ✅ Selection preserved by ID across background refreshes
14. ✅ `--debug-bindings` flag available for regression diffing
