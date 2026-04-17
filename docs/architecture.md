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
│  Data structures: Note, Tag, Query, ParentBookmark, PickResult    │
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
│   │   ├── parent.go                # Parent bookmark
│   │   └── pick.go                  # PickMatch + PickResult (tag intersection results)
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
│   │   │   ├── gui_common.go        # IGuiCommon interface
│   │   │   ├── completion.go        # CompletionItem, CompletionTrigger, CompletionState
│   │   │   ├── palette.go           # PaletteCommand, PaletteState
│   │   │   ├── popup.go             # InputPopupConfig
│   │   │   └── dialog.go            # MenuItem
│   │   │
│   │   ├── context/                 # Context implementations (own state + identity)
│   │   │   ├── base_context.go      # BaseContext (aggregates controller bindings, focus hooks)
│   │   │   ├── list_cursor.go       # ListCursor implementing IListCursor
│   │   │   ├── list_context_trait.go # Shared list selection + render/preview callbacks
│   │   │   ├── preview_common.go    # PreviewNavState, PreviewDisplayState, IPreviewContext
│   │   │   ├── global_context.go    # GlobalContext (GLOBAL_CONTEXT kind, view="")
│   │   │   ├── notes_context.go     # Owns Items []Note, cursor, CurrentTab
│   │   │   ├── tags_context.go      # Owns Items []Tag, cursor, CurrentTab
│   │   │   ├── queries_context.go   # Owns Queries + Parents, cursor, CurrentTab
│   │   │   ├── preview_context.go   # Embeds *PreviewState, implements Snapshotter
│   │   │   ├── preview_state.go     # PreviewState, PreviewLink, PreviewMode
│   │   │   ├── datepreview_context.go # MAIN_CONTEXT — date preview with 3 sections (tags, todos, notes)
│   │   │   ├── search_context.go    # PERSISTENT_POPUP — search completion state
│   │   │   ├── capture_context.go   # PERSISTENT_POPUP — capture state + completion
│   │   │   ├── pick_context.go      # TEMPORARY_POPUP — pick state + completion
│   │   │   ├── input_popup_context.go # TEMPORARY_POPUP — input popup state + completion
│   │   │   ├── palette_context.go   # TEMPORARY_POPUP — palette state
│   │   │   ├── snippet_editor_context.go # TEMPORARY_POPUP — snippet editor state + completion
│   │   │   ├── calendar_context.go  # TEMPORARY_POPUP — calendar state (year/month/day/notes)
│   │   │   ├── contrib_context.go   # TEMPORARY_POPUP — contribution chart state
│   │   │   └── context_tree.go      # ContextTree: typed accessors + All() + ActivePreviewKey
│   │   │
│   │   ├── controllers/             # Controller implementations (own keybindings)
│   │   │   ├── base_controller.go   # Null object (all methods return nil)
│   │   │   ├── attach.go            # AttachController: wires controller to its context
│   │   │   ├── controller_common.go # ControllerCommon, IHelpers interfaces
│   │   │   ├── list_controller_trait.go # Generic nav: j/k/g/G + withItem/require
│   │   │   ├── global_controller.go # quit, search, pick, new note, focus, tab/backtab
│   │   │   ├── notes_controller.go  # list nav + enter/edit/delete/copy/tag/parent/bookmark
│   │   │   ├── tags_controller.go   # list nav + filter/rename/delete
│   │   │   ├── queries_controller.go # list nav + run/delete (queries + parents tabs)
│   │   │   ├── preview_controller.go # keybinding shell + global scroll; delegates to PreviewHelper
│   │   │   ├── search_controller.go # enter/esc/tab (completion)
│   │   │   ├── capture_controller.go # ctrl+s/esc/tab
│   │   │   ├── pick_controller.go   # enter/esc/tab/ctrl+a
│   │   │   ├── input_popup_controller.go # enter/esc/tab
│   │   │   ├── palette_controller.go # enter/esc; mouse click on list
│   │   │   ├── snippet_editor_controller.go # esc/tab; enter dispatched per view
│   │   │   ├── calendar_controller.go # grid h/j/k/l, input enter/esc, notes j/k
│   │   │   ├── contrib_controller.go # grid h/j/k/l/enter, notes j/k
│   │   │   └── datepreview_controller.go # section nav )/( + PreviewNavTrait (card/line/header)
│   │   │
│   │   ├── helpers/                 # Domain operation helpers
│   │   │   ├── helpers.go           # Helpers aggregator struct + accessors
│   │   │   ├── helper_common.go     # HelperCommon, IGuiCommon interface for helpers
│   │   │   ├── refresh_helper.go    # RefreshAll, RenderAll, selection-preserving refresh
│   │   │   ├── notes_helper.go      # FetchNotesForCurrentTab, DeleteNote, tab switching
│   │   │   ├── note_actions_helper.go # AddGlobalTag, RemoveTag, SetParent, Bookmark
│   │   │   ├── tags_helper.go       # RefreshTags, tab switching
│   │   │   ├── queries_helper.go    # RefreshQueries, RefreshParents
│   │   │   ├── preview_helper.go    # Content reload, card mutations,
│   │   │   │                        #   display toggles, line ops, links, info dialog
│   │   │   ├── navigator.go         # Navigator helper: NavigateTo, ShowHover, ReplaceCurrent, Back, Forward
│   │   │   ├── editor_helper.go     # SuspendAndEdit, editor command
│   │   │   ├── confirmation_helper.go # Confirm/Menu/Prompt dialogs
│   │   │   ├── search_helper.go     # ExecuteSearch, SaveQuery
│   │   │   ├── clipboard_helper.go  # CopyToClipboard
│   │   │   ├── capture_helper.go    # OpenCapture, SubmitCapture, CancelCapture
│   │   │   ├── pick_helper.go       # OpenPick, ExecutePick, TogglePickAny
│   │   │   ├── input_popup_helper.go # OpenInputPopup, HandleEnter, HandleEsc
│   │   │   ├── snippet_helper.go    # ListSnippets, CreateSnippet, DeleteSnippet, SaveSnippet
│   │   │   ├── datepreview_helper.go # LoadDatePreview, ReloadDatePreview, date/pick utilities
│   │   │   ├── preview_nav_helper.go # Shared preview nav: card/line/header/section, Enter, links
│   │   │   └── view_helper.go       # ListClickIndex, ScrollViewport (used by controllers)
│   │   │
│   │   ├── gui.go                   # Gui struct, Run, context stack, setup*Context()
│   │   ├── gui_common.go            # IGuiCommon adapter methods on *Gui
│   │   ├── state.go                 # GuiState (cross-cutting: Dialog, search, stack)
│   │   ├── views.go                 # View name constants
│   │   ├── layout.go                # View creation and positioning
│   │   ├── commands.go              # paletteOnlyCommands() + keybinding utilities
│   │   ├── keybindings.go           # registerContextBindings(), DumpBindings()
│   │   ├── hints.go                 # Context-sensitive status bar hints
│   │   ├── statusbar.go             # Status bar + tab sync rendering
│   │   ├── colors.go                # Color/style constants
│   │   ├── handlers.go              # Search options + quit/refresh handlers
│   │   ├── palette.go               # Palette rendering and filtering
│   │   ├── calendar.go              # Calendar overlay rendering and navigation
│   │   ├── contrib.go               # Contribution chart rendering and navigation
│   │   ├── completion.go            # Completion engine, state, accept logic
│   │   ├── completion_triggers.go   # Trigger definitions per context
│   │   ├── completion_candidates.go # Candidate provider functions
│   │   ├── render.go                # List rendering (notes, tags, queries)
│   │   ├── render_preview.go        # Preview pane rendering + BuildCardContent
│   │   └── dialogs.go               # Confirmation, menu, info dialogs
│   │
│   └── testutil/                    # Shared test helpers (MockExecutor)
│
├── scripts/
│   ├── smoke-test.sh                # Automated TUI regression via tmux (60 assertions, 24 sections)
│   └── keybinding-test.sh           # Keyboard shortcut smoke test (90 assertions, 69 sections)
│
└── docs/
    ├── architecture.md              # This file
    ├── abstractions.md              # Reusable abstraction patterns
    └── keybindings.md               # Complete keybinding reference
