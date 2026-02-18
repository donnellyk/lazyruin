# Refactor Rework Plan

This document captures the focused refactor outline for pushing more logic out of
`pkg/gui/gui.go` and `pkg/gui/handlers.go` into controllers, contexts, and helpers.

## Focused Refactor: Pull Logic Out of `gui.go` and `handlers.go`

This is a scoped, low-risk approach that moves logic into controllers,
contexts, and helpers without changing behavior.

### Step 1: Wire helpers and common dependencies

**Goal**: Controllers and legacy command handlers call helpers instead of
`Gui` methods.

1. Instantiate `HelperCommon` and `Helpers` in `NewGui`.
2. Instantiate `ControllerCommon` with `IGuiCommon`, `RuinCommand`, and
   helpers, and pass it into controller constructors.
3. Convert callbacks in `setup*Context` to call helper methods where possible.

**Targets**:
- `pkg/gui/gui.go` (new helper/controller wiring)
- `pkg/gui/helpers/*`
- `pkg/gui/controllers/*_controller.go`

### Step 2: Extract global actions into `GlobalController` + helpers

**Goal**: Reduce global behaviors in `handlers.go`.

Move these handlers into helpers or controllers:
- `openSearch`, `executeSearch`, `cancelSearch` -> `SearchHelper`.
- `refresh` -> `RefreshHelper` (or new `AppHelper`).
- `nextPanel`, `prevPanel`, `focus*` -> `GlobalController` with context stack
  logic centralized there.

Maintain existing side effects via helpers and context access.

**Targets**:
- `pkg/gui/handlers.go`
- `pkg/gui/controllers/global_controller.go`
- `pkg/gui/helpers/search_helper.go`
- `pkg/gui/helpers/refresh_helper.go`

### Step 3: Move list click/scroll behaviors into controllers

**Goal**: Limit view-specific handlers in `handlers_*.go`.

1. Create a small view utilities helper for click indexing and scrolling.
2. Update controller mouse handlers to call these helpers and set selection
   via context trait APIs.

**Targets**:
- `pkg/gui/controllers/notes_controller.go`
- `pkg/gui/controllers/tags_controller.go`
- `pkg/gui/controllers/queries_controller.go`
- `pkg/gui/render.go` (or a new helper file for shared list view logic)

### Step 4: Reduce sync functions by wrapping or replacing `GuiState`

**Goal**: Stop the manual sync between context and legacy state.

Two safe options:
- Wrap existing `GuiState` structs in contexts (oracle recommendation), so
  contexts become a thin view over `GuiState` and `sync*ToLegacy` can be
  removed.
- Or, fully move state into contexts and leave `GuiState` only for cross-cutting
  concerns.

**Targets**:
- `pkg/gui/context/*_context.go`
- `pkg/gui/state.go`
- `pkg/gui/gui.go`

### Step 5: Introduce missing contexts/controllers for remaining popups

**Goal**: Pull calendar/palette/snippet/contrib bindings into contexts and
controllers, and remove their legacy binding registration.

Start with the smallest (palette, snippet editor), then calendar/contrib.

**Targets**:
- `pkg/gui/context/context_tree.go`
- `pkg/gui/keybindings.go`
- `pkg/gui/handlers_palette.go`
- `pkg/gui/handlers_snippets.go`
- `pkg/gui/calendar.go`
- `pkg/gui/contrib.go`

## Additional Refactor Items

These items come from the broader findings and should be planned alongside
the focused `gui.go`/`handlers.go` cleanup.

- **Context stack + lifecycle hooks**: Integrate `types.Context` into the
  active stack so `HandleFocus`, `HandleFocusLost`, and `HandleRender` are
  invoked (and `ContextKind` governs popup suppression).
- **Palette source of truth**: Decide whether palette entries remain derived
  from controller bindings or return to a canonical command table in
  `commands.go` for stability.
- **Mouse binding fidelity**: Adjust `registerContextBindings` to pass through
  `ViewMouseBindingOpts` and honor per-binding view names, enabling click
  coordinate logic in controllers.
- **Remove legacy listPanel**: After list navigation is fully trait-based,
  delete `listPanel`/`listMove` and related tests.

## Focused Checklist for `gui.go` and `handlers.go`

1. `gui.go`:
   - Create and store helpers (`HelperCommon` + `Helpers`).
   - Create `ControllerCommon` and pass to all controller constructors.
   - Replace callback bodies with helper calls where available.
2. `handlers.go`:
   - Move global action logic into helpers/controllers.
   - Delete `listPanel` helpers once no references remain.
   - Keep only truly GUI-specific glue (layout, view management, rendering).

## Suggested Order

1. Helper wiring + GlobalController cleanup.
2. Notes/Tags/Queries action handlers into helpers.
3. Mouse handling cleanup and view utilities.
4. Context stack improvements (wrap or replace `GuiState`).
5. Remaining popups into contexts/controllers.
