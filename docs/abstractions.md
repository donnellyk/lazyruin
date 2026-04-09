# GUI Abstractions

Reusable abstraction patterns in `pkg/gui/`. The pattern throughout is **composition via closures and struct fields** rather than inheritance, consistent with the lazygit architecture.

## 1. Configurable Completion Editor

**Files:** `editor_completion.go`

`completionEditor` is a configurable `gocui.Editor` that replaces four near-identical editors (search, pick, input popup, snippet). Configured via lambdas for `state`/`triggers` and a `DrillFlags` bitmask controlling drill-down behavior (parent stacking with `/`, wiki-link headers with `#`).

The capture editor (`editor_capture.go`) remains bespoke due to markdown continuation on Enter and abbreviation-specific Tab handling.

## 2. Completion System

**Files:** `completion.go`, `completion_triggers.go`, `completion_candidates.go`, `completion_render.go`

Shared completion lifecycle across 5 popups (search, capture, pick, input popup, snippet editor). Each popup defines its own trigger set (`searchTriggers`, `captureTriggers`, `pickTriggers`, etc.) mapping prefixes like `#`, `[[`, `>`, `!` to candidate provider functions.

Key functions: `updateCompletion` (per-keystroke), `acceptCompletion` (apply selection), `Dismiss` (reset state), `completionUp`/`completionDown` (navigation).

## 3. Completion Handler Wrappers

**Files:** `completion.go`

Higher-order functions (`completionEsc`, `completionTab`, `completionEnter`) that check completion state before delegating to the underlying handler. Used by search, pick, and input popup bindings. Capture and snippet keep bespoke handlers due to extra logic (abbreviation acceptance, two-field Tab toggling).

## 4. Dialog System

**Files:** `dialogs.go`, `types/dialog.go`

Generic modal dialog framework supporting three types:

- **Confirm:** Yes/no prompts (e.g., "Delete note?")
- **Input:** Text field with callback (e.g., "Rename tag to:")
- **Menu:** Navigable list with optional shortcut keys (e.g., help, merge direction, view options)

`MenuItem` is defined in `types/dialog.go` (shared by helpers). `DialogState` is in `dialogs.go`.

API: `showConfirm(title, message, onConfirm)`, `showInput(title, message, onConfirm)`, `ShowMenuDialog(title, items)`, `closeDialog()`.

## 5. Input Popup Configuration

**Files:** `types/popup.go`

Generic "fill in a field" popup with configurable title, seed text, completion triggers, and `OnAccept` callback. Used for parent selection, tag rename, query save, inline tag/date toggling, and any future single-field dialogs. `InputPopupConfig` lives in `types/` so helpers can create popups without importing `gui`.

## 6. List Panel Navigation

**Files:** `context/list_context_trait.go`, `controllers/list_controller_trait.go`

`ListContextTrait` (embedded in each list context) owns the selection cursor and render/preview callbacks. `ListControllerTrait[T]` (embedded in each list controller) provides shared j/k/g/G/arrow navigation against the context.

Controller-side API:
- `NavBindings()` — returns the standard j/k/g/G/arrow `*types.Binding` slice
- `withItem(fn)` — guard that calls `fn(item)` only when a selection exists
- `singleItemSelected()` — produces a `DisabledReason` for binding guards
- `require(reasons...)` — combines multiple disabled reasons

Context-side state:
- `ListCursor` — holds `selectedLineIdx` and delegates clamp/find-by-ID to the context's `IList`
- `renderFn func()` — re-renders the list view (called on selection change)
- `updatePreviewFn func()` — refreshes preview for the new selection

## 7. Generic List Rendering

**Files:** `render.go` (`renderList`)

Renders any list panel with selection highlighting, scroll management, empty-state message, and per-item formatting via a builder callback. Drives all four list panels plus the command palette.

## 8. Palette Command Sources

**Files:** `commands.go`, `handlers_palette.go`

The command palette aggregates from two sources:

1. **Controller bindings** — any `types.Binding` with a non-empty `Description` automatically appears in the palette. The binding's `Category` and `Description` become the palette entry; `Key` is shown as the shortcut hint.

2. **`paletteOnlyCommands()`** (`commands.go`) — tab-switching and snippet management commands that have no keybinding; accessible only via the palette. Returns `[]PaletteCommand` which is merged with the controller-derived entries in `handlers_palette.go`.

## 9. Context System

