# Keybindings

## Global

| Key | Action |
|-----|--------|
| `q` / `<c-c>` | Quit |
| `S` | Search |
| `p` | Pick (tag filter) |
| `n` | New Note |
| `<c-l>` | New Link |
| `c` | Calendar |
| `C` | Contributions |
| `<c-r>` | Refresh |
| `?` | Keybindings help |
| `:` | Command palette |
| `<c-o>` | Quick Open |
| `1` / `2` / `3` | Focus Notes / Queries / Tags (repeat to cycle tabs) |
| `0` | Focus Search Filter (when active) |
| `Tab` / `Shift-Tab` | Next / previous panel |

## List Navigation (Notes, Tags, Queries)

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `g` / `G` | Go to top / bottom |
| Arrow keys | Move down / up |
| Mouse wheel | Scroll |
| Left click | Select item |

## Notes

| Key | Action |
|-----|--------|
| `Enter` | View in preview (Notes tab) / Activate item (Home tab) |
| `E` | Open in external editor ($EDITOR) |
| `e` | Edit in popup (save with `<c-s>`, discard with `Esc`) |
| `d` | Delete note |
| `y` | Copy note path |
| `t` / `T` | Add / remove tag |
| `>` | Set parent |
| `P` | Remove parent |
| `b` | Toggle bookmark |
| `s` | Show info |
| `o` | Open URL |

When `notes_pane.sections_mode` is enabled (see [configuration.md](configuration.md#notes-pane-sections-mode)), the Notes pane gains a `Home`/`Notes` outer-tab toggle. Press `1` while focused on the pane to cycle outer tabs. On the Home tab `j`/`k` skip section headers; only `Enter` is meaningful (the note-action keys above are disabled and become available again on the Notes outer tab).

## Tags

| Key | Action |
|-----|--------|
| `Enter` | Filter notes by tag |
| `r` | Rename tag |
| `d` | Delete tag |

## Queries

| Key | Action |
|-----|--------|
| `Enter` | Run query / view parent |
| `d` | Delete query / parent |

## Preview

Shared across all preview modes: Card List, Pick Results, Compose, Date Preview.

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll line-by-line |
| `J` / `K` | Jump between cards |
| `{` / `}` | Previous / next header |
| `l` / `L` | Highlight next / previous link |
| `[` / `]` | Navigate history back / forward |
| `Enter` | Enter (focus note or open card) |
| `Esc` | Return focus to side pane |

### Actions

| Key | Action |
|-----|--------|
| `x` | Toggle todo checkbox |
| `D` | Toggle `#done` on current line |
| `<c-t>` | Toggle inline tag on current line |
| `<c-d>` | Toggle inline date on current line |
| `o` | Open highlighted link |
| `s` | Show info |
| `v` | View options |
| `<c-p>` | Pick (dialog) |

### Card List

| Key | Action |
|-----|--------|
| `E` | Open in external editor ($EDITOR) |
| `e` | Edit in popup (save with `<c-s>`, discard with `Esc`) |
| `d` | Delete card |
| `m` | Move card |
| `M` | Merge notes |
| `t` / `T` | Add / remove tag |
| `>` | Set parent |
| `P` | Remove parent |
| `b` | Toggle bookmark |
| `o` | Open URL |
| `R` | Re-resolve link |
| `F` | Filter cards |
| `X` | Clear filter |

### Pick Results

| Key | Action |
|-----|--------|
| `F` | Filter results |
| `X` | Clear filter |

### Compose

The cursor line resolves to a source child note via the compose `source_map`.
Blank lines and injected separators are unresolvable and Enter / `e` / `E`
no-op on those.

| Key | Action |
|-----|--------|
| `Enter` | Open source note of line under cursor in card list |
| `e` | Edit source note of line under cursor in popup |
| `E` | Open source note of line under cursor in `$EDITOR` |
| `<c-n>` | New child note |

### Date Preview

| Key | Action |
|-----|--------|
| `E` | Open in editor |
| `)` / `(` | Next / previous section |

### Palette-Only

| Command | Action |
|---------|--------|
| Toggle Frontmatter | Show/hide YAML frontmatter |
| Toggle Title | Show/hide note title in cards |
| Toggle Global Tags | Show/hide global tags |
| Toggle Markdown | Raw / rendered markdown |
| Toggle Dim Done | Dim completed items |
| Order Cards | Persist card order to frontmatter |

## Search

| Key | Action |
|-----|--------|
| `Enter` | Execute search |
| `Tab` | Accept completion |
| `Esc` | Dismiss completion or cancel |

### Completion Triggers

| Prefix | Completes |
|--------|-----------|
| `#` | Tags |
| `@` | Dates |
| `created:` | Creation date filter |
| `updated:` | Update date filter |
| `before:` | Created-before filter |
| `after:` | Created-after filter |
| `between:` | Date range filter |
| `title:` | Title search |
| `path:` | Path search |
| `parent:` | Parent filter |
| `sort:` | Sort order |
| `/` | Show all filters |

## New Note

| Key | Action |
|-----|--------|
| `<c-s>` | Save |
| `Tab` | Accept completion |
| `Esc` | Dismiss completion or cancel |

### Completion Triggers

| Prefix | Completes |
|--------|-----------|
| `#` | Tags |
| `[[` | Wiki-links (then `#` for headers) |
| `>` | Parent (then `/` to drill into children) |
| `@` | Dates |
| `/` | Markdown formatting |

## New Link

| Key | Action |
|-----|--------|
| `Enter` | Resolve URL then open editor |
| `<c-s>` | Save immediately (no resolve) |
| `Tab` | Accept completion |
| `Esc` | Cancel |

Tags can be added inline with `#` (e.g. `https://example.com #reading #tech`).

## Pick

| Key | Action |
|-----|--------|
| `Enter` | Execute pick |
| `Tab` | Accept completion |
| `<c-a>` | Toggle `--any` mode |
| `<c-t>` | Toggle `--todo` mode |
| `<c-l>` | Toggle `--all-tags` |
| `Esc` | Dismiss completion or cancel |

## Calendar

| Key | Action |
|-----|--------|
| `h` / `j` / `k` / `l` | Navigate grid |
| Arrow keys | Navigate grid |
| `Enter` | Open date preview |
| `/` | Focus date input |
| `Tab` / `Shift-Tab` | Cycle focus (grid, input, notes) |
| `Esc` | Close |

## Contributions

| Key | Action |
|-----|--------|
| `h` / `j` / `k` / `l` | Navigate grid |
| Arrow keys | Navigate grid |
| `Enter` | Open date preview |
| `Tab` | Cycle focus (grid, notes) |
| `Esc` | Close |

## Command Palette

| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| Arrow keys | Navigate |
| `Esc` | Cancel |
| Type | Filter commands |

## Mouse

| Element | Action |
|---------|--------|
| List item | Select |
| Panel | Focus |
| Tab header | Switch tab |
| Preview | Scroll (wheel), click to position |
