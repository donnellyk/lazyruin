# LazyRuin Architecture

This document describes the architecture of LazyRuin, a TUI for the `ruin` notes CLI, heavily inspired by lazygit.

## Overview

LazyRuin provides a terminal-based visual interface for managing markdown notes via the ruin CLI. It follows lazygit's architectural patterns — controllers own keybindings, contexts own state and view identity, helpers encapsulate domain operations.

## Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
│  app.go - Bootstrap, lifecycle, dependency injection            │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                          GUI Layer                               │
│  gui.go - gocui wrapper, layout, context stack management       │
│                                                                  │
│  ┌──────────────────┐  ┌──────────────────┐  ┌───────────────┐ │
│  │    Contexts       │  │   Controllers    │  │    Helpers    │ │
│  │  (own state)      │  │  (own bindings)  │  │  (domain ops) │ │
│  └──────────────────┘  └──────────────────┘  └───────────────┘ │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                       Commands Layer                             │
│  Wraps ruin CLI execution with typed Go interfaces               │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                        Models Layer                              │
│  Data structures: Note, Tag, Query, ParentBookmark               │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
lazyruin/
├── main.go                          # Entry point; CLI flags
├── pkg/
│   ├── app/
│   │   └── app.go                   # Bootstrap, vault resolution, Gui wiring
│   │
│   ├── commands/                    # ruin CLI wrappers (typed Go interfaces)
│   │   ├── ruin.go                  # Base execution, JSON parsing, Executor interface
│   │   ├── search.go                # Search operations
│   │   ├── note.go                  # Note mutations (set, append, merge)
│   │   ├── tags.go                  # Tag operations
│   │   ├── queries.go               # Saved query operations
│   │   ├── parent.go                # Parent/bookmark operations
│   │   └── pick.go                  # Pick (tag intersection) operations
│   │
│   ├── models/                      # Data structures
│   │   ├── note.go                  # Note with frontmatter fields
│   │   ├── tag.go                   # Tag with count and scope
│   │   ├── query.go                 # Saved query
│   │   └── parent.go                # Parent bookmark
│   │
│   ├── config/
│   │   └── config.go                # Configuration loading (vault path, snippets)
│   │
│   ├── gui/                         # GUI orchestration
│   │   ├── types/                   # Pure interface definitions (no implementation)
│   │   │   ├── context.go           # Context, IBaseContext, IListContext, ContextKind
│   │   │   ├── controller.go        # IController
│   │   │   ├── binding.go           # Binding, DisabledReason, KeybindingsFn
│   │   │   ├── list.go              # IList, IListCursor
│   │   │   └── common.go            # OnFocusOpts, OnFocusLostOpts
│   │   │
│   │   ├── context/                 # Context implementations (own state + identity)
│   │   │   ├── base_context.go      # BaseContext (aggregates controller bindings)
│   │   │   ├── list_cursor.go       # ListCursor implementing IListCursor
│   │   │   ├── list_context_trait.go # Shared list selection + render/preview callbacks
│   │   │   ├── global_context.go    # GlobalContext (GLOBAL_CONTEXT kind, view="")
│   │   │   ├── notes_context.go     # Owns Items []Note, cursor, CurrentTab
│   │   │   ├── tags_context.go      # Owns Items []Tag, cursor, CurrentTab
│   │   │   ├── queries_context.go   # Owns Queries + Parents, cursor, CurrentTab
│   │   │   ├── preview_context.go   # Owns Cards, ScrollOffset, CursorLine, Mode
│   │   │   ├── search_context.go    # PERSISTENT_POPUP — search completion state
│   │   │   ├── capture_context.go   # PERSISTENT_POPUP — capture completion state
│   │   │   ├── pick_context.go      # TEMPORARY_POPUP — pick completion state
│   │   │   ├── input_popup_context.go # TEMPORARY_POPUP — generic input popup
│   │   │   └── context_tree.go      # ContextTree: typed accessors + All()
│   │   │
│   │   ├── controllers/             # Controller implementations (own keybindings)
│   │   │   ├── base_controller.go   # Null object (all methods return nil)
│   │   │   ├── controller_common.go # ControllerCommon (shared deps)
│   │   │   ├── list_controller_trait.go # Generic nav: j/k/g/G + withItem/require
│   │   │   ├── global_controller.go # quit, search, pick, new note, focus, tab/backtab
│   │   │   ├── notes_controller.go  # list nav + enter/edit/delete/copy/tag/parent/bookmark
│   │   │   ├── tags_controller.go   # list nav + filter/rename/delete
│   │   │   ├── queries_controller.go # list nav + run/delete (queries + parents tabs)
│   │   │   ├── preview_controller.go # scroll/card nav + edit/delete/done/move/links
│   │   │   ├── search_controller.go # enter/esc/tab (completion)
│   │   │   ├── capture_controller.go # ctrl+s/esc/tab
│   │   │   ├── pick_controller.go   # enter/esc/tab/ctrl+a
│   │   │   └── input_popup_controller.go # enter/esc/tab
│   │   │
│   │   ├── helpers/                 # Domain operation helpers
│   │   │   ├── helpers.go           # Helpers aggregator struct
│   │   │   ├── helper_common.go     # HelperCommon (shared deps)
│   │   │   ├── refresh_helper.go    # Selection-preserving refresh (ID-based)
│   │   │   ├── notes_helper.go      # Edit/Delete/AddTag/RemoveTag/SetParent/Bookmark
│   │   │   ├── editor_helper.go     # SuspendAndEdit, editor command
│   │   │   ├── confirmation_helper.go # Confirm/Menu/Prompt dialogs
│   │   │   ├── search_helper.go     # ExecuteSearch, SaveQuery
│   │   │   └── clipboard_helper.go  # CopyToClipboard
│   │   │
│   │   ├── gui.go                   # Gui struct, Run, context stack, setup*Context()
│   │   ├── state.go                 # GuiState (cross-cutting: Dialog, NavHistory, stack)
│   │   ├── views.go                 # View name constants
│   │   ├── layout.go                # View creation and positioning
│   │   ├── commands.go              # paletteOnlyCommands() + keybinding utilities
│   │   ├── keybindings.go           # registerContextBindings(), DumpBindings()
│   │   ├── hints.go                 # Context-sensitive status bar hints
│   │   ├── statusbar.go             # Status bar rendering
│   │   ├── colors.go                # Color/style constants
│   │   │
│   │   ├── handlers.go              # Search/refresh/help/focus handlers (legacy)
│   │   ├── handlers_notes.go        # Notes panel handlers (legacy)
│   │   ├── handlers_tags.go         # Tags panel handlers (legacy)
│   │   ├── handlers_queries.go      # Queries panel handlers (legacy)
│   │   ├── handlers_parents.go      # Parent bookmark handlers (legacy)
│   │   ├── handlers_capture.go      # Capture (new note) handlers (legacy)
│   │   ├── handlers_pick.go         # Pick popup handlers (legacy)
│   │   ├── handlers_snippets.go     # Snippet editor handlers (legacy)
│   │   ├── handlers_palette.go      # Command palette handlers
│   │   ├── handlers_input_popup.go  # Generic input popup handlers (legacy)
│   │   │
│   │   ├── preview_controller.go    # Preview-specific handler methods (legacy)
│   │   ├── calendar.go              # Calendar overlay handlers
│   │   ├── contrib.go               # Contributions overlay handlers
│   │   │
│   │   ├── completion.go            # Completion engine, state, accept logic
│   │   ├── completion_triggers.go   # Trigger definitions per context
│   │   ├── completion_candidates.go # Candidate provider functions
│   │   ├── render.go                # List rendering (notes, tags, queries)
│   │   ├── render_preview.go        # Preview pane rendering
│   │   └── dialogs.go               # Confirmation, menu, info dialogs
│   │
│   └── testutil/                    # Shared test helpers (MockExecutor)
│
├── scripts/
│   └── smoke-test.sh                # Automated TUI regression via tmux (49 assertions)
│
└── docs/
    ├── ARCHITECTURE.md              # This file
    ├── KEYBINDINGS.md               # Complete keybinding reference
    ├── UI_MOCKUPS.md                # Visual mockups and responsive layouts
    └── REFACTOR_PLAN.md             # Controller refactor plan and progress
