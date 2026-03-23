# Refactor Candidates

Multi-agent codebase review. Findings organized by priority and area.

---

## Priority 1: High Impact

### 1.1 Generic List Refresh with Selection Preservation

**Files:** `notes_helper.go`, `queries_helper.go`, `tags_helper.go`

All three helpers implement identical refresh-and-preserve-selection logic:
1. Get previous selection ID
2. Call RuinCmd to fetch data
3. Update context items
4. Find previous selection by ID, restore index
5. Clamp selection
6. Render

**Action:** Extract `RefreshListWithPreservation[T]()` into `RefreshHelper`. Takes a load function, a context with `GetSelectedItemId/SetSelectedLineIdx/ClampSelection`, and a preserve flag. Eliminates ~100 lines.

---

### 1.2 ExecuteAndUnmarshal[T] for Command Layer

**Files:** All files in `pkg/commands/` (search.go, tags.go, queries.go, parent.go, pick.go, link.go)

Every command that returns parsed data repeats:
```go
output, err := cmd.ruin.Execute(args...)
if err != nil { return nil, err }
return unmarshalJSON[T](output)
```

**Action:** Add `ExecuteAndUnmarshal[T](args ...string) (T, error)` to `RuinCommand`. Every list/fetch method becomes a one-liner. Also fixes inconsistent error suppression in `Today()`, `Queries.Run()`, and `Tags.List()` which silently swallow JSON parse errors.

---

### 1.3 Note Command Helper in note.go

**Files:** `pkg/commands/note.go` (15 methods, 200+ LOC)

15 methods follow `_, err := n.ruin.Execute("note", "set", id, "-f", ...); return err`. The only variance is the trailing flags.

**Action:** Extract `executeNoteSet(id string, args ...string) error`. Cuts the file roughly in half.

---

### 1.4 Split layout.go (974 lines)

**File:** `pkg/gui/layout.go`

Contains sidebar views, popup creation, complex layouts, and orchestration all in one file.

**Action:** Split into:
- `layout.go` - Main orchestration (~230 lines)
- `layout_sidebar.go` - Notes/Queries/Tags/Preview/SearchFilter views
- `layout_popups.go` - Search/InputPopup/Capture/Pick/Palette popups
- `layout_complex.go` - SnippetEditor/InboxBrowser/Calendar/Contrib/PickDialog

---

### 1.5 Frame/Title Color Helper

**File:** `pkg/gui/layout.go` (43 occurrences)

Every view creation repeats:
```go
if gui.contextMgr.Current() == "contextName" {
    v.FrameColor = gocui.ColorGreen
    v.TitleColor = gocui.ColorGreen
} else {
    v.FrameColor = gocui.ColorDefault
    v.TitleColor = gocui.ColorDefault
}
```

**Action:** Extract `applyFocusFrameColors(v *gocui.View, contextKey string)`.

---

### 1.6 Sidebar List View Factory

**File:** `pkg/gui/layout.go` (lines 266-342)

Notes, Queries, Tags views use identical setup: `SetView`, store ref, `TitlePrefix`, `Tabs`, `SelFgColor`, `UpdateTab`, `setRoundedCorners`, `Highlight`, frame colors.

**Action:** Extract `createListView(name, prefix string, tabs []string, updateTab func())` helper.

---

### 1.7 PreviewContextTrait

**Files:** `cardlist_context.go`, `pickresults_context.go`, `compose_context.go`, `datepreview_context.go`

All four implement identical 6-method `IPreviewContext` delegation (NavState, DisplayState, SelectedCardIndex, SetSelectedCardIndex, CardCount, NavHistory). 24 lines of identical boilerplate.

**Action:** Create `PreviewContextTrait` struct that provides all 6 methods. Each context embeds it instead of reimplementing.

---

### 1.8 Preview Context Initialization Factory

**Files:** Same four preview contexts

All four initialize identically: `BaseContext + NavState{HighlightedLink: -1} + DisplayState{RenderMarkdown: true, DimDone: true} + navHistory`.

**Action:** Extract `NewPreviewContextBase(key, title, navHistory)` factory.

---

### 1.9 Context Registration Boilerplate in gui_setup.go

**File:** `pkg/gui/gui_setup.go` (493 lines)

4 preview contexts and 4 popup contexts repeat identical registration steps: create context, register, add focus handler, create controller, attach controller.

