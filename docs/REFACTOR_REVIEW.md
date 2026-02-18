# Refactor Plan Review

Oracle review of the controller refactoring plan with suggestions for improvement.

---

## Overall Assessment

The plan is strong and very close to lazygit's architecture, but it's a bit "all-in" for the current codebase. The main gaps are:

1. How contexts own/reflect state
2. How controllers and the existing Command table/palette coexist
3. How keybindings/mouse/overlays are wired
4. Avoiding unnecessary duplication in interfaces

With a few simplifications and clarifications, this can be landed safely and kept maintainable.

---

## Key Recommendations

### A. Clarify Context vs GuiState Ownership (High Leverage)

**Problem**: `GuiState` owns panel state (`NotesState`, `TagsState`, etc.). The plan introduces richer `Context` structs but doesn't specify where state lives. This creates risk of double sources of truth.

**Recommendation**: Let contexts *wrap* existing state rather than replace it initially.

```go
// pkg/gui/context/notes_context.go
type NotesContext struct {
    *BaseContext
    state *gui.NotesState  // wrap, don't replace
}

func NewNotesContext(common *ContextCommon, state *gui.NotesState) *NotesContext {
    ctx := &NotesContext{
        BaseContext: NewBaseContext(
            common,
            types.ContextKey("notes"),
            types.SIDE_CONTEXT,
            gui.NotesView,
            "Notes",
        ),
        state: state,
    }
    return ctx
}

func (c *NotesContext) Items() []models.Note {
    return c.state.Items
}
```

Same pattern for `TagsContext`, `QueriesContext`, `PreviewContext`, etc. Only once everything is migrated and stable, consider moving state fields directly into contexts.

---

### B. Keep the Command Table as the "Public API" for Actions

**Problem**: The existing `Command` table drives palette, keybindings, documentation, and hints. The plan moves *all* keybindings into controllers without mentioning the Command table—a big conceptual shift that complicates palette/hints.

**Recommendation**:

1. **Retain `commands.go` as the canonical list of user actions** (for palette and docs)

2. Let controllers focus on:
   - Navigation (j/k/g/G, arrows, mouse wheel)
   - Context-specific shortcuts that don't need palette entries
   - Maybe a subset of actions that map 1:1 to Command handlers

3. For actions used by both, extract into helpers:

   ```go
   // helpers/notes_helper.go
   func (h *NotesHelper) EditSelectedNote() error {
       note, ok := h.selectedNote()
       if !ok {
           return nil
       }
       return h.c.Editor.OpenInEditor(note.Path)
   }
   ```

   Command entry:
   ```go
   {Name: "Open in Editor", ..., Handler: func(g *gocui.Gui, v *gocui.View) error {
       return gui.helpers.Notes.EditSelectedNote()
   }}
   ```

   Controller binding:
   ```go
   {Key: 'E', Handler: c.withItem(c.helpers.Notes.EditSelectedNote), ...}
   ```

4. Palette stays driven by `[]Command`. Controllers don't need to know about palette at all.

---

### C. Simplify Interfaces: Avoid Double HasKeybindings

**Problem**: Both controllers and contexts implement `HasKeybindings`, which is confusing. You also have `AddKeybindingsFn` on context that accepts functions from controllers.

**Recommendation**:

1. **Let only contexts implement `HasKeybindings`**
2. Controllers *do not* implement `HasKeybindings` directly; they supply binding producer functions

Simplified controller interface:
```go
type IController interface {
    Context() Context
    GetKeybindingsFn() types.KeybindingsFn        // or nil
    GetMouseKeybindingsFn() types.MouseBindingsFn // or nil
    GetOnFocus() func(types.OnFocusOpts)          // or nil
    GetOnFocusLost() func(types.OnFocusLostOpts)  // or nil
    GetOnRenderToMain() func()                    // or nil
}
```

Context side:
```go
type IBaseContext interface {
    // aggregated
    GetKeybindings(opts KeybindingsOpts) []*Binding
    GetMouseKeybindings(opts KeybindingsOpts) []*gocui.ViewMouseBinding
    GetOnClick() func() error

    AddKeybindingsFn(KeybindingsFn)
    // ...
}
```

Updated `attachControllers`:
```go
func attachControllers(common *ControllerCommon, controllers ...types.IController) {
    for _, c := range controllers {
        ctx := c.Context()
        if f := c.GetKeybindingsFn(); f != nil {
            ctx.AddKeybindingsFn(f)
        }
        if f := c.GetOnFocus(); f != nil {
            ctx.AddOnFocusFn(f)
        }
        if f := c.GetOnFocusLost(); f != nil {
            ctx.AddOnFocusLostFn(f)
        }
    }
}
```

This clarifies responsibilities:
- **Context** = keybinding aggregator + lifecycle manager
- **Controller** = describes behavior and attaches itself to a single context

---

### D. Be Explicit About Binding to gocui Handler Mapping

**Problem**: `Binding.Handler` is `func() error` but gocui expects `func(g *gocui.Gui, v *gocui.View) error`.

**Recommendation**: Decide once how to bridge this in `keybindings.go`:

```go
func (gui *Gui) registerContextBindings() error {
    opts := types.KeybindingsOpts{GetKey: gui.keyConfig.Lookup}
    for _, ctx := range gui.contexts.All() {
        for _, b := range ctx.GetKeybindings(opts) {
            key := b.Key
            handler := b.Handler
            if err := gui.g.SetKeybinding(ctx.GetViewName(), key, gocui.ModNone,
                func(g *gocui.Gui, v *gocui.View) error {
                    if r := b.GetDisabledReason; r != nil && r() != nil {
                        return nil
                    }
                    return handler()
                },
            ); err != nil {
                return err
            }
        }
    }
    return nil
}
```

