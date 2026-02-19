# Final Cleanup Plan

This plan eliminates all remaining legacy code, temporary scaffolding, type aliases, redundant wrappers, and handler files that survive after the REGRESSION_REFACTOR_PLAN and DRY_REFACTOR_PLAN complete. After this plan, the refactor is complete.

**Prerequisite:** REGRESSION_REFACTOR_PLAN Steps 1–3 and DRY_REFACTOR_PLAN Steps 1–10 are done.

---

## Overview

| # | Step | What it removes | LOC removed (est.) | Risk |
|---|------|----|-----|------|
| 1 | Delete `handlers_note_actions.go` | Duplicate handler methods that have helper equivalents | ~180 | Low |
| 2 | Migrate capture handlers to `CaptureHelper` | `handlers_capture.go` + open/close handlers on `*Gui` | ~120 | Medium |
| 3 | Migrate pick handlers to `PickHelper` | `handlers_pick.go` + open/close handlers on `*Gui` | ~100 | Medium |
| 4 | Migrate input popup handlers to `InputPopupHelper` | `handlers_input_popup.go` + open/close handlers on `*Gui` | ~90 | Medium |
| 5 | Migrate snippet handlers to `SnippetHelper` | `handlers_snippets.go` + create/list/delete on `*Gui` | ~280 | Medium |
| 6 | Migrate palette handlers to `PaletteHelper` | `handlers_palette.go` handlers on `*Gui` | ~350 | Medium |
| 7 | Migrate calendar handlers to `CalendarHelper` | `calendar.go` handlers on `*Gui` | ~480 | Medium |
| 8 | Migrate contrib handlers to `ContribHelper` | `contrib.go` handlers on `*Gui` | ~400 | Medium |
| 9 | Delete type aliases from `state.go` and `completion.go` | 13 type aliases, 2 const re-exports, 1 var alias | ~30 | Low |
| 10 | Delete `ContextKey` constants and `mainPanelContexts` from `state.go` | 13 constants + map; replace with `GetKind()` | ~30 | Low |
| 11 | Replace `activateContext` with `HandleFocus` hooks | Move per-context refresh/preview logic to focus hooks | ~40 | Medium |
| 12 | Collapse private render/refresh wrappers | Delete double-wrapped private methods, use public directly | ~60 | Low |
| 13 | Slim down `gui_common.go` passthrough methods | Inline trivial one-liners where the public method IS the implementation | ~30 | Low |
| 14 | Clean up `keybindings.go` dead comments and `globalNavBindings` | Remove 15 "removed" comments, inline remaining bindings | ~30 | Low |
| 15 | Update `ARCHITECTURE.md` and delete completed plan docs | Docs reflect final state | — | None |

---

## Step 1: Delete `handlers_note_actions.go`

### Current state

`handlers_note_actions.go` contains 5 methods on `*Gui`:

| Handler method | Helper equivalent | Live callers outside own file |
|---|---|---|
| `gui.addGlobalTag` | `NoteActionsHelper.AddGlobalTag` | **None** — `NotesController` and `PreviewController` both call the helper |
| `gui.removeTag` | `NoteActionsHelper.RemoveTag` | **None** |
| `gui.setParentDialog` | `NoteActionsHelper.SetParentDialog` | **None** |
| `gui.removeParent` | `NoteActionsHelper.RemoveParent` | **None** |
| `gui.toggleBookmark` | `NoteActionsHelper.ToggleBookmark` | **None** |

Every caller already goes through the helper. These handler methods are dead code.

### Action

Delete `pkg/gui/handlers_note_actions.go` entirely.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh`

---

## Step 2: Migrate Capture Handlers to `CaptureHelper`

### Current state

`handlers_capture.go` contains 5 methods on `*Gui`:
- `openCapture` — guards popup, resets state, pushes context
- `submitCapture` — reads view content, calls `ruin log`, closes, refreshes
- `cancelCapture` — dismisses completion or closes
- `captureTab` — accepts completion (abbreviation, parent, or generic)
- `closeCapture` — resets state, disables cursor, pops context

Callers:
- `gui.go:setupGlobalContext` → `OnNewNote: func() error { return gui.openCapture(gui.g, nil) }`
- `gui.go:setupCaptureContext` → bindings reference `gui.submitCapture`, `gui.cancelCapture`, `gui.captureTab`
- `handlers_test.go` → `tg.gui.openCapture(tg.g, nil)` (2 sites)
- `palette_test.go` → `tg.gui.openCapture(tg.g, nil)` (1 site)

### Action

1. Create `pkg/gui/helpers/capture_helper.go`:

```go
type CaptureHelper struct {
    c *HelperCommon
}

func NewCaptureHelper(c *HelperCommon) *CaptureHelper {
    return &CaptureHelper{c: c}
}