**Action:** Extract `registerPreviewContext()` and `registerPopupContext()` helpers.

---

## Priority 2: Medium Impact

### 2.1 CardListSource Factory

**Files:** `search_helper.go`, `queries_helper.go`, `tags_helper.go`, `pick_helper.go`, `preview_helper.go`

Six+ locations build nearly identical `CardListSource{Query, Requery}` with the same closure pattern around `BuildSearchOptions()`.

**Action:** Create `BuildSearchCardListSource(baseQuery, sort string) CardListSource` and `BuildTagCardListSource(tagName string) CardListSource`.

---

### 2.2 Filter/Clear Pattern in CardListFilterHelper

**File:** `pkg/gui/helpers/cardlist_filter_helper.go`

`openCardListFilter/applyCardListFilter/clearCardListFilter` are duplicated for PickResults (6 methods, ~120 lines doing the same thing for two contexts).

**Action:** Define a `Filterable` interface (`FilterText/SetFilterText/FilterActive/Source`) and write one generic filter handler.

---

### 2.3 Tab Cycling Logic

**Files:** `notes_helper.go`, `queries_helper.go`, `tags_helper.go`

All three implement identical tab cycling and switch-by-index:
```go
idx := (ctx.TabIndex() + 1) % len(tabs)
ctx.CurrentTab = tabs[idx]
ctx.SetSelectedLineIdx(0)
refresh()
```

**Action:** Extract generic `CycleTab(tabsList, onTabChange)` and `SwitchTabByIndex(tabs, idx, onTabChange)`.

---

### 2.4 NoteActionHandlersTrait

**Files:** `notes_controller.go`, `cardlist_controller.go`

Both have identical `addTag/removeTag/setParent/removeParent/toggleBookmark` methods delegating to `NoteActionsHelper`.

**Action:** Extract a `NoteActionHandlersTrait` with these 5 methods.

---

### 2.5 GridNavigationTrait

**Files:** `calendar_controller.go`, `contrib_controller.go`

Both implement identical grid navigation (left/right/up/down) with `MoveDay(delta)`. Only the delta values differ.

**Action:** Extract `GridNavigationTrait` parameterized by horizontal/vertical deltas.

---

### 2.6 Popup Centering Helper

**File:** `pkg/gui/layout.go` (4 locations)

Same centering math repeated in search, input, pick, and palette popup creation:
```go
width := 60; if width > maxX-4 { width = maxX-4 }
x0 := (maxX - width) / 2; y0 := (maxY-height)/2 - 2
```

**Action:** Extract `centerPopup(maxX, maxY, preferredWidth, preferredHeight) (x0, y0, x1, y1)`.

---

### 2.7 Conditional Argument Builder for Commands

**Files:** `search.go` (8 conditional appends), `pick.go` (6), `link.go` (5), `note.go` (3), `queries.go` (2)

Every command builds args with repetitive `if opts.X != "" { args = append(args, "--flag", opts.X) }`.

**Action:** Create fluent `ArgBuilder` with `AppendIf(condition, args...)` method.

---

### 2.8 Search Options Duplication

**Files:** `pkg/commands/search.go`, `pkg/commands/queries.go`

`--strip-global-tags`, `--strip-title`, `--content` flags are applied in 3 different places with different subsets. Which options apply where is implicit.

**Action:** Create `applyDisplayOptions(args, opts)` that centralizes which display flags get applied.

---

### 2.9 Completion Trigger Composition

**File:** `pkg/gui/completion_triggers.go`

`searchTriggers()`, `captureTriggers()`, `pickTriggers()`, `snippetExpansionTriggers()` define overlapping trigger slices.

**Action:** Define core triggers once, compose into context-specific sets to avoid repetition when adding new triggers.

---

### 2.10 Popup State Reset Pattern

**Files:** `search_helper.go`, `pick_helper.go`, `capture_helper.go`, `input_popup_helper.go`

All popup open/close handlers repeat: reset CompletionState, reset context fields, toggle cursor, push/pop context.

**Action:** Extract `OpenPopup(contextKey, resetFn)` and `ClosePopup(contextKey, cleanupFn)` in HelperCommon.

---

### 2.11 Delete Confirmation Pattern

**Files:** `notes_helper.go`, `queries_helper.go`, `tags_helper.go`, `preview_mutations_helper.go`, `link_helper.go`

