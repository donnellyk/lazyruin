# LazyRuin Keybindings

Complete keybinding reference for LazyRuin. All bindings are derived from the command table in `pkg/gui/commands.go` and navigation bindings in `pkg/gui/keybindings.go`.

## Design Principles

1. **Vim-style navigation** - `j/k` for up/down, `g/G` for top/bottom
2. **Mnemonic keys** - `d` for delete, `r` for rename, `n` for new
3. **Consistency with lazygit** - Same keys for similar operations
4. **Discoverable** - Status bar shows context-sensitive hints; `?` opens full reference

## Global

These work in any panel (suppressed when a dialog is open).

| Key | Action |
|-----|--------|
| `q`, `<c-c>` | Quit |
| `?` | Keybindings help |
| `/` | Search |
| `\` | Pick (tag filter) |
| `n` | New note (capture popup) |
| `:` | Command palette |
| `<c-r>` | Refresh all data |
| `Tab` | Next panel |
| `Shift+Tab` | Previous panel |
| `1` | Focus Notes (cycle tabs if already focused) |
| `2` | Focus Queries (cycle tabs if already focused) |
| `3` | Focus Tags (cycle tabs if already focused) |
| `p` | Pick |
| `0` | Focus Search Filter (when active) |

## Navigation (All Lists)

Standard navigation within Notes, Queries, and Tags panels.

| Key | Action |
|-----|--------|
| `j` / `Arrow Down` | Move down |
| `k` / `Arrow Up` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| Mouse wheel | Scroll list |
| Left click | Select item |

## Notes

| Key | Action |
|-----|--------|
| `Enter` | View in preview |
| `E` | Open in editor |
| `d` | Delete note |
| `y` | Copy note path |
| `t` | Add tag |
| `T` | Remove tag |
| `>` | Set parent |
| `P` | Remove parent |
| `b` | Toggle bookmark |
| `s` | Show info (parent tree, children, TOC) |

## Queries

| Key | Action |
|-----|--------|
| `Enter` | Run query |
| `d` | Delete query |

### Parents Tab

| Key | Action |
|-----|--------|
| `Enter` | View parent |
| `d` | Delete parent bookmark |

## Tags

| Key | Action |
|-----|--------|
| `Enter` | Filter notes by tag |
| `r` | Rename tag |
| `d` | Delete tag |

## Preview (Card List)

The preview panel displays either a single note or a card list (from search/tag/query).

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll line-by-line |
| `J` / `K` | Jump between cards |
| `}` / `{` | Next/prev header |
| `]` / `[` | Go forward/back in nav history |
| `l` / `L` | Highlight next/prev link |

### Actions

| Key | Action |
|-----|--------|
| `Enter` | Focus note (switch to Notes panel) |
| `Esc` | Back to previous panel |
| `E` | Open in editor |
| `d` | Delete card |
| `D` | Append #done to current line |
| `m` | Move card (persists order if `order` field exists) |
| `M` | Merge notes |
| `t` | Add global tag |
| `T` | Remove tag |
| `<c-t>` | Toggle inline tag on current line |
| `>` | Set parent |
| `P` | Remove parent |
| `b` | Toggle bookmark |
| `s` | Show info |
| `o` | Open highlighted link |
| `x` | Toggle todo checkbox |
| `f` | Toggle frontmatter |
| `v` | View options (title, global tags, markdown toggles) |

### Palette-Only Commands

| Command | Action |
|---------|--------|
| Toggle Title | Show/hide note title in cards |
| Toggle Global Tags | Show/hide global tags in cards |
| Toggle Markdown | Raw/rendered markdown |
| Order Cards | Persist current card order to `order` frontmatter |

## Date Preview

Shown on startup (today's date), or when selecting a date from Calendar/Contributions. Displays three sections: Inline Tags, Todos, and Notes.

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll line-by-line |
| `J` / `K` | Jump between cards |
| `)` / `(` | Jump to next/prev section |
| `}` / `{` | Next/prev header |
| `]` / `[` | Go forward/back in nav history |
| `l` / `L` | Highlight next/prev link |

### Actions

| Key | Action |
|-----|--------|
| `Enter` | Open selected card (pick results open at line, notes focus in Notes panel) |
| `Esc` | Back to previous panel |
| `E` | Open in editor |
| `d` | Delete card |
| `D` | Append #done to current line |
| `t` | Add global tag |
| `T` | Remove tag |
| `<c-t>` | Toggle inline tag on current line |
| `>` | Set parent |
| `P` | Remove parent |
| `b` | Toggle bookmark |
| `s` | Show info |
| `o` | Open highlighted link |
| `x` | Toggle todo checkbox |
| `f` | Toggle frontmatter |
| `v` | View options |

## Search

Activated with `/` from any panel.

| Key | Action |
|-----|--------|
| `Enter` | Execute search |
| `Tab` | Accept completion |
| `Esc` | Dismiss completion, or cancel search |
| `Arrow Down/Up` | Navigate completion items |

### Trigger Prefixes

Type these in the search box to activate completion:

| Prefix | Completes |
|--------|-----------|
| `#` | Tags |
| `!` | Abbreviation snippets |
| `created:` | Creation date filters |
| `updated:` | Update date filters |
| `before:` | Created-before filter |
| `after:` | Created-after filter |
| `between:` | Date range filter |
| `title:` | Title search |
| `path:` | Path search |
| `parent:` | Parent filter |
| `sort:` | Sort order |
| `/` | Show all available filters |

## Search Filter

When a search is active, the filter bar appears as panel `[0]`.

| Key | Action |
|-----|--------|
| `x` | Clear search filter |

## Capture (New Note)

Activated with `n` from any panel.

| Key | Action |
|-----|--------|
| `<c-s>` | Save note |
| `Esc` | Dismiss completion, or cancel |
| `Tab` | Accept completion |

### Trigger Prefixes

| Prefix | Completes |
|--------|-----------|
| `#` | Tags |
| `!` | Abbreviation snippets (expands inline) |
| `[[` | Wiki-links (notes, then headers with `#`) |
| `>` | Parent (drill into children with `/`) |
| `/` | Markdown formatting (headings, lists, etc.) |

## Pick

Activated with `\` from any panel. Filters notes by tag intersection.

| Key | Action |
|-----|--------|
| `Enter` | Execute pick |
| `Tab` | Accept completion |
| `<c-a>` | `--any` mode (intersection vs union) |
| `<c-t>` | `--todo` mode |
| `<c-l>` | `--all-tags` (all scoped inline tags, dialog only) |
| `Esc` | Dismiss completion, or cancel |

## Command Palette

Activated with `:` from any panel.

| Key | Action |
|-----|--------|
| `Enter` | Execute selected command |
| `Arrow Up/Down` | Navigate commands |
| `Esc` | Cancel |
| Type | Filter commands |

## Snippet Editor

Opened via command palette (Create Snippet).

| Key | Action |
|-----|--------|
| `Tab` | Toggle focus between name/expansion, or accept completion |
| `Enter` | Save snippet (or accept completion) |
| `Esc` | Dismiss completion, or close editor |

## Mouse

| Element | Left Click |
|---------|------------|
| List item | Select |
| Panel | Focus panel |
| Tab header | Switch tab |
| Snippet name/expansion | Focus field |