func (self *CaptureHelper) OpenCapture() error { ... }
func (self *CaptureHelper) SubmitCapture(content string, parent *CaptureParentInfo, quickCapture bool) error { ... }
func (self *CaptureHelper) CancelCapture(completionActive bool, quickCapture bool) error { ... }
func (self *CaptureHelper) CloseCapture(quickCapture bool) error { ... }
```

The helper methods take explicit parameters instead of reading from `gui.views.Capture` or `gui.QuickCapture` — the view content is extracted at the call site (in the controller binding closure) and passed in.

2. Add `CaptureHelper` to `Helpers` aggregator and `IHelpers` interface.

3. Move `CaptureParentInfo` from `state.go` to `helpers/capture_helper.go` (it's only used by capture logic). If it's referenced from `state.go`'s `GuiState`, keep a reference there pointing to the helpers package type.

4. Update `gui.go:setupCaptureContext` bindings to call the helper:

```go
{Key: gocui.KeyCtrlS, Handler: func() error {
    v := gui.views.Capture
    content := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
    return gui.helpers.Capture().SubmitCapture(content, gui.state.CaptureParent, gui.QuickCapture)
}},
```

5. Update `gui.go:setupGlobalContext` `OnNewNote` to call `gui.helpers.Capture().OpenCapture()`.

6. Update tests to call `tg.gui.helpers.Capture().OpenCapture()`.

7. Delete `handlers_capture.go`.

### Completion wrapping

The capture tab handler calls `gui.acceptAbbreviationInCapture`, `gui.acceptParentCompletion`, `gui.acceptCompletion`, and `gui.renderCaptureTextArea`. These are completion engine functions in `completion.go` that operate on `*gocui.View` and `*CompletionState`. The controller binding closure should extract the view and state, then call the completion functions. The completion functions themselves stay in `completion.go` (they're shared across search/capture/pick/snippet).

### Note on `captureTab`

`captureTab` needs access to completion engine internals (`isAbbreviationCompletion`, `isParentCompletion`, `acceptCompletion`, `acceptAbbreviationInCapture`, `acceptParentCompletion`, `renderCaptureTextArea`). These are all functions on `*Gui` in `completion.go`. Since the helper can't call `*Gui` methods directly, the tab handler should remain as a closure in `setupCaptureContext` that calls the completion functions on `gui`:

```go
{Key: gocui.KeyTab, Handler: func() error {
    return gui.captureTabAction(gui.g, gui.views.Capture)
}},
```

Where `captureTabAction` is a renamed `captureTab` that stays in `completion.go` (or a new `completion_actions.go`). The key point is that `handlers_capture.go` as a file is deleted — the tab logic moves to the completion engine where it belongs.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (capture flow: open, type, submit, cancel)

---

## Step 3: Migrate Pick Handlers to `PickHelper`

### Current state

`handlers_pick.go` contains 4 methods on `*Gui`:
- `openPick` — guards popup, resets state, pushes context
- `togglePickAny` — toggles `PickAnyMode`, updates footer
- `executePick` — parses tags/filters from view, calls `ruin pick`, shows results
- `cancelPick` — resets state, pops context

Callers:
- `gui.go:setupGlobalContext` → `OnPick: func() error { return gui.openPick(gui.g, nil) }`
- `gui.go:setupPickContext` → bindings reference `gui.executePick`, `gui.cancelPick`, `gui.togglePickAny`
- `palette_test.go` → `tg.gui.openPick(tg.g, nil)` (1 site)

### Action

1. Create `pkg/gui/helpers/pick_helper.go`:

```go
type PickHelper struct {
    c *HelperCommon
}

func NewPickHelper(c *HelperCommon) *PickHelper {
    return &PickHelper{c: c}
}

func (self *PickHelper) OpenPick() error { ... }
func (self *PickHelper) ExecutePick(raw string, anyMode bool) error { ... }
func (self *PickHelper) CancelPick() { ... }
func (self *PickHelper) TogglePickAny() { ... }
```

2. Add to `Helpers` aggregator and `IHelpers`.

3. Update `gui.go` bindings to call helpers, extracting view content in closures.

4. Update tests.

5. Delete `handlers_pick.go`.

### Note on `executePick`

`executePick` reads raw content from `gui.views.Pick`, parses tags and `@date` filters, calls `gui.ruinCmd.Pick.Pick(...)`, then calls `gui.helpers.Preview().ShowPickResults(...)` and `gui.replaceContextByKey(PreviewContext)`. The helper version takes the raw string and anyMode as parameters:

```go
func (self *PickHelper) ExecutePick(raw string, anyMode bool) error {
    if raw == "" {
        self.CancelPick()
        return nil
    }
    // Parse tags and filters
    ...
    results, err := self.c.RuinCmd().Pick.Pick(tags, anyMode, filterStr)
    gui := self.c.GuiCommon()
    // Reset pick state
    gui.SetCursorEnabled(false)
    // Show results
    gui.SetPreviewPickResults(results, 0, 1, 0, " Pick: "+raw+" ")
    gui.ReplaceContextByKey("preview")
    return nil
}
```

The `PickQuery` and `PickCompletion` state fields on `GuiState` need to be accessible from the helper. Add `SetPickQuery(string)`, `ResetPickCompletion()` to `IGuiCommon`, or move pick state to a `PickContext` field (cleaner — pick state belongs with the pick context, not `GuiState`).

**Decision:** Move `PickQuery string`, `PickAnyMode bool`, `PickSeedHash bool` from `GuiState` to `PickContext`. This is consistent with how `NotesContext` owns `CurrentTab` and `PreviewContext` owns `Mode`. The helper reads/writes these through `gui.Contexts().Pick.*`.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (pick flow: open, type tags, execute, toggle any mode)

---

## Step 4: Migrate Input Popup Handlers to `InputPopupHelper`

### Current state

`handlers_input_popup.go` contains 5 methods on `*Gui`:
- `openInputPopup` — resets state, stores config, pushes context
- `closeInputPopup` — resets state, deletes views, pops context
- `inputPopupEnter` — accepts completion or calls config.OnAccept
- `inputPopupTab` — accepts completion (delegates to enter)
- `inputPopupEsc` — dismisses completion or closes popup

Callers:
- `gui_common.go:OpenInputPopup` → `gui.openInputPopup(config)` (the `IGuiCommon` adapter)
- `gui.go:setupInputPopupContext` → bindings reference `gui.inputPopupEnter`, `gui.inputPopupTab`, `gui.inputPopupEsc`
- Helpers call `gui.OpenInputPopup(...)` (through `IGuiCommon`) — 6 sites in `note_actions_helper.go` and `preview_helper.go`

### Action

1. Create `pkg/gui/helpers/input_popup_helper.go`:

```go
type InputPopupHelper struct {
    c *HelperCommon
}

func NewInputPopupHelper(c *HelperCommon) *InputPopupHelper {
    return &InputPopupHelper{c: c}
}

