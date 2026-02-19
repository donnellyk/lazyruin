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
│   │   ├── types/                   # Pure interface + data type definitions
│   │   │   ├── context.go           # Context, IBaseContext, IListContext, ContextKind
│   │   │   ├── controller.go        # IController
│   │   │   ├── binding.go           # Binding, DisabledReason, KeybindingsFn
│   │   │   ├── list.go              # IList, IListCursor
│   │   │   ├── common.go            # OnFocusOpts, OnFocusLostOpts
│   │   │   ├── completion.go        # CompletionItem, CompletionTrigger, CompletionState
│   │   │   ├── popup.go             # InputPopupConfig
│   │   │   └── dialog.go            # MenuItem
│   │   │
│   │   ├── context/                 # Context implementations (own state + identity)
│   │   │   ├── base_context.go      # BaseContext (aggregates controller bindings)
│   │   │   ├── list_cursor.go       # ListCursor implementing IListCursor
│   │   │   ├── list_context_trait.go # Shared list selection + render/preview callbacks
│   │   │   ├── global_context.go    # GlobalContext (GLOBAL_CONTEXT kind, view="")
│   │   │   ├── notes_context.go     # Owns Items []Note, cursor, CurrentTab
│   │   │   ├── tags_context.go      # Owns Items []Tag, cursor, CurrentTab
│   │   │   ├── queries_context.go   # Owns Queries + Parents, cursor, CurrentTab
│   │   │   ├── preview_context.go   # Embeds *PreviewState, owns NavHistory
│   │   │   ├── preview_state.go     # PreviewState, PreviewLink, NavEntry, PreviewMode
│   │   │   ├── search_context.go    # PERSISTENT_POPUP — search completion state
│   │   │   ├── capture_context.go   # PERSISTENT_POPUP — capture completion state
│   │   │   ├── pick_context.go      # TEMPORARY_POPUP — pick completion state
│   │   │   ├── input_popup_context.go # TEMPORARY_POPUP — generic input popup
│   │   │   ├── palette_context.go   # TEMPORARY_POPUP — command palette
│   │   │   ├── snippet_editor_context.go # TEMPORARY_POPUP — two-view snippet editor
│   │   │   ├── calendar_context.go  # TEMPORARY_POPUP — three-view calendar overlay
│   │   │   ├── contrib_context.go   # TEMPORARY_POPUP — two-view contribution chart
│   │   │   └── context_tree.go      # ContextTree: typed accessors + All()
│   │   │
│   │   ├── controllers/             # Controller implementations (own keybindings)
│   │   │   ├── base_controller.go   # Null object (all methods return nil)
│   │   │   ├── attach.go            # AttachController: wires controller to its context
│   │   │   ├── controller_common.go # ControllerCommon, IGuiCommon, IHelpers interfaces
│   │   │   ├── list_controller_trait.go # Generic nav: j/k/g/G + withItem/require
│   │   │   ├── global_controller.go # quit, search, pick, new note, focus, tab/backtab
│   │   │   ├── notes_controller.go  # list nav + enter/edit/delete/copy/tag/parent/bookmark
│   │   │   ├── tags_controller.go   # list nav + filter/rename/delete
│   │   │   ├── queries_controller.go # list nav + run/delete (queries + parents tabs)
│   │   │   ├── preview_controller.go # Thin keybinding shell; delegates to PreviewHelper
│   │   │   ├── search_controller.go # enter/esc/tab (completion)
│   │   │   ├── capture_controller.go # ctrl+s/esc/tab
│   │   │   ├── pick_controller.go   # enter/esc/tab/ctrl+a
│   │   │   ├── input_popup_controller.go # enter/esc/tab
│   │   │   ├── palette_controller.go # enter/esc; mouse click on list
│   │   │   ├── snippet_editor_controller.go # esc/tab; enter dispatched per view
│   │   │   ├── calendar_controller.go # grid h/j/k/l, input enter/esc, notes j/k
│   │   │   └── contrib_controller.go # grid h/j/k/l/enter, notes j/k
│   │   │
│   │   ├── helpers/                 # Domain operation helpers
│   │   │   ├── helpers.go           # Helpers aggregator struct + accessors
│   │   │   ├── helper_common.go     # HelperCommon, IGuiCommon interface for helpers
│   │   │   ├── refresh_helper.go    # RefreshAll, RenderAll, selection-preserving refresh
│   │   │   ├── notes_helper.go      # FetchNotesForCurrentTab, DeleteNote, tab switching
│   │   │   ├── note_actions_helper.go # AddGlobalTag, RemoveTag, SetParent, Bookmark
│   │   │   ├── tags_helper.go       # RefreshTags, tab switching
│   │   │   ├── queries_helper.go    # RefreshQueries, RefreshParents
│   │   │   ├── preview_helper.go    # Nav history, content reload, card mutations,
│   │   │   │                        #   display toggles, line ops, links, info dialog
│   │   │   ├── editor_helper.go     # SuspendAndEdit, editor command
│   │   │   ├── confirmation_helper.go # Confirm/Menu/Prompt dialogs
│   │   │   ├── search_helper.go     # ExecuteSearch, SaveQuery
│   │   │   ├── clipboard_helper.go  # CopyToClipboard
│   │   │   └── view_helper.go       # ListClickIndex, ScrollViewport (used by controllers)
│   │   │
│   │   ├── gui.go                   # Gui struct, Run, context stack, setup*Context()
│   │   ├── gui_common.go            # IGuiCommon adapter methods on *Gui
│   │   ├── state.go                 # GuiState (cross-cutting: Dialog, completion, stack)
│   │   ├── views.go                 # View name constants
│   │   ├── layout.go                # View creation and positioning
│   │   ├── commands.go              # paletteOnlyCommands() + keybinding utilities
│   │   ├── keybindings.go           # registerContextBindings(), DumpBindings()
│   │   ├── hints.go                 # Context-sensitive status bar hints
│   │   ├── statusbar.go             # Status bar rendering
│   │   ├── colors.go                # Color/style constants
│   │   │
│   │   ├── handlers.go              # Search/refresh/help/focus handlers
│   │   ├── handlers_notes.go        # Notes tab-switch and data-load handlers
│   │   ├── handlers_note_actions.go # Note action handlers (add tag, etc.)
│   │   ├── handlers_tags.go         # Tags tab-switch and data-load handlers
│   │   ├── handlers_parents.go      # Queries/Parents tab-switch and data-load handlers
│   │   ├── handlers_capture.go      # Capture (new note) handlers
│   │   ├── handlers_pick.go         # Pick popup handlers
│   │   ├── handlers_snippets.go     # Snippet editor handlers
│   │   ├── handlers_palette.go      # Command palette handlers
│   │   ├── handlers_input_popup.go  # Generic input popup handlers
│   │   │
│   │   ├── calendar.go              # Calendar overlay handlers
│   │   ├── contrib.go               # Contributions overlay handlers
│   │   │
│   │   ├── completion.go            # Completion engine, state, accept logic
│   │   ├── completion_triggers.go   # Trigger definitions per context
│   │   ├── completion_candidates.go # Candidate provider functions
│   │   ├── render.go                # List rendering (notes, tags, queries)
│   │   ├── render_preview.go        # Preview pane rendering + buildCardContent
│   │   └── dialogs.go               # Confirmation, menu, info dialogs
│   │
│   └── testutil/                    # Shared test helpers (MockExecutor)
│
├── scripts/
│   └── smoke-test.sh                # Automated TUI regression via tmux
│
└── docs/
    ├── ARCHITECTURE.md              # This file
    ├── ABSTRACTIONS.md              # Reusable abstraction patterns
    ├── KEYBINDINGS.md               # Complete keybinding reference
    └── UI_MOCKUPS.md                # Visual mockups and responsive layouts