All deletions: check nil, get display name, call `ConfirmDelete(type, name, deleteFn, onSuccess)`, refresh in onSuccess.

**Action:** Already partially abstracted via `ConfirmationHelper.ConfirmDelete()`. Consider adding `ConfirmDeleteEntity(entity, getDisplayName, deleteFn, refreshFn)` that handles the nil check + name extraction.

---

## Priority 3: Low Impact / Nice-to-Have

### 3.1 Completion Candidate Date Wrappers

**File:** `completion_candidates.go` (lines 231-246)

Four one-liner methods (`createdCandidates`, `updatedCandidates`, `beforeCandidates`, `afterCandidates`) just call `dateCandidates(prefix, filter)`.

**Action:** Use inline closures in trigger definitions instead of receiver methods.

---

### 3.2 Split completion_candidates.go (591 lines)

**File:** `pkg/gui/completion_candidates.go`

20 candidate functions mixing dates, references, tags, markdown.

**Action:** Split by domain: `completion_candidates_dates.go`, `completion_candidates_references.go`, etc.

---

### 3.3 FilterablePreviewTrait

**Files:** `cardlist_controller.go`, `pickresults_controller.go`

Both define identical `openFilter/clearFilter/filterNotActive` methods.

**Action:** Extract trait with these 3 methods.

---

### 3.4 Link-Note Check Utility

**Files:** `notes_controller.go`, `cardlist_controller.go`

Nearly identical `isLinkNote/notLinkNote` disabled-reason functions.

**Action:** Extract `requireLinkNote(provider func() *models.Note) func() *DisabledReason`.

---

### 3.5 Tag String Formatting in Models

**File:** `pkg/models/note.go`

`GlobalTagsString()` and `TagsString()` duplicate tag formatting logic (`if tag[0] != '#' { result += "#" }`).

**Action:** Extract `formatTag(tag) string` and `joinTags(tags) string` helpers.

---

### 3.6 Command Initialization Duplication

**File:** `pkg/commands/ruin.go`

`NewRuinCommand` and `NewRuinCommandWithExecutor` both initialize 7 sub-commands identically.

**Action:** Extract `initializeSubcommands()` method.

---

### 3.7 Nested Error Handling in parent.go

**File:** `pkg/commands/parent.go` (lines 60-68)

Three levels of nested `if err == nil` in `parseComposeResult`.

**Action:** Flatten with early returns or extract `enrichNoteMetadata()`.

---

### 3.8 QueriesController / InboxBrowserController Not Using ListControllerTrait

**Files:** `queries_controller.go`, `inbox_browser_controller.go`

Both reimplement `nextItem/prevItem` navigation instead of using `ListControllerTrait`.

**Action:** Adopt `ListControllerTrait` (or `SimpleCursorTrait` for InboxBrowser's integer index).

---

### 3.9 PreviewBindings Builder

**Files:** All preview controllers (CardList, Compose, PickResults, DatePreview)

All repeat: `bindings := self.NavBindings(); bindings = append(bindings, self.LineOpsBindings(prefix)...); bindings = append(bindings, custom...)`.

**Action:** Create `BuildPreviewBindings(prefix, customBindings)` in PreviewNavTrait.

---

## Known Constraints

### gocui blocks character-based bindings on editable views

`gocui.matchView()` returns false for any binding where `kb.ch != 0` on an editable view. This means `Alt+<letter>` bindings silently fail on capture, search, and other text input views — the keystroke falls through to the editor and types the character.

Only special keys (`Key`-based, `ch == 0`) work on editable views: `Ctrl-*`, `Tab`, `Esc`, `Enter`, arrow keys, function keys. This is why capture uses `Ctrl-J` (jot) and `Ctrl-O` (open inbox) instead of `Alt-j`/`Alt-i`. The palette editor works around this by handling Alt-j/k in its `Edit()` function directly.

Keep this in mind when adding new keybindings to editable contexts.

---

## Estimated Impact Summary

| Priority | Items | Est. Lines Eliminated | Risk |
|----------|-------|-----------------------|------|
| P1 (High) | 9 | ~500-600 | Low-Medium |
| P2 (Medium) | 11 | ~300-400 | Low |
| P3 (Low) | 9 | ~150-200 | Low |
| **Total** | **29** | **~950-1200** | |