func (self *InputPopupHelper) OpenInputPopup(config *types.InputPopupConfig) { ... }
func (self *InputPopupHelper) CloseInputPopup() { ... }
func (self *InputPopupHelper) HandleEnter(raw string, selectedItem *types.CompletionItem) error { ... }
func (self *InputPopupHelper) HandleEsc(completionActive bool) error { ... }
```

2. Add to `Helpers` aggregator and `IHelpers`.

3. Replace `gui_common.go:OpenInputPopup` to delegate to helper:

```go
func (gui *Gui) OpenInputPopup(config *types.InputPopupConfig) {
    gui.helpers.InputPopup().OpenInputPopup(config)
}
```

4. Move `InputPopupConfig`, `InputPopupSeedDone`, `InputPopupCompletion` state ownership. `InputPopupConfig` and `InputPopupSeedDone` should move to `InputPopupContext` (state belongs with context). `InputPopupCompletion` is already on `GuiState` — move to `InputPopupContext` as well.

5. Update `gui.go:setupInputPopupContext` bindings to call helpers with extracted parameters:

```go
{Key: gocui.KeyEnter, Handler: func() error {
    v, _ := gui.g.View(InputPopupView)
    raw := strings.TrimSpace(v.TextArea.GetUnwrappedContent())
    state := gui.contexts.InputPopup.Completion
    var item *types.CompletionItem
    if state.Active && len(state.Items) > 0 {
        selected := state.Items[state.SelectedIndex]
        item = &selected
    }
    return gui.helpers.InputPopup().HandleEnter(raw, item)
}},
```

6. Delete `handlers_input_popup.go`.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (add tag, remove tag, set parent, bookmark — all use input popup)

---

## Step 5: Migrate Snippet Handlers to `SnippetHelper`

### Current state

`handlers_snippets.go` contains ~277 lines with these methods on `*Gui`:
- `listSnippets` — shows menu dialog listing all abbreviations
- `snippetExpansionTriggers` — builds completion triggers (merges search + capture triggers, excludes `!`)
- `acceptSnippetParentCompletion` — accepts parent completion keeping `>path` literal
- `createSnippet` — opens two-field snippet editor
- `snippetEditorTab` — toggles focus or accepts completion
- `snippetEditorEnter` — saves snippet or accepts completion
- `snippetEditorClickName` / `snippetEditorClickExpansion` — mouse focus handlers
- `snippetEditorEsc` — dismisses completion or closes editor
- `closeSnippetEditor` — tears down views, pops context
- `deleteSnippet` — shows menu then confirmation, deletes snippet

Callers:
- `commands.go:paletteOnlyCommands` → `gui.listSnippets`, `gui.createSnippet`, `gui.deleteSnippet`
- `gui.go:setupSnippetEditorContext` (not shown but inferred) → bindings reference snippet methods
- `layout.go:94` → `gui.createSnippetEditor(...)` (view creation)

### Action

1. Create `pkg/gui/helpers/snippet_helper.go`:

```go
type SnippetHelper struct {
    c *HelperCommon
}

func NewSnippetHelper(c *HelperCommon) *SnippetHelper {
    return &SnippetHelper{c: c}
}

func (self *SnippetHelper) ListSnippets() error { ... }
func (self *SnippetHelper) CreateSnippet() error { ... }
func (self *SnippetHelper) DeleteSnippet() error { ... }
func (self *SnippetHelper) EditorTab() error { ... }
func (self *SnippetHelper) EditorEnter(nameContent string, expansionContent string) error { ... }
func (self *SnippetHelper) EditorEsc(completionActive bool) error { ... }
func (self *SnippetHelper) CloseEditor() error { ... }
```

2. Add to `Helpers` aggregator and `IHelpers`.

3. The helper needs access to `gui.config.Abbreviations` and `gui.config.Save()`. Add `Config() *config.Config` to `IGuiCommon` and implement on `*Gui`. Or pass config through `HelperCommon` constructor (it already receives `ruinCmd` — adding `config` is consistent).

4. `snippetExpansionTriggers` and `acceptSnippetParentCompletion` are completion-engine functions that need `*gocui.View` access. These can remain as view-level helpers called from controller binding closures, or the helper can accept extracted content. The cleanest approach: these functions move to the snippet helper, and the binding closures in `setupSnippetEditorContext` extract the view and pass it in.

5. Move `SnippetEditorFocus int` and `SnippetEditorCompletion *CompletionState` from `GuiState` to `SnippetEditorContext`.

6. Update `commands.go` palette commands to call `gui.helpers.Snippet().ListSnippets()`, etc.

7. Delete `handlers_snippets.go`.

### Dependency: completion engine access

Several snippet methods call completion functions: `isAbbreviationCompletion`, `isParentCompletion`, `acceptCompletion`, `acceptSnippetParentCompletion`. These are currently `*Gui` methods in `completion.go`. For the helper to call them, either:

- **(a)** Add the needed completion functions to `IGuiCommon` (e.g., `AcceptCompletion(v, state, triggers)`, `IsParentCompletion(v, state) bool`).
- **(b)** Extract the completion functions into a standalone `completion` package with no `*Gui` receiver. They only need `*gocui.View` and `*CompletionState`.

Option (b) is cleaner but is a larger refactor. For this plan, use option (a): add 4 methods to `IGuiCommon`:
- `AcceptCompletion(viewName string, state *CompletionState, triggers []CompletionTrigger)`
- `IsParentCompletion(viewName string, state *CompletionState) bool`
- `IsAbbreviationCompletion(viewName string, state *CompletionState) bool`
- `AcceptParentCompletion(viewName string, state *CompletionState)`

The `*Gui` implementations delegate to the existing private methods, passing `gui.g.View(viewName)`.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (create snippet, list snippets, delete snippet)

---

## Step 6: Migrate Palette Handlers to `PaletteHelper`

### Current state

`handlers_palette.go` contains ~374 lines with these methods on `*Gui`:
- `paletteCommands` — builds command list from controllers + palette-only commands
- `isPaletteCommandAvailable` (standalone function)
- `openPalette` — guards popup, builds commands, pushes context
- `closePalette` — resets state, pops context
- `executePaletteCommand` — runs selected command
- `filterPaletteCommands` — filters by substring
- `paletteSelectMove` — moves selection, renders
- `scrollPaletteToSelection` — scroll management
- `paletteEnter`, `paletteEsc`, `paletteListClick` — controller callbacks
- `renderPaletteList` — renders filtered list to view
- `quickOpenItems` — builds quick-open entries from notes/tags/queries/parents
- `filterQuickOpenItems` — filters quick-open by substring

Callers:
- `gui.go:setupGlobalContext` → `OnPalette: func() error { return gui.openPalette(gui.g, nil) }`
- `gui.go:setupPaletteContext` → `OnEnter`, `OnEsc`, `OnListClick` reference handler methods
- `palette_test.go` → `tg.gui.openPalette(tg.g, nil)` (7 sites)

### Action

1. Create `pkg/gui/helpers/palette_helper.go`:

```go
type PaletteHelper struct {
    c *HelperCommon
}

