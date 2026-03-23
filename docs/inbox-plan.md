# Inbox Plan

Quick-capture scratch pad for jotting down unrelated thoughts while writing a note. Items live in a local inbox until explicitly promoted to real notes.

## UX

### Inbox Input (from Capture)

1. While in the New Note dialog, press `<C-i>` → inbox input popup appears on top
2. Type a one-liner with completion support (`#` tags, `[[` wiki-links, `@` dates, `/` markdown)
3. Press `Enter` → item saved to inbox, popup closes, capture restored with content intact
4. Press `Esc` → popup dismissed without saving

Visual example:
```
╭─ New Note ─────────────────────────────── <c-s> to save ─╮
│ Working on the API refactor for the billing module...     │
│                                                           │
│  ╭─ Inbox ───────────────────────────────────────────╮    │
│  │ Remember to check #ops dashboard for alerts       │    │
│  ╰───────────────────────────────────────────────────╯    │
│                                                           │
╰──────────────────────────────────────────────── Parent: ─╯
```

### Inbox Browser

1. Press `i` (global) → inbox browser dialog opens
2. Navigate with `j`/`k`, scroll with `J`/`K`
3. Press `Enter` on an item → opens New Note prepopulated with the item's text, item removed from inbox
4. Press `d` on an item → deletes it from the inbox (with confirmation)
5. Press `Esc` → closes browser
6. Footer shows item count and position

Visual example:
```
╭─ Inbox ──────────────────────────────────────────────────╮
│  Remember to check #ops dashboard for alerts             │
│ ▌Ask Sarah about the deploy schedule #work               │
│  Look into [[Project Alpha]] timeline                    │
╰──────────────────────────────────────────── 2 of 3 items ╯
```

## Storage

### Why local file, not ruin notes?

Inbox items are scratch fragments, not real notes. Using `ruin log` with an `#inbox` tag would:
- Pollute the vault with one-liner fragments that aren't meaningful as standalone notes
- Require CLI round-trips for save/delete (slower for quick jots)
- Make "promote to note" awkward (delete old note, create new note with same content)

Local file storage keeps inbox items ephemeral and instant.

### `~/.config/lazyruin/inbox.json`

```json
[
  {"id": "a1b2c3", "text": "Remember to check #ops dashboard", "created": "2026-03-18T10:00:00Z"},
  {"id": "d4e5f6", "text": "Ask Sarah about deploy schedule #work", "created": "2026-03-18T10:05:00Z"}
]
```

- IDs are short random hex strings (6 chars, sufficient for local uniqueness)
- File is read on app start and on inbox open; written on every add/delete
- File is per-user (not per-vault) — inbox items are "me" state, not "vault" state

### `pkg/inbox/inbox.go` (~60 lines)

```go
type Item struct {
    ID      string    `json:"id"`
    Text    string    `json:"text"`
    Created time.Time `json:"created"`
}

type Store struct {
    path  string
    items []Item
}

func NewStore() *Store               // resolves path from XDG/home
func (s *Store) Load() error         // reads JSON file (no-op if missing)
func (s *Store) Save() error         // writes JSON file
func (s *Store) Add(text string)     // appends item with generated ID
func (s *Store) Delete(id string)    // removes item by ID
func (s *Store) Items() []Item       // returns all items (newest first)
func (s *Store) Len() int
```

## Inbox Input (Capture Overlay)

### How it works

The inbox input reuses the existing `InputPopupContext` — a `TEMPORARY_POPUP` with full completion support. It's pushed onto the context stack on top of `capture`.

**Problem**: The layout code (`layout.go:86-123`) uses a `switch` on `contextMgr.Current()`. When `inputPopup` is current, the capture view is deleted (line 130-134). This would destroy the in-progress note.

**Fix**: Add `ContextMgr.Contains(key)` and use it in layout to keep capture alive when inputPopup is stacked on top:

```go
// context_mgr.go
func (m *ContextMgr) Contains(key types.ContextKey) bool {
    m.mu.Lock()
    defer m.mu.Unlock()
    for _, k := range m.stack {
        if k == key {
            return true
        }
    }
    return false
}
```