For handlers that genuinely need `*gocui.View` (e.g., click coordinates), treat them as mouse bindings with gocui's native signature.

---

### E. Don't Over-Abstract Mouse Handling on v1

**Problem**: Many existing mouse handlers depend on the view:

```go
func (gui *Gui) notesClick(g *gocui.Gui, v *gocui.View) error {
    idx := listClickIndex(v, 3)
    ...
}
```

**Recommendation**: Keep mouse bindings in `keybindings.go` for now, but call into controller/helper methods:

```go
// keybindings.go
{v, gocui.MouseLeft, gui.notesController.Click},

// controllers/notes_controller.go
func (c *NotesController) Click(g *gocui.Gui, v *gocui.View) error {
    idx := listClickIndex(v, 3)
    // update selection via c.context / c.state
}
```

Defer `GetMouseKeybindings`/`GetOnClick` until you actually need them consistently across many contexts.

---

### F. Integrate Overlays/Dialogs with Contexts Carefully

**Problem**: `OverlayType` + `overlayActive()` control when global/main-panel bindings are suppressed. The plan's `ContextKind` includes `PERSISTENT_POPUP` and `TEMPORARY_POPUP`, which could collide with current overlay suppression logic.

**Recommendation for v1**:

1. Use richer `Context` objects **only for main panels and core popups** (Search, Capture, maybe Palette)
2. Keep `OverlayType` and `overlayActive()` as-is initially
3. In new keybinding registration:
   - Wrap main-panel bindings with `suppressDuringDialog`, just like now
   - Popups/search/capture controllers should probably *not* be suppressed by `overlayActive()`

Only later consider:
- Representing overlays as proper contexts with `ContextKind` = `PERSISTENT_POPUP`/`TEMPORARY_POPUP`
- Replacing `overlayActive()` with "is there a popup context on the context stack?"

---

### G. Reuse listPanel Logic as the Basis for Traits

**Problem**: The plan designs traits from scratch, but working logic already exists.

**Recommendation**: When implementing `ListContextTrait` and `ListControllerTrait`, literally port the logic out of `listPanel` and `notesDown/notesUp/notesTop/notesBottom`:

```go
type listPanel struct {
    selectedIndex *int
    itemCount     func() int
    render        func()
    updatePreview func()
    context       ContextKey
}
```

You know that logic already works—don't rethink list behavior from scratch.

---

## Risks and Guardrails

| Risk | Guardrail |
|------|-----------|
| Two sources of truth for actions (Command handlers vs controller handlers) | Put all non-trivial logic into Helpers; call helpers from both; keep controllers and Command handlers thin |
| Confusing context/overlay interactions | In early phases, avoid changing `OverlayType` semantics; add tests to ensure dialogs still suppress global/main-panel shortcuts exactly as before |
| Interface creep (too many methods on `IController`/`Context`) | Keep `IController` small and explicit; let contexts own aggregation and binding APIs |
| Keybinding registration regressions | After wiring new binding registration, temporarily keep old `setupKeybindings` behind a feature flag and compare behavior; extend `smoke-test.sh` if needed |

---

## When to Consider the Advanced Path

It's worth revisiting a more complex, fully-lazygit-like design when:

- You want **user-configurable keybindings** and need a robust key config system
- You're adding **many more contexts/panels/popups**, and a uniform context tree becomes critical
- You need **fine-grained per-context disabling/enabling** of commands (e.g., advanced workflows or plugin systems)
- Palette and keybindings need to be fully DRY with zero duplication

At that point, it becomes worth:
- Moving all state into contexts
- Deriving the Command table views/contexts directly from controllers
- Fully representing overlays as contexts

---

## Optional Advanced Path (Full lazygit Parity)

If/when you decide to go "all in":

1. **Adopt a keybinding config layer**
   - Use `KeybindingsOpts.GetKey` to pull keys from config
   - Controllers declare "commands" (names, descriptions, default keys); a central keybinding manager resolves configured keys

2. **Generate the Command palette from controllers**
   - Each `Binding` gains `Category`, `Name`, and `ContextKey` fields
   - Palette scans all contexts/controllers, creates entries from bindings

3. **Unify overlays into the context tree**
   - `ContextTree` manages a real stack
   - Any context with kind `PERSISTENT_POPUP`/`TEMPORARY_POPUP` is treated as overlay
   - `overlayActive()` simply checks for popup contexts on top of the stack

4. **Move state into contexts**
   - Replace `GuiState.Notes` etc. with context-owned state
   - `GuiState` becomes primarily context stack + cross-cutting concerns

---

## Summary of Changes to Original Plan

| Original Plan | Revised Approach |
|---------------|------------------|
| Contexts replace state structs | Contexts wrap existing state initially |
| Controllers define all keybindings | Controllers handle nav; Command table remains canonical for actions |
| Both IController and Context implement HasKeybindings | Only Context implements HasKeybindings; controllers provide producer functions |
| Abstract mouse handling in interfaces | Keep mouse in keybindings.go, call into controllers |
| New overlay/popup context system | Keep OverlayType + overlayActive() for v1 |
| Design list traits from scratch | Port existing listPanel logic directly |