func NewPaletteHelper(c *HelperCommon) *PaletteHelper {
    return &PaletteHelper{c: c}
}

func (self *PaletteHelper) OpenPalette() error { ... }
func (self *PaletteHelper) ClosePalette() { ... }
func (self *PaletteHelper) ExecuteCommand() error { ... }
func (self *PaletteHelper) FilterCommands(filter string) { ... }
func (self *PaletteHelper) SelectMove(delta int) { ... }
func (self *PaletteHelper) QuickOpenItems() []PaletteCommand { ... }
func (self *PaletteHelper) FilterQuickOpenItems(filter string) { ... }
```

2. Add to `Helpers` aggregator and `IHelpers`.

3. `PaletteCommand` and `PaletteState` types need to be accessible from the helper. Move `PaletteCommand` and `PaletteState` from `state.go` to `helpers/palette_helper.go` (or a shared `types/palette.go` if needed by contexts). Keep a type alias in `state.go` temporarily if `GuiState.Palette` references the type, then update `GuiState` to use the new package path.

4. `renderPaletteList` needs `gui.views.PaletteList` to write to. Add `GetView(name) *gocui.View` (already on `IGuiCommon`) and render via that. The helper calls `gui.GetView("paletteList")`.

5. `paletteCommands` iterates `gui.contexts.All()` and calls `paletteOnlyCommands()`. The helper calls `gui.Contexts().All()` and a method on `IGuiCommon` that returns the palette-only commands. Add `PaletteOnlyCommands() []PaletteCommand` to `IGuiCommon`.

6. Move `PaletteSeedDone` from `GuiState` to `PaletteContext`.

7. Update `gui.go` bindings and `palette_test.go` to call helpers.

8. Delete `handlers_palette.go`. Keep `commands.go` (it still provides `paletteOnlyCommands()` and utility functions).

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (command palette, quick open)

---

## Step 7: Migrate Calendar Handlers to `CalendarHelper`

### Current state

`calendar.go` contains ~480 lines with ~25 methods on `*Gui`:
- `openCalendar`, `closeCalendar`
- `calendarSelectedDate`, `calendarRefreshNotes`
- `calendarGridLeft/Right/Up/Down/Enter/Click`
- `calendarInputEnter/Esc/Click`, `calendarFocusInput`
- `calendarNoteDown/Up/Enter`
- `calendarEsc`, `calendarTab`, `calendarBacktab`
- `renderCalendarGrid`, `renderCalendarNotes`

Callers:
- `gui.go:setupGlobalContext` → `OnCalendar`
- `gui.go:setupCalendarContext` → ~16 callback references

### Action

1. Create `pkg/gui/helpers/calendar_helper.go` with all calendar logic.

2. Add to `Helpers` aggregator and `IHelpers`.

3. Move `CalendarState` from `state.go` to `CalendarContext` (calendar state belongs with the calendar context, not cross-cutting `GuiState`).

4. The helper needs gocui view access for `renderCalendarGrid` and `renderCalendarNotes`. Use `gui.GetView(CalendarGridView)` etc. Move the rendering logic into the helper — it only needs the `*gocui.View` to clear and write to.

5. Update `gui.go:setupCalendarContext` to create closures that call the helper.

6. Update `gui.go:setupGlobalContext` `OnCalendar` to call `gui.helpers.Calendar().OpenCalendar()`.

7. Delete `calendar.go`.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (calendar flow: open, navigate, select day, input date, view notes)

---

## Step 8: Migrate Contrib Handlers to `ContribHelper`

### Current state

`contrib.go` contains ~400 lines with ~20 methods on `*Gui`:
- `openContrib`, `closeContrib`
- `contribLoadData`, `contribRefreshNotes`
- `contribGridLeft/Right/Up/Down/Enter`
- `contribNoteDown/Up/Enter`
- `contribEsc`, `contribTab`
- `renderContribGrid`, `renderContribNotes`

Callers:
- `gui.go:setupGlobalContext` → `OnContrib`
- `gui.go:setupContribContext` → ~11 callback references

### Action

1. Create `pkg/gui/helpers/contrib_helper.go` with all contrib logic.

2. Add to `Helpers` aggregator and `IHelpers`.

3. Move `ContribState` from `state.go` to `ContribContext`.

4. Update `gui.go` bindings and global controller callbacks.

5. Delete `contrib.go`.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (contrib flow: open, navigate, view notes)

---

## Step 9: Delete Type Aliases from `state.go` and `completion.go`

### Current state

**`state.go`** has these aliases (lines 10–51):
```go
type InputPopupConfig = types.InputPopupConfig
type ContextKey = types.ContextKey
type PreviewMode = context.PreviewMode
type PreviewLink = context.PreviewLink
type PreviewState = context.PreviewState
type NavEntry = context.NavEntry
const PreviewModeCardList = context.PreviewModeCardList
const PreviewModePickResults = context.PreviewModePickResults
```

**`completion.go`** has these aliases (lines 11–17):
```go
type CompletionItem = types.CompletionItem
type CompletionTrigger = types.CompletionTrigger
type ParentDrillEntry = types.ParentDrillEntry
type CompletionState = types.CompletionState
var NewCompletionState = types.NewCompletionState
```

**`dialogs.go`** has:
```go
type MenuItem = types.MenuItem
```

These aliases exist so code in the `gui` package can write `CompletionState` instead of `types.CompletionState`. After steps 1–8 reduced the handler code, far fewer files in `gui/` reference these types directly.

### Action

1. Find all uses of each alias within the `gui` package (excluding the alias definition itself).

2. Replace each bare type name with its qualified version:
   - `CompletionState` → `types.CompletionState` (or add `types` import)
   - `CompletionItem` → `types.CompletionItem`
   - `CompletionTrigger` → `types.CompletionTrigger`
   - `ParentDrillEntry` → `types.ParentDrillEntry`
   - `NewCompletionState()` → `types.NewCompletionState()`
   - `InputPopupConfig` → `types.InputPopupConfig`
   - `ContextKey` → `types.ContextKey`
   - `PreviewMode` → `context.PreviewMode`
   - `PreviewLink` → `context.PreviewLink`
   - `PreviewState` → `context.PreviewState`
   - `NavEntry` → `context.NavEntry`
   - `PreviewModeCardList` → `context.PreviewModeCardList`
   - `PreviewModePickResults` → `context.PreviewModePickResults`
   - `MenuItem` → `types.MenuItem`

3. Delete the alias lines from `state.go`, `completion.go`, and `dialogs.go`.

4. The `types` and `context` imports will already exist in most files. For files that don't have them, add the import.

### Scope

After steps 1–8, the remaining files in `gui/` that use these types are:
- `gui.go` — `ContextKey` in `pushContextByKey`, `replaceContextByKey`, `activateContext` signatures
- `gui_common.go` — scattered references
- `state.go` — `GuiState` struct fields use `ContextKey`, `CompletionState`, etc.
- `completion.go` — functions use `CompletionItem`, `CompletionTrigger`, `CompletionState`
- `completion_triggers.go` — uses `CompletionTrigger`
- `completion_candidates.go` — uses `CompletionItem`
- `dialogs.go` — uses `MenuItem`, `DialogState` references
- `render_preview.go` — uses `PreviewMode`, `PreviewLink`
- `layout.go` — uses `ContextKey` in view creation

This is a mechanical find-and-replace per type. Do each type as a sub-step to keep diffs small.

### Verification

- `go build ./...` after each sub-step

---

## Step 10: Delete `ContextKey` Constants and `mainPanelContexts` from `state.go`

### Current state

`state.go` lines 16–40 define 13 `ContextKey` constants and a `mainPanelContexts` map:

```go
const (
    NotesContext        ContextKey = "notes"
    QueriesContext      ContextKey = "queries"
    ...
)

