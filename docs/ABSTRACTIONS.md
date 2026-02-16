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

**Files:** `dialogs.go`, `state.go` (`DialogState`)

Generic modal dialog framework supporting three types:

- **Confirm:** Yes/no prompts (e.g., "Delete note?")
- **Input:** Text field with callback (e.g., "Rename tag to:")
- **Menu:** Navigable list with optional shortcut keys (e.g., help, merge direction)

API: `showConfirm(title, message, onConfirm)`, `showInput(title, message, onConfirm)`, `closeDialog()`.

## 5. Input Popup Configuration

**Files:** `state.go` (`InputPopupConfig`), `handlers_preview.go`

Generic "fill in a field" popup with configurable title, seed text, completion triggers, and `OnAccept` callback. Used for parent selection, tag rename, query save, and any future single-field dialogs.

## 6. List Panel Navigation

**Files:** `handlers.go` (`listPanel`)

Shared j/k/g/G/arrow/mouse-wheel navigation for Notes, Queries, Tags, and Parents panels. Configured via closures:

- `selectedIndex *int` — pointer to the panel's selection state
- `itemCount func() int` — current item count
- `render func()` — re-render the list
- `updatePreview func()` — refresh preview for new selection

## 7. Generic List Rendering

**Files:** `render.go` (`renderList`)

Renders any list panel with selection highlighting, scroll management, empty-state message, and per-item formatting via a builder callback. Drives all four list panels plus the command palette.

## 8. Command Table

**Files:** `commands.go`

Single `Command` struct is the source of truth for both keybinding registration and command palette generation:

- `Keys` / `Views` — gocui keybindings (nil = palette-only)
- `Handler` / `OnRun` — action callbacks
- `Contexts` — palette context filtering (nil = always visible)
- `NoPalette` — suppress from palette
- `KeyHint` — display hint (auto-derived if empty)

## 9. Context System

**Files:** `gui.go`, `state.go` (`ContextKey`)

`ContextKey` enum (Notes, Queries, Tags, Preview, Search, SearchFilter, Capture, Pick, Palette) controls view focus, active keybindings, palette filtering, and status bar hints. `setContext()` handles all transitions and tracks `PreviousContext` for back-navigation.

## 10. Hint Definitions

**Files:** `hints.go`

`contextHintDef` struct is the single source of truth for both status bar hints and the help menu. Each context defines a full hint list (for help) and an optional shortened list (for the status bar).

## 11. Suggestion Dropdown Rendering

**Files:** `completion_render.go` (`renderSuggestionView`)

Generic suggestion dropdown renderer with scrolling and column-aligned label + detail. Used by all five completion-capable views (search, capture, pick, input popup, snippet).