```

## Core Components

### 1. Context System

Each panel has a **Context** that owns its state and view identity. Contexts implement `types.Context` and are stored in a typed `context.ContextTree`.

**Context kinds** (`types.ContextKind`):
- `SIDE_CONTEXT` — Notes, Tags, Queries panels
- `MAIN_CONTEXT` — Preview panel
- `PERSISTENT_POPUP` — Search, Capture (can return to previous context)
- `TEMPORARY_POPUP` — Pick, Palette, InputPopup (ephemeral overlays)
- `GLOBAL_CONTEXT` — Bindings that fire in any view (view name `""`)

**Context stack** (`GuiState.ContextStack []ContextKey`) replaces the old overlay enum. `pushContext()` / `popContext()` manage the stack. `popupActive()` checks whether the top-of-stack is a popup kind.

```go
// Context ownership example
type NotesContext struct {
    BaseContext
    *ListContextTrait
    Items      []models.Note
    CurrentTab NotesTab
}
```

### 2. Controller System

Controllers own keybindings and handlers. They implement `types.IController` and are attached to their context via `controllers.AttachController(ctrl)`.

**Null object pattern**: `baseController` implements all interface methods as no-ops. Concrete controllers override only what they need.

**Trait composition**: `ListControllerTrait[T]` provides shared j/k/g/G navigation, `withItem()` (selected-item guard), `singleItemSelected()` (disabled-reason producer), and `require()` (combining disabled reasons).

```go
func (self *TagsController) GetKeybindingsFn() types.KeybindingsFn {
    return func(opts types.KeybindingsOpts) []*types.Binding {
        return append(
            self.NavBindings(), // j/k/g/G from ListControllerTrait
            &types.Binding{
                Key:         gocui.KeyEnter,
                Handler:     self.withItem(self.filterByTag),
                Description: "Filter by Tag",
                Category:    "Tags",
            },
        )
    }
}
```

### 3. Keybinding Registration

`registerContextBindings()` iterates `gui.contexts.All()` and bridges controller bindings into gocui:

- Global context bindings are registered on view `""` (fires everywhere)
- Popup context bindings are NOT suppressed during overlays; main/side panel bindings ARE
- `DumpBindings()` produces a sorted, stable list for regression diffing (`--debug-bindings` flag)

### 4. Palette System

`paletteCommands()` aggregates:
1. **Controller bindings** — any binding with a `Description` appears in the palette
2. **`paletteOnlyCommands()`** — Tab switching, Snippets management (no controller home)

Global context bindings use `Contexts: nil` (always available). Other context bindings use `Contexts: []ContextKey{ctxKey}` (filtered by active context).

### 5. Helper Layer

Helpers encapsulate domain operations and are injected into controllers:

- **`RefreshHelper`** — Preserves selection by stable ID (`GetSelectedItemId` + `FindIndexById`) across data refreshes, not raw index
- **`NotesHelper`** — Edit, delete, tag, parent, bookmark operations
- **`EditorHelper`** — Suspend and edit in `$EDITOR`
- **`ConfirmationHelper`** — Confirm/menu/prompt dialogs

### 6. State Management

`GuiState` holds only cross-cutting concerns:

```go
type GuiState struct {
    ContextStack []ContextKey  // focus management; replaces old OverlayType enum
    Dialog       *DialogState  // confirmation/menu popups
    NavHistory   []NavEntry    // preview navigation history
    // ... search/capture/pick/palette/calendar state during hybrid period
}
```

Panel-specific state (items, cursor, tab) lives in the respective context struct.

### 7. Commands Layer

Wraps ruin CLI with typed Go interfaces. All commands use `--json` output:

```go
type RuinCommand struct {
    Search  *SearchCommand
    Note    *NoteCommand
    Tags    *TagsCommand
    Queries *QueriesCommand
    Parent  *ParentCommand
    Pick    *PickCommand
}
```

The `Executor` interface enables test mocking via `testutil.MockExecutor`.

## Data Flow

### Search Flow

```
User presses / → openSearch() → pushContext(SearchContext)
         │