var mainPanelContexts = map[ContextKey]bool{
    NotesContext: true, ...
}
```

The constants duplicate the key values already stored in each context's `BaseContext.key` field (passed via `NewBaseContextOpts{Key: "notes"}`). The `mainPanelContexts` map duplicates `BaseContext.GetKind()` — a context is a "main panel" if its kind is `SIDE_CONTEXT` or `MAIN_CONTEXT`.

### Action

**10A. Replace constant usage with context key references**

Throughout the `gui` package, replace:
```go
case NotesContext:       → case gui.contexts.Notes.GetKey():
gui.pushContextByKey(NotesContext) → gui.pushContextByKey(gui.contexts.Notes.GetKey())
```

For `NewGuiState()` which initializes `ContextStack: []ContextKey{NotesContext}`, use a string literal `"notes"` or pass the key from the context tree (the context tree is already initialized before `NewGuiState` is called — verify this).

**Actually:** `NewGuiState()` is called in `NewGui()` at line 49 (`state: NewGuiState()`), before the context setup calls on lines 57–69. So `NewGuiState()` can't reference `gui.contexts.Notes.GetKey()` yet. Use the string literal `"notes"` directly — this is the canonical key value and won't change. Or refactor `NewGuiState` to accept the initial context key as a parameter.

**10B. Replace `mainPanelContexts` with `GetKind()` check**

Replace `popupActive()`:

```go
func (s *GuiState) popupActive() bool {
    return !mainPanelContexts[s.currentContext()]
}
```

With a method that uses the context tree:

```go
// Move popupActive to *Gui where it can access the context tree
func (gui *Gui) popupActive() bool {
    ctx := gui.contextByKey(gui.state.currentContext())
    if ctx == nil {
        return false
    }
    kind := ctx.GetKind()
    return kind != types.SIDE_CONTEXT && kind != types.MAIN_CONTEXT
}
```

This requires changing `gui.state.popupActive()` calls to `gui.popupActive()` — there are approximately 5 call sites (handlers_capture.go:openCapture, handlers_pick.go:openPick, handlers_palette.go:openPalette, calendar.go:openCalendar, contrib.go:openContrib). After steps 2–8, these are all in helpers that call `gui.PopupActive()` through `IGuiCommon`, which already exists. So `popupActive` on `GuiState` becomes unused and can be deleted along with `mainPanelContexts`.

Check: `gui.overlayActive()` calls `gui.state.popupActive()`. Update to `gui.popupActive()`.

**10C. Delete constants and map**

Delete the `const ( ... )` block (lines 16–30) and `var mainPanelContexts` (lines 32–40) from `state.go`.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh`

---

## Step 11: Replace `activateContext` with `HandleFocus` Hooks

### Current state

`activateContext` (gui.go lines 388–418) sets the gocui view, re-renders all three list panels, then has a per-context switch for refresh + preview:

