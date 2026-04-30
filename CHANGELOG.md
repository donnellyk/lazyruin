# Changelog

## [Unreleased]

### Added
- **Notes pane sections mode** (opt-in via `notes_pane.sections_mode: true`). Replaces the four `All`/`Today`/`Recent`/`Links` sub-tabs with a `Home`/`Notes` outer-tab UX. The Home tab lists hardcoded items (Inbox, Today, Next 7 Days, Pinned), plus user-defined sections from `notes_pane.custom_sections`. Each custom item runs a `ruin embed eval` against an `![[…]]` string and commits the result to Preview. See [configuration.md](docs/configuration.md#notes-pane-sections-mode).
- `pkg/commands/embed.go` — typed wrapper around `ruin embed eval --json`, dispatching the result envelope by embed type (`search`/`query` → notes, `pick` → pick results, `compose` → expanded markdown + source map).
- **Upgrade migrations**. Lazyruin now records the lazyruin / ruin-cli version pair seen on each launch in `~/.config/lazyruin/state.json`. When a registered migration matches the upgrade path, a blocking modal prompts to run `ruin doctor`; the user can `[y] Run now` or `[q] Quit`. The migration registry starts empty — entries are appended only when a real breaking change ships that requires re-indexing the vault. Dev builds (`version == "dev"`) suppress the prompt entirely. See [docs/internal/upgrade-migrations-plan.md](docs/internal/upgrade-migrations-plan.md).

### Changed
- The local "Inbox" feature (scratch-pad jot via `<c-j>` from capture, browser via `i`) was renamed to **Scratchpad** to free the name for the new Inbox section. The slash command `/inbox` becomes `/scratchpad`. Storage path moves from `~/.config/lazyruin/inboxes/<hash>.json` to `~/.config/lazyruin/scratchpads/<hash>.json`; the directory is renamed in place on first launch when only the legacy path exists. Keybindings are unchanged.

## [0.1.0] - 2026-04-21

First tagged release. See [README.md](README.md) for installation and quick-start. Min ruin-cli version is v0.3.0