User types (completion triggers via completionEditor)
         │
User presses Enter → executeSearch()
         │
ruinCmd.Search.Search(query) → ruin search "<q>" --json
         │
Parse JSON → []models.Note → state.Preview.Cards
         │
renderPreview() → replaceContext(PreviewContext)
```

### Selection Preservation on Refresh

```
refreshTags(preserve=true)
    │
prevID = tagsCtx.GetSelectedItemId()  ← stable UUID
    │
tagsCtx.Items = newItems
tagsCtx.ClampSelection()
    │
newIdx = tagsCtx.FindIndexById(prevID)  ← -1 if item gone
if newIdx >= 0 → tagsCtx.SetSelectedLineIdx(newIdx)
```

## Concurrency Model

- All GUI updates and ruin CLI calls run on the main gocui goroutine
- Background refresh uses `gui.g.Update(fn)` to schedule mutations on the main loop
- Helpers doing I/O return results; mutations are applied inside the `Update` callback

## Testing

- `go test ./...` — unit tests across all packages
- `scripts/smoke-test.sh` — 49 tmux-driven TUI assertions (build with `go build -o /tmp/lazyruin-test`)
- `./lazyruin --debug-bindings` — dump all registered controller bindings for regression diffing
- `testutil.MockExecutor` — fluent mock for CLI command testing without a real `ruin` binary
