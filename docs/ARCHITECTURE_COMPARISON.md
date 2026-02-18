# Architecture Comparison: lazyruin vs lazygit

This document catalogs where lazyruin's architecture has diverged from lazygit's patterns, and where adopting lazygit patterns could improve maintainability.

---

## 1. God Struct vs Controller Layer

**Lazygit:** Controllers are standalone structs implementing `IController`. Each controller owns a slice of keybindings and focus/blur callbacks. A `FilesController`, `BranchesController`, etc. are separate types composed with `baseController` (null object) and `ListControllerTrait`. Controllers are attached to contexts, not to the Gui.

**Lazyruin:** Every handler is a method on `*Gui`. The `handlers_preview.go` file alone is 1,500+ lines. There is no controller interface — keybindings point directly to `gui.someMethod`.

**Impact:** The Gui struct is the coupling point for everything. Any handler can reach into any state, any view, any command. As complexity grows, this makes it hard to reason about what a given handler touches.

**Refactor opportunity:** Extract domain controllers (`PreviewController`, `NotesController`, `TagsController`) that receive a narrow interface (`ControllerCommon`) instead of the full Gui. The command table entries would reference `controller.method` rather than `gui.method`. This is a large refactor but would make the codebase significantly more testable and navigable.

**Priority: Medium-high.** `handlers_preview.go` is already the largest file and accumulates every new preview feature.

---

## 2. Context Stack vs Context Enum

**Lazygit:** Contexts are full objects implementing the `Context` interface. They form a stack managed by `ContextMgr` with `Push()`, `Pop()`, and `Replace()` operations. Each context carries its own keybindings, focus/blur handlers, and rendering logic. Popups are contexts that push onto the stack. Escape pops back to the previous context automatically.

**Lazyruin:** Context is a `ContextKey` string enum. `setContext()` stores `PreviousContext` as a single value (not a stack). Modal state is tracked via boolean flags (`SearchMode`, `CaptureMode`, `PaletteMode`, `CalendarMode`, etc.) checked by `overlayActive()`.

**Impact:**
- Only one level of "back" is possible with `PreviousContext`. The nav history we just added works around this for preview content, but the context system itself can't go back more than one step.
- Every new modal requires a new boolean flag, a new check in `overlayActive()`, and new conditional blocks in `layout()`.
- Keybindings can't be scoped to the active context automatically — the `suppressDuringDialog()` wrapper is a manual workaround.

**Refactor opportunity:** Replace the boolean flags with a context stack. Each popup/modal would push a context. Escape always pops. Keybindings would be registered per-context so suppression is automatic. The `overlayActive()` function disappears — you just check if the top of the stack is a popup context.

**Priority: Medium.** The current approach works but scales poorly. Each new modal type requires changes in 3-4 places.

---

## 3. Helper Layer

**Lazygit:** 40+ helper structs (`RefsHelper`, `StagingHelper`, `RefreshHelper`, etc.) encapsulate reusable business logic. Controllers call helpers; helpers call commands. Helpers are injected via `HelperCommon`.

**Lazyruin:** No helper layer. Business logic lives directly in handler methods on Gui. For example, `reloadContent()`, `reloadPreviewCards()`, `buildSearchOptions()`, `refreshAll()` are all Gui methods mixing concerns of refresh orchestration, CLI calls, and state mutation.

**Impact:** Handler methods that should be simple (tag filter → show results) end up interleaved with refresh logic, state preservation, and rendering calls. Code reuse between handlers happens by calling other Gui methods, creating implicit coupling.

**Refactor opportunity:** Extract helpers for cross-cutting concerns:
- `RefreshHelper` — coordinates data reload and selection preservation
- `PreviewHelper` — manages preview state, nav history, card manipulation
- `CompletionHelper` — already partially isolated in `completion*.go` files

This is lower cost than the controller refactor and could be done incrementally.

**Priority: Medium.**

---

## 4. Rendering Ownership

**Lazygit:** Each context owns its rendering via `HandleRender()` and `HandleRenderToMain()`. The context decides what to render and when. Refresh marks contexts dirty; rendering is pull-based.

**Lazyruin:** Rendering is push-based and scattered. After any state change, handlers explicitly call `gui.renderNotes()`, `gui.renderPreview()`, `gui.updateStatusBar()`, etc. Missing a render call means stale UI. `setContext()` re-renders all three list panels every time focus changes, regardless of whether they changed.

**Impact:** Easy to forget a render call after a state change. The full-rerender in `setContext()` is wasteful. As the UI grows, this pattern means more rendering paths to maintain.

**Refactor opportunity:** A dirty-flag system where state mutations mark views as needing re-render, and a single `renderDirty()` pass handles it. Lower priority since gocui rendering is fast enough that the current approach doesn't cause visible lag.