**Files:** `gui.go`, `state.go` (`ContextKey`)

`ContextKey` enum (Notes, Queries, Tags, Preview, DatePreview, Search, SearchFilter, Capture, Pick, Palette, InputPopup, SnippetEditor, Calendar, Contrib) controls view focus, active keybindings, palette filtering, and status bar hints. `setContext()` handles all transitions. `ContextStack` tracks the navigation path for back-navigation.

## 10. Hint Definitions

**Files:** `hints.go`

`contextHintDef` struct is the single source of truth for both status bar hints and the help menu. Each context defines a full hint list (for help) and an optional shortened list (for the status bar).

## 11. Suggestion Dropdown Rendering

**Files:** `completion_render.go` (`renderSuggestionView`)

Generic suggestion dropdown renderer with scrolling and column-aligned label + detail. Used by all five completion-capable views (search, capture, pick, input popup, snippet).

## 12. IGuiCommon Interface Bridge

**Files:** `gui_common.go`, `helpers/helper_common.go`, `controllers/controller_common.go`

Two separate `IGuiCommon` interfaces prevent circular imports between `gui`, `helpers`, and `controllers`. `*Gui` satisfies both via adapter methods in `gui_common.go`. This allows helpers and controllers to call GUI operations (render, refresh, show dialogs, push context) without importing the `gui` package.

Controllers access helpers through the `IHelpers` interface, which provides typed accessors for each helper (e.g., `Helpers().Preview()`, `Helpers().NoteActions()`).

## 13. Preview Navigation History

**Files:** `helpers/preview_helper.go`, `context/preview_state.go`

The preview pane maintains a navigation history stack (`NavHistory []NavEntry`) in a `SharedNavHistory` struct shared across all preview contexts. Each entry captures the full preview state (cards, mode, cursor, scroll, title, and for date preview: target date, tag picks, todo picks, notes). `PushNavHistory()` snapshots before navigating, `NavBack()`/`NavForward()` restore entries. The history caps at 50 entries and supports `ShowNavHistory()` to jump to any entry via a menu dialog.

## 14. IPreviewContext Interface

**Files:** `context/preview_common.go`

`IPreviewContext` is the shared interface for all preview-mode contexts, enabling `PreviewNavHelper` and rendering code to work generically across `PreviewContext` (card list, pick results, compose) and `DatePreviewContext` (date-based view).

```go
type IPreviewContext interface {
    types.Context
    NavState() *PreviewNavState
    DisplayState() *PreviewDisplayState
    SelectedCardIndex() int
    SetSelectedCardIndex(int)
    CardCount() int
    NavHistory() *SharedNavHistory
}
```

Both preview contexts share `PreviewNavState` (scroll, cursor, card line ranges, header lines, links) and `PreviewDisplayState` (frontmatter, title, global tags, markdown toggles). `ActivePreviewKey` in `ContextTree` tracks which context currently owns the `preview` view.

## 15. Preview Navigation Trait

**Files:** `controllers/datepreview_controller.go`, `helpers/preview_nav_helper.go`

`PreviewNavTrait` provides shared preview keybindings (card jump J/K, line scroll j/k, header jump {/}, nav history [/], link highlight l/L, line operations) that work across preview contexts via `IPreviewContext`. Concrete controllers embed the trait and add context-specific bindings — e.g., `DatePreviewController` adds section jump `)` / `(`.

`PreviewNavHelper` provides the underlying navigation logic: `MoveDown`/`MoveUp`, `NextCard`/`PrevCard`, `NextHeader`/`PrevHeader`, `NextSection`/`PrevSection`, `PreviewEnter`, and all line operation handlers (todo toggle, done, inline tag, date, delete card, etc.).

## 16. Section-Based Card Layout (Date Preview)

**Files:** `context/datepreview_context.go`, `render_preview.go`

`DatePreviewContext` organizes cards into three sections: Inline Tags (pick results without todos, done items last), Todos (checkbox pick results), and Notes (created + updated, deduplicated). The context tracks `SectionRanges` (card index ranges per section) and `SectionLineRanges` (line number ranges per section), populated during rendering.

Helper methods `SectionForCard(idx)`, `SectionForLine(line)`, and `LocalCardIdx(globalIdx)` enable section-aware navigation. `filterOutTodoLines()` removes checkbox lines from tag picks, and `sortDonePicksLast()` splits mixed active/done pick results into separate groups with done groups appended after all active groups.