```go
func (gui *Gui) activateContext(ctx ContextKey) {
    viewName := gui.contextToView(ctx)
    gui.g.SetCurrentView(viewName)
    gui.renderNotes()
    gui.renderQueries()
    gui.renderTags()
    switch ctx {
    case NotesContext:
        gui.RefreshNotes(true)
        gui.helpers.Preview().UpdatePreviewForNotes()
    case QueriesContext:
        ...
    case TagsContext:
        ...
    case PreviewContext:
        gui.renderPreview()
    }
    gui.updateStatusBar()
}
```

After REGRESSION_REFACTOR_PLAN Step 3, `pushContext`/`popContext`/`replaceContext` call `HandleFocus`/`HandleFocusLost` on the context objects. But `activateContext` still duplicates what those hooks should do.

### Action

**11A. Register `HandleFocus` hooks on each context during setup**

In `gui.go:setupNotesContext`:
```go
notesCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
    gui.RefreshNotes(true)
    gui.helpers.Preview().UpdatePreviewForNotes()
})
```

In `gui.go:setupTagsContext`:
```go
tagsCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
    gui.RefreshTags(true)
    gui.helpers.Tags().UpdatePreviewForTags()
})
```

In `gui.go:setupQueriesContext`:
```go
queriesCtx.AddOnFocusFn(func(_ types.OnFocusOpts) {
    if gui.contexts.Queries.CurrentTab == context.QueriesTabParents {
        gui.RefreshParents(true)
        gui.helpers.Queries().UpdatePreviewForParents()
    } else {
        gui.RefreshQueries(true)
        gui.helpers.Queries().UpdatePreviewForQueries()
    }
})
```

In `gui.go:setupPreviewContext`:
```go
previewCtx.AddOnRenderToMainFn(func() {
    gui.renderPreview()
})
```

**11B. Simplify `activateContext`**

```go
func (gui *Gui) activateContext(key types.ContextKey) {
    viewName := gui.contexts.ViewNameForKey(key)
    gui.g.SetCurrentView(viewName)
    // Re-render lists to update highlight visibility
    gui.renderNotes()
    gui.renderQueries()
    gui.renderTags()
    gui.updateStatusBar()
}
```

The per-context switch is gone — the focus hooks handle refresh and preview updates. The three `renderX()` calls update selection highlighting (the active panel shows blue selection, inactive panels show plain text). This is needed regardless of which context gains focus.

**11C. Verify no double-refresh**

`pushContext` calls `activateContext` then `ctx.HandleFocus()`. The focus hook calls `RefreshNotes(true)`. `activateContext` no longer calls `RefreshNotes(true)`. So there's exactly one refresh per context push — correct.

### Verification

- `go build ./...`
- `go test ./...`
- `scripts/smoke-test.sh` (exercise all panel transitions — notes, tags, queries, preview, back)

---

## Step 12: Collapse Private Render/Refresh Wrappers

### Current state

`gui.go` and other files define private methods that are thin wrappers called only from `gui_common.go` public methods:

| Private method | Public method (gui_common.go) | Other internal callers |
|---|---|---|
| `gui.renderNotes()` | `gui.RenderNotes()` | `gui.go:activateContext` (after Step 11: 1 site), `render.go:renderQueries` (0), `gui.go:setupNotesContext` callback |
| `gui.renderTags()` | `gui.RenderTags()` | `gui.go:activateContext`, `gui.go:setupTagsContext` callback |
| `gui.renderQueries()` | `gui.RenderQueries()` | `gui.go:activateContext`, `gui.go:setupQueriesContext` callbacks |
| `gui.renderPreview()` | `gui.RenderPreview()` | `gui.go:activateContext`, `gui.go:backgroundRefreshData` |
| `gui.updateNotesTab()` | `gui.UpdateNotesTab()` | `layout.go` initial setup |
| `gui.updateTagsTab()` | `gui.UpdateTagsTab()` | `layout.go` initial setup |
| `gui.updateQueriesTab()` | `gui.UpdateQueriesTab()` | `layout.go` initial setup |
| `gui.updateStatusBar()` | `gui.UpdateStatusBar()` | `gui.go:activateContext`, `gui.go:backgroundRefreshData`, `statusbar.go:showError` |

### Action

For each pair, choose one of:
- **(a)** Delete the private method, change all internal callers to use the public method. This is the simplest approach and has no behavioral difference — the public method just calls the private one.
- **(b)** Inline the implementation into the public method and delete the private one.

Use approach (a) for all — it's purely mechanical.

1. In `gui_common.go`, change each passthrough to contain the actual implementation:

```go
// Before:
func (gui *Gui) RenderNotes() { gui.renderNotes() }

// After:
func (gui *Gui) RenderNotes() {
    // (move renderNotes body here)
}
```

Actually, `renderNotes()` is ~45 lines in `render.go`. Don't move it — instead, just delete the private method name and make the `render.go` functions the public methods directly. Since `render.go` and `gui_common.go` are in the same package, rename `renderNotes` to `RenderNotes` in `render.go` and delete the passthrough in `gui_common.go`.

Wait — `RenderNotes()` is defined in `gui_common.go` as a method on `*Gui`. `renderNotes()` is also a method on `*Gui` defined in `render.go`. Two methods with the same receiver can't have the same name. The solution: delete the `gui_common.go` line `func (gui *Gui) RenderNotes() { gui.renderNotes() }` and rename `renderNotes` to `RenderNotes` in `render.go`.

Do this for: `renderNotes→RenderNotes`, `renderTags→RenderTags`, `renderQueries→RenderQueries`, `renderPreview→RenderPreview`.

For `updateNotesTab`, `updateTagsTab`, `updateQueriesTab`, `updateStatusBar` — these are defined in `statusbar.go`. Same approach: rename private to public in `statusbar.go`, delete passthrough in `gui_common.go`.