```go
// layout.go — in the switch
case "inputPopup":
    if gui.contextMgr.Contains("capture") {
        gui.createCapturePopup(g, maxX, maxY)
    }
    gui.createInputPopup(g, maxX, maxY)

// layout.go — in the cleanup
if ctx != "capture" && !gui.contextMgr.Contains("capture") {
    g.DeleteView(CaptureView)
    g.DeleteView(CaptureSuggestView)
    gui.views.Capture = nil
}
```

### Trigger

The `<C-i>` binding is added to the capture popup controller (`gui_setup.go:185-199`):

```go
{Key: gocui.KeyCtrlI, Description: "Inbox", Handler: func() error {
    return gui.helpers.Inbox().OpenInboxInput()
}},
```

### `pkg/gui/helpers/inbox_helper.go` — `OpenInboxInput()`

```go
func (self *InboxHelper) OpenInboxInput() error {
    gui := self.c.GuiCommon()
    gui.Helpers().InputPopup().OpenInputPopup(&types.InputPopupConfig{
        Title:    "Inbox",
        Triggers: gui.InboxTriggers, // #, [[, @, /
        OnAccept: func(raw string, _ *types.CompletionItem) error {
            if raw == "" {
                return nil
            }
            self.store.Add(raw)
            return self.store.Save()
        },
    })
    return nil
}
```

### Completion triggers

Inbox input uses a subset of capture triggers — tags, wiki-links, dates, and markdown helpers. No abbreviations (`!`) or parent completion (`>`), since inbox items are quick scratch text.

```go
func (gui *Gui) inboxTriggers() []types.CompletionTrigger {
    // Reuse existing triggers: #, [[, @, /
    // Omit: ! (abbreviations), > (parents)
}
```

## Inbox Browser

### Context

New `InboxBrowserContext` as a `TEMPORARY_POPUP`:

```go
// pkg/gui/context/inbox_browser_context.go
type InboxBrowserContext struct {
    BaseContext
    Items       []inbox.Item
    SelectedIdx int
}
```

### Layout

A centered list view, similar to the menu dialog but using the context system for proper keybinding registration:

```go
// layout.go — new case in switch
case "inboxBrowser":
    gui.createInboxBrowser(g, maxX, maxY)

// new createInboxBrowser function
func (gui *Gui) createInboxBrowser(g *gocui.Gui, maxX, maxY int) error {
    // Centered, height based on item count (max 60% terminal height)
    // Renders items with selection highlight
    // Footer: "N of M items"
}
```

### Controller

```go
// pkg/gui/controllers/inbox_browser_controller.go
type InboxBrowserController struct {
    baseController
    c          *ControllerCommon
    getContext func() *InboxBrowserContext
}

func (self *InboxBrowserController) GetKeybindings(...) []*types.Binding {
    return []*types.Binding{
        {Key: 'j', Handler: self.nextItem},
        {Key: 'k', Handler: self.prevItem},
        {Key: gocui.KeyArrowDown, Handler: self.nextItem},
        {Key: gocui.KeyArrowUp, Handler: self.prevItem},
        {Key: gocui.KeyEnter, Handler: self.promoteItem},
        {Key: 'd', Handler: self.deleteItem},
        {Key: gocui.KeyEsc, Handler: self.close},
    }
}
```

### Promote flow

When Enter is pressed on an inbox item:

1. Read the item's text
2. Delete the item from the inbox store
3. Close the inbox browser (pop context)
4. Open the capture popup with the text pre-filled

```go
func (self *InboxBrowserController) promoteItem() error {
    ctx := self.getContext()
    if len(ctx.Items) == 0 {
        return nil
    }
    item := ctx.Items[ctx.SelectedIdx]
    self.c.Helpers().Inbox().DeleteItem(item.ID)
    self.c.GuiCommon().PopContext() // close browser
    self.c.Helpers().Capture().OpenCaptureWithContent(item.Text)
    return nil
}
```

This requires a new `OpenCaptureWithContent(text string)` method on `CaptureHelper` that opens the capture popup and pre-fills the text area. Similar to `OpenCaptureWithParent` but for content instead of parent metadata.

### Delete flow

When `d` is pressed:

1. Show confirmation dialog ("Delete this inbox item?")
2. On confirm: delete from store, refresh browser view
3. Adjust selection index if needed

### Global keybinding