**Priority: Low.** Correctness issue more than performance.

---

## 5. Keybinding Scoping

**Lazygit:** Keybindings are registered per-context via each controller's `GetKeybindings()`. When a context is active, only its bindings apply. Global bindings exist on the global controller.

**Lazyruin:** Two systems coexist:
1. The `Command` table in `commands.go` — has `Views` and `Contexts` fields but these are only used for palette filtering, not for actual keybinding scoping.
2. Per-view nav bindings in `keybindings.go` — registered on specific gocui views, which does provide view-level scoping.

The `Command.Contexts` field is misleading — it filters what appears in the command palette, but the underlying gocui keybinding fires regardless of context. The `suppressDuringDialog()` wrapper is the actual mechanism preventing actions during overlays.

**Impact:** A key bound to `PreviewView` fires whether the user is in the preview normally or has a dialog open over it. The suppression wrapper handles this but it's fragile — every handler that shouldn't fire during dialogs needs to be wrapped.

**Refactor opportunity:** If context stack is adopted (point 2), keybindings would register/unregister with context push/pop, making suppression automatic.

**Priority: Low** (coupled to the context stack refactor).

---

## 6. Dependency Injection

**Lazygit:** The `Common` struct provides shared dependencies (logging, translations, config, filesystem). `ControllerCommon` and `HelperCommon` extend it with layer-appropriate access. All layers receive their dependencies through these structs.

**Lazyruin:** No dependency injection. Everything hangs off `*Gui`. Handlers access `gui.ruinCmd`, `gui.state`, `gui.views`, `gui.config` directly. Test setup requires building a full Gui instance.

**Impact:** Unit testing handlers in isolation is impossible — you need the entire Gui wired up. The `commands` package is well-isolated (uses the `Executor` interface for testing), but the GUI layer has no seams.

**Refactor opportunity:** Define a narrow interface that handlers/controllers need (access to commands, state, view operations) and inject it. This would enable testing handler logic without a full gocui instance.

**Priority: Low.** The smoke test and integration test approach works well enough for now.

---

## 7. State Organization

**Lazygit:** Separates concerns into `Model` (data), `Modes` (operation context like rebase/cherry-pick), `SearchState`, and context-local state (selection, range). Each lives in its own struct.

**Lazyruin:** `GuiState` is a flat bag of everything:
- Panel data (`NotesState`, `TagsState`, etc.)
- Preview state (`PreviewState` with 15 fields)
- Modal flags (8 boolean fields)
- Completion states (5 `*CompletionState` fields)
- Nav history
- Context tracking
- UI metadata (`lastWidth`, `lastHeight`)

**Impact:** The struct is large but manageable at current size. The modal flags are the worst part — they grow linearly with each new modal type.

**Refactor opportunity:** Group related fields:
- `ModalState` — replaces the individual boolean flags (or eliminated by context stack)
- Move completion states into their respective modal states
- Preview state is already well-contained in `PreviewState`

**Priority: Low.** Would be solved naturally by the context stack refactor.

---

## 8. Refresh Coordination

**Lazygit:** `RefreshHelper` coordinates all data loading. Supports `SYNC`, `ASYNC`, and `BLOCK_UI` refresh strategies. Refresh is atomic — load data, update model under mutex, mark dirty.

**Lazyruin:** Refresh is ad-hoc. Each handler calls the specific refresh functions it needs (`refreshNotes(true)`, `refreshTags(false)`, `renderPreview()`). The `preserve` boolean controls whether selection is reset. `setContext()` does its own refresh on every focus change.

**Impact:** Inconsistency in which handlers refresh what. Some handlers call `reloadContent()` (which does notes + preview), some call `refreshNotes()` + `renderPreview()` separately, some just call `renderPreview()`. The `setContext()` re-fetch on every focus change means switching from Tags to Notes triggers a CLI call even if nothing changed.

**Refactor opportunity:** Centralize refresh into a coordinator that:
- Tracks what's dirty
- Batches CLI calls
- Preserves selections uniformly
- Doesn't re-fetch on focus change unless data is stale

**Priority: Medium.** The redundant CLI calls on every focus change are the most wasteful part.

---

## Recommended Order of Attack

If refactoring, I'd suggest this order based on impact-to-effort ratio:

1. **Context stack** (replaces boolean flags, enables automatic keybinding scoping, eliminates `overlayActive()`) — medium effort, high impact
2. **Extract PreviewController** (just the preview domain, as a pilot for the controller pattern) — medium effort, high impact on the largest file
3. **Refresh coordinator** (stop re-fetching on every focus change) — low effort, medium impact
4. **Helper extraction** (start with `RefreshHelper` and `PreviewHelper`) — low effort, medium impact

Items 3 and 4 can be done incrementally without architectural changes. Items 1 and 2 are larger but address the root causes.
