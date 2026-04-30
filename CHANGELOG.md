# Changelog

## [Unreleased]

### Added
- Experimental new side panel modeled off of expected Mac side pane with a more date-focused UX. Ppt-in via `notes_pane.sections_mode: true`
- System for handling migrations via `ruin doctor`, when necessary.

### Changed
- The local "Inbox" feature (scratch-pad jot via `<c-j>` from capture, browser via `i`) was renamed to **Scratchpad** to free the name for the new Inbox section. The slash command `/inbox` becomes `/scratchpad`. Storage path moves from `~/.config/lazyruin/inboxes/<hash>.json` to `~/.config/lazyruin/scratchpads/<hash>.json`; the directory is renamed in place on first launch when only the legacy path exists. Keybindings are unchanged.

## [0.1.0] - 2026-04-21

First tagged release. See [README.md](README.md) for installation and quick-start. Min ruin-cli version is v0.3.0