For `showConfirm`, `showInput`, `showError` — these are defined in `dialogs.go` and `statusbar.go`. Rename to public, delete passthrough.

For `buildSearchOptions` — defined in `handlers.go`. Rename to `BuildSearchOptions`, delete passthrough in `gui_common.go`.

### Caveat

Check that renaming doesn't conflict. `gui_common.go` defines public methods satisfying `IGuiCommon`. After renaming, the implementation moves to `render.go`/`statusbar.go`/`dialogs.go` but the method signature is unchanged — it still satisfies the interface.

### Scope of changes

After steps 1–8, internal callers of the private methods are:
- `gui.go:activateContext` (after Step 11) — switch to public calls
- `gui.go:backgroundRefreshData` — switch to public calls
- `gui.go:setupNotesContext` callback — switch to public call
- `layout.go` — switch to public calls
- `statusbar.go:showError` — switch to public call
- `dialogs.go:closeDialog` — uses `gui.contextToView` (already fine)

### Verification

- `go build ./...`
- `go test ./...`

---

## Step 13: Slim Down `gui_common.go`

### Current state

After Step 12, `gui_common.go` no longer has render/statusbar/dialog passthroughs. What remains are:

1. **Truly necessary adapters** — methods where `gui_common.go` wraps a helper call and the method name differs:
   - `RefreshNotes(preserve bool)` → `gui.helpers.Notes().FetchNotesForCurrentTab(preserve)` (name differs)
   - `RefreshTags(preserve bool)` → `gui.helpers.Tags().RefreshTags(preserve)`
   - `RefreshQueries(preserve bool)` → `gui.helpers.Queries().RefreshQueries(preserve)`
   - `RefreshParents(preserve bool)` → `gui.helpers.Queries().RefreshParents(preserve)`
   - `RefreshAll()` → `gui.helpers.Refresh().RefreshAll()`

2. **Preview delegation methods** — 7 methods that forward to `PreviewHelper`:
   - `PreviewPushNavHistory`, `PreviewReloadContent`, `PreviewUpdatePreviewForNotes`, `PreviewUpdatePreviewCardList`, `PreviewCurrentCard`, `SetPreviewCards`, `SetPreviewPickResults`

3. **Completion candidate methods** — 3 methods forwarding to completion engine:
   - `TagCandidates`, `CurrentCardTagCandidates`, `ParentCandidatesFor`

4. **State accessors** — `SetSearchQuery`, `GetSearchQuery`, `SetSearchCompletion`, etc.

5. **Context navigation** — `PushContext`, `PopContext`, `ReplaceContext`, `PushContextByKey`, `ReplaceContextByKey`

### Action

These methods are structurally necessary — they satisfy `types.IGuiCommon`. Don't delete them. But:

1. **Remove the `Render()` duplicate** — `gui_common.go` has both `Render()` and `RenderAll()` doing the same thing. Delete `Render()` from `gui_common.go` and `types.IGuiCommon`, or make `Render` an alias for `RenderAll`. Check callers of `Render()` — if none exist outside the compile-time assertion, delete it from the interface.

