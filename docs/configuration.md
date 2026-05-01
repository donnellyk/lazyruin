# Configuration

Lazyruin stores its configuration in `~/.config/lazyruin/config.yml`.

The file is created on demand the first time a setting is persisted from the TUI (e.g. toggling a view option or completing onboarding). It is safe to hand-edit.

## Config file

```yaml
vault_path: ~/notes
editor: nvim
chroma_theme: catppuccin-mocha
sidebar_width: 40
view_options:
  hide_done: false
notes_pane:
  sections_mode: false
```

## Keys

| Key | Type | Default | Env Override | Description |
|-----|------|---------|-------------|-------------|
| `vault_path` | string | _(none)_ | `LAZYRUIN_VAULT` | Path to the notes vault directory |
| `editor` | string | `$EDITOR`, then `vim` | — | Command used when opening a note for editing |
| `chroma_theme` | string | `catppuccin-mocha` (dark) / `catppuccin-latte` (light) | — | [Chroma](https://github.com/alecthomas/chroma) style name used for preview syntax highlighting |
| `sidebar_width` | int | `min(terminal_width / 3, 40)` | — | Width of the side panels in columns. Clamped at runtime to `[20, terminal_width - 20]` so the preview keeps a usable minimum. Set `0` or omit for the default. |
| `view_options.hide_done` | bool | `false` | — | Hide completed checkbox items in the preview pane |
| `disable_bare_url_as_link` | bool | `false` | — | When `true`, saving a New Note whose entire body is a URL takes the plain `ruin log` path instead of routing through the link-resolution flow |
| `notes_pane.sections_mode` | bool | `false` | — | Reshape the Notes pane into a `Home`/`Notes` outer-tab UX. When true, the four `All`/`Today`/`Recent`/`Links` sub-tabs are replaced; see [Notes pane sections mode](#notes-pane-sections-mode) below. |
| `notes_pane.custom_sections` | list | _(empty)_ | — | User-defined sections in the Home tab. Only consulted when `sections_mode` is `true`; see below. |

Additional internal fields (e.g. `onboarding_offered`) are managed automatically by the TUI; they are written back to the file but not intended for hand-editing.

## Vault path resolution

Lazyruin resolves the vault path in this order, stopping at the first match:

1. `--vault /path/to/vault` CLI flag
2. `vault_path` in `~/.config/lazyruin/config.yml`
3. `LAZYRUIN_VAULT` environment variable
4. The vault configured in ruin itself (`ruin config vault_path`)

If none resolve, lazyruin exits with an error at startup.

## Editor

The editor is used when opening a note from the TUI (for example with `e`). Resolution order:

1. `editor` from `config.yml`
2. `$EDITOR` environment variable
3. `vim`

The value may include arguments (e.g. `code --wait`); it is split on whitespace and executed with the note path appended.

## Chroma theme

`chroma_theme` accepts any style name supported by Chroma — see the [style gallery](https://xyproto.github.io/splash/docs/all.html). Unknown names fall back to Chroma's default style. If unset, lazyruin auto-picks a Catppuccin variant based on whether the terminal reports a dark or light background.

## View options

`view_options.hide_done` is toggled from the TUI (see `docs/keybindings.md`) and persisted here so the choice survives restarts. Editing the value directly has the same effect on the next launch.

## Notes pane sections mode

When `notes_pane.sections_mode` is `true`, the Notes pane swaps from a single flat list (with `All`/`Today`/`Recent`/`Links` sub-tabs) to a two-tab UX:

- **Home** — a list of selectable items grouped into sections. Activating an item runs a query and commits the result to Preview.
- **Notes** — a flat list equivalent to today's `All` sub-tab. Same keybindings as the legacy Notes pane.

The Home tab always shows three hardcoded items (Inbox, Today, Next 7 Days) plus a Pinned section with all saved parents and saved queries. Custom sections come from `notes_pane.custom_sections`.

### Custom sections

Each custom item maps to a [dynamic embed](https://github.com/donnellyk/ruin-note-cli/blob/main/docs/cli-reference.md#embed) string evaluated by `ruin embed eval`. The embed string is the same syntax accepted in real notes — author and preview an embed in a note, copy the string, paste it into config.

```yaml
notes_pane:
  sections_mode: true
  custom_sections:
    - title: Reading Queue
      items:
        - title: Articles to read
          embed: "![[search: #article !#done | limit=20]]"
        - title: Long-form
          embed: "![[search: #article #longform !#done]]"
    - title: Daily Review
      items:
        - title: Yesterday's open todos
          embed: "![[search: created:yesterday todo:open]]"
        - title: Followups this past week
          embed: "![[search: #followup between:today-7,today | sort=created:desc]]"
```

Field semantics:

- `title` (section): optional. If omitted, items render with no header.
- `items[].title`: required. Displayed verbatim in the Home tab.
- `items[].embed`: required. A complete `![[…]]` embed string (`search:`, `pick:`, `query:`, or `compose:`). Items missing a title or embed are silently skipped at load.

ruin's native date tokens (`@today`, `@yesterday`, `@2026-04-27`, `between:today-7,today`, etc.) work inside embed strings — there is no separate variable substitution layer in lazyruin.

### Tab switching

In sections_mode, pressing `1` while focused on the Notes pane cycles the outer tab (Home ↔ Notes). Clicking either tab label with the mouse switches directly. The legacy four sub-tabs are not rendered.
