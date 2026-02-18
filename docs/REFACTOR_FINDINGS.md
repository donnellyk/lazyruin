# Refactor Review Findings

This document captures discrepancies against the refactor plan and oracle review.

It is based on the current code state after the controller/context migration.

## Discrepancies vs Plan/Review

### 1) Context stack still uses legacy `ContextKey` and `GuiState`

**Expectation**: The refactor plan describes a context stack of rich
`types.Context` objects (with `ContextKind` driving popup suppression), and
controllers/contexts owning focus/render lifecycle hooks.

**Current**:
- The stack is still `[]ContextKey` in `GuiState`.
- `popupActive()` uses a legacy main-panel allowlist.
- `activateContext` uses `contextToView` and calls render/refresh manually.

**Impact**: The context lifecycle hooks (`HandleFocus`, `HandleFocusLost`,
`HandleRender`) are never invoked, and `ContextKind` does not govern
popup suppression or focus behavior.

**Relevant files**:
- `pkg/gui/state.go`
- `pkg/gui/gui.go`
- `pkg/gui/context/base_context.go`

### 2) Duplicate state between contexts and `GuiState`

**Expectation**: Either wrap existing `GuiState` structs with contexts (oracle
recommendation), or move all panel state into contexts as a single source of
truth.

**Current**:
- Contexts store their own `Items` and tab/cursor state.
- `GuiState` still stores `NotesState`, `TagsState`, `QueriesState`, etc.
- `sync*ToLegacy()` copies context state back into `GuiState`.

**Impact**: Dual sources of truth and manual sync are fragile. In particular,
`refresh*` functions update context data and then copy into `GuiState`, which
can drift if other code paths mutate `GuiState` directly.

**Relevant files**:
- `pkg/gui/context/notes_context.go`
- `pkg/gui/context/tags_context.go`
- `pkg/gui/context/queries_context.go`
- `pkg/gui/gui.go`
- `pkg/gui/state.go`

### 3) Helpers are defined but not wired

**Expectation**: Controllers and legacy command handlers should call into
helpers to keep `Gui` thin, with domain logic centralized.

**Current**:
- Helpers exist in `pkg/gui/helpers/`.
- Controllers are wired with callbacks that call back into `Gui` handlers.
- `ControllerCommon`/`HelperCommon` are defined, but `Gui` does not create or
  inject them.

**Impact**: The migration keeps large handler files as the source of truth and
limits testable units around helper logic.

**Relevant files**:
- `pkg/gui/helpers/`
- `pkg/gui/controllers/controller_common.go`
- `pkg/gui/gui.go`

### 4) Palette source diverges from oracle guidance

**Expectation (oracle)**: Keep `commands.go` as the canonical public action list
for palette/docs/help, while controllers handle navigation and some shortcuts.

**Current**:
- Palette commands are derived from controller bindings + palette-only entries.
- The legacy command table is no longer used for palette aggregation.

**Impact**: Changes to controller bindings immediately affect palette output,
which is the opposite of the oracle recommendation to keep palette stable
and controller-specific.

**Relevant files**:
- `pkg/gui/handlers_palette.go`
- `pkg/gui/commands.go`

### 5) Migration is incomplete for non-core contexts

**Expectation**: Each popup/panel gets a context/controller pair and is
registered via the context tree.

**Current**:
- Context tree includes only a subset (global, notes, tags, queries, preview,
  search, capture, pick, input popup).
- Palette, calendar, snippet editor, and contrib remain in legacy binding
  registration.

**Impact**: The binding system is split between controller registration and
legacy registration, increasing complexity and the chance of suppression bugs.

**Relevant files**:
- `pkg/gui/context/context_tree.go`
- `pkg/gui/keybindings.go`

### 6) Mouse bindings do not use full gocui mouse opts

**Expectation**: Mouse handling would move into controllers with access to
click coordinates when needed.

**Current**:
- Controller mouse bindings register `ViewMouseBinding` handlers but
  `registerContextBindings` ignores `ViewMouseBindingOpts` and the per-binding
  `ViewName`.
- The handlers call back into legacy click helpers without the original view
  context.

**Impact**: Controller mouse bindings are limited and do not support the
coordinate-driven click logic used in legacy handlers.

**Relevant files**:
- `pkg/gui/keybindings.go`
- `pkg/gui/controllers/*_controller.go`

### 7) Lifecycle hooks are defined but unused

**Expectation**: Contexts should receive focus/render hooks through the stack.

**Current**: `BaseContext` aggregates hooks, but no code calls
`ctx.HandleFocus`, `ctx.HandleFocusLost`, or `ctx.HandleRender`.

**Impact**: Controllers cannot participate in focus lifecycle or render-to-main
flows, limiting the benefit of the context system.

**Relevant files**:
- `pkg/gui/context/base_context.go`
- `pkg/gui/gui.go`

### 8) Legacy listPanel still exists

**Expectation**: List navigation should be fully ported into traits and old
helpers removed.

**Current**: `listPanel` and `listMove` remain in `handlers.go` alongside the
new traits.

**Impact**: Residual code increases maintenance surface and suggests the
migration is not fully cleaned up.

**Relevant files**:
- `pkg/gui/handlers.go`
- `pkg/gui/context/list_context_trait.go`
- `pkg/gui/controllers/list_controller_trait.go`

## Focused Refactor Plan

The refactor plan has been moved to `docs/REFACTOR_REWORK_PLAN.md`.