```

## Core Components

### 1. Context System

Each panel has a **Context** that owns its state and view identity. Contexts implement `types.Context` and are stored in a typed `context.ContextTree`.

**Context kinds** (`types.ContextKind`):
- `SIDE_CONTEXT` — Notes, Tags, Queries panels
- `MAIN_CONTEXT` — Preview, DatePreview panels
- `PERSISTENT_POPUP` — Search, Capture (can return to previous context)
- `TEMPORARY_POPUP` — Pick, Palette, InputPopup, SnippetEditor, Calendar, Contrib (ephemeral overlays)
- `GLOBAL_CONTEXT` — Bindings that fire in any view (view name `""`)

**Context stack** (`GuiState.ContextStack []ContextKey`) manages focus. `pushContext()` / `popContext()` manage the stack. `popupActive()` uses `GetKind()` to check whether the top-of-stack is a popup.

**Focus hooks**: Each context can register `HandleFocus` callbacks via `AddOnFocusFn()`. When a context gains focus, its hooks fire — this drives per-panel refresh and preview updates without a central switch statement.

```go
// Context ownership example — NotesContext owns list items, cursor, and tab
type NotesContext struct {
    BaseContext
    *ListContextTrait
    Items      []models.Note
    CurrentTab string
}

// Popup contexts own their runtime state
type CalendarContext struct {
    BaseContext
    State *CalendarState  // Year, Month, SelectedDay, Focus, Notes, NoteIndex
}