```

## Core Components

### 1. Context System

Each panel has a **Context** that owns its state and view identity. Contexts implement `types.Context` and are stored in a typed `context.ContextTree`.

**Context kinds** (`types.ContextKind`):
- `SIDE_CONTEXT` — Notes, Tags, Queries panels
- `MAIN_CONTEXT` — Preview panel
- `PERSISTENT_POPUP` — Search, Capture (can return to previous context)
- `TEMPORARY_POPUP` — Pick, Palette, InputPopup, SnippetEditor, Calendar, Contrib (ephemeral overlays)
- `GLOBAL_CONTEXT` — Bindings that fire in any view (view name `""`)

**Context stack** (`GuiState.ContextStack []ContextKey`) replaces the old overlay enum. `pushContext()` / `popContext()` manage the stack. `popupActive()` checks whether the top-of-stack is a popup kind.

```go
// Context ownership example — NotesContext owns list items, cursor, and tab
type NotesContext struct {
    BaseContext
    *ListContextTrait
    Items      []models.Note
    CurrentTab string
}

// PreviewContext embeds all preview state + navigation history
type PreviewContext struct {
    BaseContext
    *PreviewState          // Cards, Mode, CursorLine, ScrollOffset, Links, etc.
    NavHistory []NavEntry
    NavIndex   int
}
```

### 2. Controller System

Controllers own keybindings and handlers. They implement `types.IController` and are attached to their context via `controllers.AttachController(ctrl)`.

**Null object pattern**: `baseController` implements all interface methods as no-ops. Concrete controllers override only what they need.

**Trait composition**: `ListControllerTrait[T]` provides shared j/k/g/G navigation, `withItem()` (selected-item guard), `singleItemSelected()` (disabled-reason producer), and `require()` (combining disabled reasons).

**Thin controllers**: Controllers are keybinding shells that delegate to helpers. No business logic lives in controllers.

```go
// PreviewController delegates everything to PreviewHelper
func (self *PreviewController) GetKeybindingsFn() types.KeybindingsFn {
    return func(opts types.KeybindingsOpts) []*types.Binding {
        return []*types.Binding{
            {Key: 'j', Handler: self.p().MoveDown},
            {Key: 'd', Handler: self.p().DeleteCard, Description: "Delete Card"},
            {Key: 't', Handler: self.addTag, Description: "Add Tag"},
            // ...
        }
    }
}
```

### 3. Helper Layer

Helpers encapsulate domain operations. They access the GUI through an `IGuiCommon` interface (avoiding circular imports) and are injected into controllers via `IHelpers`.

**Dependency injection chain**: `*Gui` satisfies both the helpers' and controllers' `IGuiCommon` interfaces via adapter methods in `gui_common.go`.

| Helper | Responsibility |
|--------|---------------|
| `RefreshHelper` | `RefreshAll`, `RenderAll`, selection-preserving refresh by stable ID |
| `NotesHelper` | `FetchNotesForCurrentTab`, `DeleteNote`, tab switching |
| `NoteActionsHelper` | `AddGlobalTag`, `RemoveTag`, `SetParentDialog`, `ToggleBookmark` |
| `TagsHelper` | `RefreshTags`, tab switching |
| `QueriesHelper` | `RefreshQueries`, `RefreshParents` |
| `PreviewHelper` | Nav history, content reload, card navigation, card mutations (delete/move/merge/order), display toggles, line operations (todo/done/inline tag/date), link handling, info dialog |
| `EditorHelper` | Suspend and edit in `$EDITOR` |
| `SearchHelper` | `ExecuteSearch`, `SaveQuery` |
| `ConfirmationHelper` | Confirm/menu/prompt dialog wrappers |
| `ClipboardHelper` | `CopyToClipboard` |

### 4. Keybinding Registration

`registerContextBindings()` iterates `gui.contexts.All()` and bridges controller bindings into gocui:

- Global context bindings are registered on view `""` (fires everywhere)
- Popup context bindings are NOT suppressed during overlays; main/side panel bindings ARE
- `DumpBindings()` produces a sorted, stable list for regression diffing (`--debug-bindings` flag)

### 5. Palette System

The palette aggregates entries from two sources:
1. **Controller bindings** — any `types.Binding` with a non-empty `Description` automatically appears in the palette with its key hint and category
2. **`paletteOnlyCommands()`** (`commands.go`) — tab switching and snippet management commands with no keybinding (palette-only access)

`handlers_palette.go` merges both sources into the rendered palette list.

### 6. State Management

`GuiState` holds only cross-cutting concerns that don't belong to any single context:

```go
type GuiState struct {
    ContextStack []ContextKey  // focus management
    Dialog       *DialogState  // confirmation/menu popups
    SearchQuery  string        // active search filter
    // ... completion state for search/capture/pick/palette/snippet
    // ... calendar/contrib overlay state
}
```

Panel-specific state lives in the respective context struct:
- Notes items/cursor/tab → `NotesContext`
- Tags items/cursor/tab → `TagsContext`
- Queries/Parents items/cursor/tab → `QueriesContext`
- Preview cards/mode/cursor/scroll/links/nav history → `PreviewContext`

### 7. Interface Boundaries

Two separate `IGuiCommon` interfaces prevent circular imports:

- **`helpers.IGuiCommon`** — rich interface with rendering, refresh, dialogs, context access, view access, search/preview/completion methods. Satisfied by `*Gui` adapter methods.
- **`controllers.IGuiCommon`** — minimal interface with context navigation, view access, render, refresh. Controllers access domain logic through `IHelpers`, not through IGuiCommon.

```
Controllers ──→ IHelpers ──→ Helpers ──→ IGuiCommon ──→ *Gui
     │                                                    │
     └──→ controllers.IGuiCommon ─────────────────────────┘
```

### 8. Commands Layer

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
Parse JSON → []models.Note → PreviewContext.Cards
         │
renderPreview() → replaceContext(PreviewContext)
```

### Selection Preservation on Refresh

```
RefreshTags(preserve=true)
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
- `scripts/smoke-test.sh` — tmux-driven TUI assertions (build with `go build -o /tmp/lazyruin-test`)
- `./lazyruin --debug-bindings` — dump all registered controller bindings for regression diffing
- `testutil.MockExecutor` — fluent mock for CLI command testing without a real `ruin` binary
