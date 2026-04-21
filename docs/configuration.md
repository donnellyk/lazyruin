# Configuration

Lazyruin stores its configuration in `~/.config/lazyruin/config.yml`.

The file is created on demand the first time a setting is persisted from the TUI (e.g. toggling a view option or completing onboarding). It is safe to hand-edit.

## Config file

```yaml
vault_path: ~/notes
editor: nvim
chroma_theme: catppuccin-mocha
view_options:
  hide_done: false
```

## Keys

| Key | Type | Default | Env Override | Description |
|-----|------|---------|-------------|-------------|
| `vault_path` | string | _(none)_ | `LAZYRUIN_VAULT` | Path to the notes vault directory |
| `editor` | string | `$EDITOR`, then `vim` | — | Command used when opening a note for editing |
| `chroma_theme` | string | `catppuccin-mocha` (dark) / `catppuccin-latte` (light) | — | [Chroma](https://github.com/alecthomas/chroma) style name used for preview syntax highlighting |
| `view_options.hide_done` | bool | `false` | — | Hide completed checkbox items in the preview pane |
| `disable_bare_url_as_link` | bool | `false` | — | When `true`, saving a New Note whose entire body is a URL takes the plain `ruin log` path instead of routing through the link-resolution flow |

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