// DatePreviewContext implements IPreviewContext for date-based preview
type DatePreviewContext struct {
    BaseContext
    TargetDate string
    TagPicks   []PickResult     // Inline tag matches (done sorted last)
    TodoPicks  []PickResult     // Checkbox matches
    Notes      []Note           // Created + updated (deduplicated)
}
```

**Multiple preview contexts**: Preview and DatePreview both share the `preview` view and implement `IPreviewContext`. The active context is tracked by `ActivePreviewKey` in `ContextTree`. Both share `PreviewNavState` for scroll/cursor/links. Back/forward navigation is managed by the `Navigator` helper and `NavigationManager` history stack (see section 10).

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

**Dependency injection chain**: `*Gui` satisfies the `IGuiCommon` interfaces via adapter methods in `gui_common.go`. Methods that are pure implementations (rendering, completion candidates, etc.) live in their source files as public methods on `*Gui`.

| Helper | Responsibility |
|--------|---------------|
| `RefreshHelper` | `RefreshAll`, `RenderAll`, selection-preserving refresh by stable ID |
| `NotesHelper` | `FetchNotesForCurrentTab`, `DeleteNote`, tab switching |
| `NoteActionsHelper` | `AddGlobalTag`, `RemoveTag`, `SetParentDialog`, `ToggleBookmark` |
| `TagsHelper` | `RefreshTags`, tab switching |
| `QueriesHelper` | `RefreshQueries`, `RefreshParents` |
| `PreviewHelper` | Content reload, card navigation, card mutations (delete/move/merge/order), display toggles, line operations (todo/done/inline tag/date), link handling, info dialog, scroll |
| `Navigator` | Single entry point for all preview pane transitions: `NavigateTo`, `ShowHover`, `ReplaceCurrent`, `Back`, `Forward`; owns capture-on-departure logic |
| `PreviewNavHelper` | Shared preview navigation: card/line/header/section jump, Enter handler, link highlight — works across Preview and DatePreview via `IPreviewContext` |
| `DatePreviewHelper` | `LoadDatePreview`, `ReloadDatePreview`, date utilities (`CurrentWeekday`, `ISOWeekday`), `DeduplicateNotes`, `filterOutTodoLines`, `sortDonePicksLast` |
| `EditorHelper` | Suspend and edit in `$EDITOR` |
| `SearchHelper` | `ExecuteSearch`, `SaveQuery` |
| `ConfirmationHelper` | Confirm/menu/prompt dialog wrappers |
| `ClipboardHelper` | `CopyToClipboard` |
| `CaptureHelper` | `OpenCapture`, `SubmitCapture`, `CancelCapture` |
| `PickHelper` | `OpenPick`, `ExecutePick`, `TogglePickAny` |
| `InputPopupHelper` | `OpenInputPopup`, `HandleEnter`, `HandleEsc` |
| `SnippetHelper` | `ListSnippets`, `CreateSnippet`, `DeleteSnippet`, `SaveSnippet` |

### 4. Keybinding Registration

`registerContextBindings()` iterates `gui.contexts.All()` and bridges controller bindings into gocui:

- Global context bindings are registered on view `""` (fires everywhere)
- Popup context bindings are NOT suppressed during overlays; main/side panel bindings ARE
- `DumpBindings()` produces a sorted, stable list for regression diffing (`--debug-bindings` flag)

### 5. Palette System

The palette aggregates entries from two sources:
1. **Controller bindings** — any `types.Binding` with a non-empty `Description` automatically appears in the palette with its key hint and category
2. **`paletteOnlyCommands()`** (`commands.go`) — tab switching and snippet management commands with no keybinding (palette-only access)

`palette.go` merges both sources into the rendered palette list.

### 6. State Management

`GuiState` holds only cross-cutting concerns that don't belong to any single context:

```go
type GuiState struct {
    ContextStack     []ContextKey     // focus management
    Dialog           *DialogState     // confirmation/menu popups
    SearchQuery      string           // active search filter
    SearchCompletion *CompletionState // search completion state
}
```

All panel-specific and popup-specific state lives in the respective context structs:
- Notes items/cursor/tab → `NotesContext`
- Tags items/cursor/tab → `TagsContext`
- Queries/Parents items/cursor/tab → `QueriesContext`
- Preview cards/mode/cursor/scroll/links → `PreviewContext`; nav history → `NavigationManager` (via `Navigator` helper)
- Date preview target date/tag picks/todo picks/notes/section ranges → `DatePreviewContext`
- Capture parent/completion → `CaptureContext`
- Pick query/anyMode/completion → `PickContext`
- Input popup config/completion → `InputPopupContext`
- Palette state/seedDone → `PaletteContext`
- Snippet editor focus/completion → `SnippetEditorContext`
- Calendar year/month/day/notes → `CalendarContext`
- Contribution chart dayCounts/selectedDate/notes → `ContribContext`

### 7. Preview Interface

`IPreviewContext` (`context/preview_common.go`) is the shared interface for all preview-mode contexts. It allows `PreviewNavHelper` and rendering code to work generically across `PreviewContext` (card list, pick results, compose) and `DatePreviewContext` (date-based view).

```go
type IPreviewContext interface {
    types.Context
    NavState() *PreviewNavState
    DisplayState() *PreviewDisplayState
    SelectedCardIndex() int
    SetSelectedCardIndex(int)
    CardCount() int
    CaptureSnapshot() Snapshot
    RestoreSnapshot(Snapshot)
}
```

`ActivePreviewKey` in `ContextTree` tracks which preview context currently owns the `preview` view. Rendering dispatches based on this key.

### 8. Interface Boundaries

`types.IGuiCommon` is the authoritative interface for GUI operations. Both controllers and helpers use it:

- **`types.IGuiCommon`** — rendering, refresh, dialogs, context navigation, view access, search/preview/completion methods
- **`helpers.IGuiCommon`** — embeds `types.IGuiCommon` + adds `Contexts() *context.ContextTree` (can't live in `types/` due to import cycle)
- **`controllers.IGuiCommon`** — type alias for `types.IGuiCommon`

```
Controllers ──→ IHelpers ──→ Helpers ──→ helpers.IGuiCommon ──→ *Gui
     │                                                           │
     └──→ types.IGuiCommon ──────────────────────────────────────┘