```go
// global_controller.go
{ID: "global.inbox", Key: 'i', Handler: self.openInbox, Description: "Inbox", Category: "Global"},
```

Handler opens the inbox browser with current items:

```go
func (self *GlobalController) openInbox() error {
    return self.c.Helpers().Inbox().OpenBrowser()
}
```

`OpenBrowser()` loads items from the store, populates the context, and pushes it.

## Files Changed

| File | Change | Lines |
|------|--------|-------|
| `pkg/inbox/inbox.go` | New: Store, Item, CRUD operations, JSON persistence | ~60 |
| `pkg/gui/helpers/inbox_helper.go` | New: OpenInboxInput, OpenBrowser, DeleteItem | ~50 |
| `pkg/gui/helpers/helpers.go` | Register InboxHelper | ~3 |
| `pkg/gui/context/inbox_browser_context.go` | New: InboxBrowserContext | ~25 |
| `pkg/gui/context/context_tree.go` | Register inbox browser context | ~2 |
| `pkg/gui/controllers/inbox_browser_controller.go` | New: keybindings, promote, delete, navigation | ~70 |
| `pkg/gui/gui_setup.go` | Setup inbox browser context + controller; add `<C-i>` to capture bindings | ~15 |
| `pkg/gui/controllers/global_controller.go` | Add `i` keybinding for inbox browser | ~3 |
| `pkg/gui/context_mgr.go` | Add `Contains(key)` method | ~10 |
| `pkg/gui/layout.go` | Keep capture alive under inputPopup; add `inboxBrowser` case; cleanup | ~20 |
| `pkg/gui/views.go` | Add `InboxBrowserView` constant | ~1 |
| `pkg/gui/completion_triggers.go` | Add `inboxTriggers()` | ~10 |
| `pkg/gui/helpers/capture_helper.go` | Add `OpenCaptureWithContent(text)` | ~10 |

**Total: ~280 lines across ~13 files (3 new files, 10 edits).**

## Scope Boundaries

- **Inbox is local** — stored in config dir, not in the vault. No ruin CLI changes.
- **One-liners only** — inbox input is a single-line text field, not a multi-line editor.
- **No sync** — inbox is per-machine. If you need cross-device inbox, use a note tagged `#inbox` manually.
- **No editing** — items can be promoted or deleted, not edited in place. To change an item, promote it, edit in capture, and re-save.
- **Inbox browser is read-only** — no reordering, no bulk operations.

## Test Plan

### Unit tests (`pkg/inbox/inbox_test.go`)

1. `Add` appends item with generated ID and timestamp
2. `Delete` removes item by ID
3. `Items` returns newest-first order
4. `Save`/`Load` round-trips through JSON correctly
5. `Load` on missing file returns empty list (no error)
6. `Delete` with unknown ID is a no-op

### Unit tests (`pkg/gui/helpers/inbox_helper_test.go`)

1. `OpenInboxInput` configures InputPopup with correct title and triggers
2. `OnAccept` with empty text does not add item
3. `OnAccept` with text adds item to store

### Keybinding test (`scripts/keybinding-test.sh`)

New section:

```bash
# Inbox: <C-i> from capture, i for browser
echo "[7x] Inbox: C-i from capture"
send n; settle                          # open capture
send C-i                                # open inbox input
assert_contains "inbox input opens" "Inbox"
send -l "test thought"
send Enter; settle                      # save to inbox
assert_not_contains "inbox input closed" "Inbox"
send Escape; settle                     # close capture

echo "[7x] Inbox: i browser"
send i; settle                          # open inbox browser
assert_contains "inbox browser opens" "Inbox"
assert_contains "item visible" "test thought"
send d; settle                          # delete
send y; settle                          # confirm
send Escape; settle                     # close browser
```

### Manual smoke test

1. Open app, press `n` for New Note
2. Start typing a note, then press `<C-i>`
3. Verify inbox input appears on top of capture, capture content is visible behind it
4. Type `check #ops dashboard` — verify `#` triggers tag completion
5. Press Enter — inbox input closes, capture restored with original content intact
6. Press Esc to close capture
7. Press `i` — inbox browser opens showing the item
8. Press Enter — capture opens pre-filled with "check #ops dashboard", item removed from inbox
9. Press Esc, press `i` again — inbox is empty