2. **Inline trivial state accessors** — Methods like `SetSearchQuery`, `GetSearchQuery` that just read/write `gui.state.SearchQuery` can stay as-is (they're already one-liners and provide interface abstraction).

3. **Remove `BuildCardContent`** — This is a `gui_common.go` passthrough to `gui.buildCardContent`. After Step 12, the rendering function is public. If `BuildCardContent` on `IGuiCommon` is only called from `PreviewHelper`, and `PreviewHelper` could call it via `gui.Contexts()` or a render function, it could be simplified. But if it's used from helpers via `IGuiCommon`, keep it.

### Verification

- `go build ./...`
- `go vet ./...`

---

## Step 14: Clean Up `keybindings.go`

### Current state

`keybindings.go` has ~15 lines of "removed" comments (lines 32–44, 170–203) documenting which binding groups were migrated to controllers. These served as migration breadcrumbs but are now noise.

`globalNavBindings()` returns only 2 bindings (mouse wheel up/down for preview scrolling). These are global mouse bindings that could be registered through the `GlobalController` or `PreviewController` instead of the legacy `navBindings` path.

### Action

1. Delete all "removed" comments (lines 32–44, 170–203).

2. Move the 2 global mouse wheel bindings to `PreviewController.GetMouseKeybindingsFn()` with `ViewName: ""` (global scope). Delete `globalNavBindings()`.

3. After this, `setupKeybindings` only registers:
   - The SearchFilterView 'x' binding (clear search)
   - Dialog keybindings
   - Context/controller bindings
   - Tab click bindings

4. The SearchFilterView 'x' binding should move to a controller. SearchFilterView doesn't have its own context — it's a lightweight view that appears when a search query is active. Create a minimal context for it (or register the binding through `GlobalController` with `ViewName: SearchFilterView`). The simplest approach: add a `{Key: 'x', ViewName: SearchFilterView, Handler: ...}` binding to `GlobalController.GetKeybindingsFn()` with a disabled-reason that checks `!gui.SearchQueryActive()`.

### Verification

- `go build ./...`
- `scripts/smoke-test.sh`
- `./lazyruin --debug-bindings` diff (wheel bindings now show under preview controller, 'x' under global)

---

## Step 15: Update Documentation and Delete Completed Plans

### Action

1. Update `docs/ARCHITECTURE.md`:
   - Remove references to "hybrid migration", "handler files", "dual context stack"
   - Update package structure to reflect deleted files
   - Update interface boundary docs to show single `IGuiCommon` in `types/`
   - Document that focus hooks drive per-context refresh/preview logic
   - Remove "Two separate `IGuiCommon` interfaces" section — there's now one

2. Delete completed planning docs:
   - `docs/DRY_REFACTOR_PLAN.md`
   - `docs/REGRESSION_REFACTOR_PLAN.md`
   - `docs/FINAL_CLEANUP_PLAN.md` (this file)

3. Update `CLAUDE.md` if any guidance references deleted files or patterns.

### Verification

- Review `docs/ARCHITECTURE.md` for accuracy against codebase
- `go build ./...` (ensure no references to deleted files)

---

## Execution Order

```
Step 1   (delete handlers_note_actions.go)   ← Already dead code, instant win
Step 9   (delete type aliases)               ← Mechanical, zero-risk
Step 10  (delete ContextKey constants)        ← Depends on nothing
Step 2   (capture → CaptureHelper)           ← Independent handler file
Step 3   (pick → PickHelper)                 ← Independent handler file
Step 4   (input popup → InputPopupHelper)    ← Independent handler file
Step 5   (snippets → SnippetHelper)          ← Independent handler file, needs completion access
Step 6   (palette → PaletteHelper)           ← Independent handler file
Step 7   (calendar → CalendarHelper)         ← Independent handler file
Step 8   (contrib → ContribHelper)           ← Independent handler file
Step 11  (activateContext → focus hooks)      ← After all handler files deleted
Step 12  (collapse render wrappers)           ← After Step 11 simplifies activateContext
Step 13  (slim gui_common.go)                 ← After Step 12
Step 14  (clean keybindings.go)               ← Independent, do anytime
Step 15  (update docs)                        ← Last
```

Steps 2–8 are independent of each other and can be done in any order. Each creates one helper, migrates one handler file's callers, and deletes the handler file.

Steps 1, 9, 10, 14 can be done at any time — they have no dependencies on other steps.

---

## Files Deleted After This Plan

| File | Step |
|------|------|
| `pkg/gui/handlers_note_actions.go` | 1 |
| `pkg/gui/handlers_capture.go` | 2 |
| `pkg/gui/handlers_pick.go` | 3 |
| `pkg/gui/handlers_input_popup.go` | 4 |
| `pkg/gui/handlers_snippets.go` | 5 |
| `pkg/gui/handlers_palette.go` | 6 |
| `pkg/gui/calendar.go` | 7 |
| `pkg/gui/contrib.go` | 8 |
| `docs/DRY_REFACTOR_PLAN.md` | 15 |
| `docs/REGRESSION_REFACTOR_PLAN.md` | 15 |
| `docs/FINAL_CLEANUP_PLAN.md` | 15 |

## Files Created After This Plan

| File | Step |
|------|------|
| `pkg/gui/helpers/capture_helper.go` | 2 |
| `pkg/gui/helpers/pick_helper.go` | 3 |
| `pkg/gui/helpers/input_popup_helper.go` | 4 |
| `pkg/gui/helpers/snippet_helper.go` | 5 |
| `pkg/gui/helpers/palette_helper.go` | 6 |
| `pkg/gui/helpers/calendar_helper.go` | 7 |
| `pkg/gui/helpers/contrib_helper.go` | 8 |

## State Migrations (Context Ownership)

| State field(s) | From | To | Step |
|---|---|---|---|
| `PickQuery`, `PickAnyMode`, `PickSeedHash` | `GuiState` | `PickContext` | 3 |
| `InputPopupConfig`, `InputPopupSeedDone`, `InputPopupCompletion` | `GuiState` | `InputPopupContext` | 4 |
| `SnippetEditorFocus`, `SnippetEditorCompletion` | `GuiState` | `SnippetEditorContext` | 5 |
| `PaletteSeedDone`, `Palette` (PaletteState) | `GuiState` | `PaletteContext` | 6 |
| `Calendar` (CalendarState) | `GuiState` | `CalendarContext` | 7 |
| `Contrib` (ContribState) | `GuiState` | `ContribContext` | 8 |

After all state migrations, `GuiState` contains only:
- `Dialog *DialogState` — cross-cutting (confirm/menu/input dialogs can be triggered from any context)
- `ContextStack []types.ContextKey` — cross-cutting focus management
- `SearchQuery string` — cross-cutting (affects layout, search filter view visibility)
- `CaptureParent *CaptureParentInfo` — move to `CaptureContext` in Step 2
- `CaptureCompletion`, `SearchCompletion` — move to respective contexts in Step 2/Step 9 cleanup
- `Initialized`, `lastWidth`, `lastHeight` — layout bookkeeping (stays)

## Risk Notes

1. **Completion engine access (Steps 5, 2):** Several handler methods call completion engine functions (`acceptCompletion`, `isParentCompletion`, etc.) that are `*Gui` methods. The cleanest solution is adding these as `IGuiCommon` methods. An alternative is extracting completion functions to be standalone (receiver-free), but that's a larger refactor — defer to a future cleanup if needed.

2. **View access from helpers (Steps 6–8):** Palette, calendar, and contrib helpers need to write to gocui views for rendering. `IGuiCommon.GetView(name)` already provides this. The helpers use `v.Clear()`, `fmt.Fprintln(v, ...)`, `v.SetOrigin()`, etc. — standard gocui view operations that don't require the full `*Gui`.

3. **State migration to contexts (Steps 3–8):** Moving state from `GuiState` to context structs changes how `layout.go` accesses state for view creation. Layout currently reads `gui.state.Calendar`, `gui.state.Contrib`, etc. After migration, it reads `gui.contexts.Calendar.State`, etc. This is mechanical but touches `layout.go` which is a large file.

4. **`NewGuiState` initialization (Step 10):** After deleting `ContextKey` constants, `NewGuiState()` uses `[]types.ContextKey{"notes"}` for the initial stack. This is safe — the string "notes" is the canonical key value.

5. **Test updates (Steps 2–8):** Tests that call handler methods directly need updating to call helpers. The test setup (`testGui`) creates a mock executor and wires `*Gui`. After migration, tests call `tg.gui.helpers.Capture().OpenCapture()` instead of `tg.gui.openCapture(tg.g, nil)`. The assertions remain identical.