```

### 9. Commands Layer

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

### 10. Navigator + NavigationManager + Snapshotter

**Files:** `helpers/navigator.go`, `context/nav_manager.go`, `types/snapshot.go`

The `Navigator` helper is the single entry point for all preview pane transitions.

**NavigationManager** (`context/nav_manager.go`) is the history stack. It holds `NavigationEvent` entries (ContextKey, Title, Snapshot, Timestamp) with a 50-entry cap.

**Snapshotter interface** (`types/snapshot.go`): every preview context implements `CaptureSnapshot()` / `RestoreSnapshot()`. Each context is responsible for serialising and restoring its own state.

**Navigator API**:
- `NavigateTo(destination, title, load)` — committed navigation; records a history entry; pushes context for preview destinations.
- `ShowHover(destination, title, load)` — hover (no history entry); title rendered in italics via ANSI escape codes.
- `ReplaceCurrent(destination, title, load)` — like `NavigateTo` but replaces rather than pushes the context stack (used by search, execute-pick).
- `Back()` / `Forward()` — rewind/advance history; re-run the Requery closure via the Snapshotter so renames/edits/deletes are reflected automatically.

**Capture-on-departure**: Navigator carries a `currentIsCommitted` flag. Every method snapshots the current preview context (if committed) into the current history entry before doing anything else. Toggle, scroll, and filter handlers never need to know about Navigator.

**Re-query on restore**: each context's snapshot carries a Requery closure (via `CardListSource.Requery`, `PickResultsSource.Requery`, `ComposeContext.Requery`, `DatePreviewContext.Requery`). Restore re-runs the query so the data is fresh; frozen data is a fallback only.

**Two independent stacks**: the context stack (`Esc` → `popContext`) and the nav history stack (`[` / `]` → `Navigator.Back` / `Forward`) are separate by design. `Esc` in the preview pane returns focus to the last side pane; it never touches history.

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
RenderPreview() → replaceContext(PreviewContext)
```

### Date Preview Flow

```
Calendar/Contrib Enter → LoadDatePreview(date)
         │
ruinCmd.Pick.Pick(@date, all=true) → tagPicks (filterOutTodoLines, sortDonePicksLast)
ruinCmd.Pick.Pick(@date, todo=true) → todoPicks
ruinCmd.Search.Search(created:date) + Search(updated:date) → DeduplicateNotes → notes
         │
DatePreviewContext ← { TargetDate, TagPicks, TodoPicks, Notes }
         │
renderDatePreview() → three sections with headers, card line ranges, section ranges
         │
pushContext("datePreview")
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
- `scripts/smoke-test.sh` — tmux-driven TUI regression (60 assertions, 24 sections)
- `scripts/keybinding-test.sh` — keyboard shortcut smoke test (90 assertions, 69 sections covering all contexts)
- `./lazyruin --debug-bindings` — dump all registered controller bindings for regression diffing
- `testutil.MockExecutor` — fluent mock for CLI command testing without a real `ruin` binary
