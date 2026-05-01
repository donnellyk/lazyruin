# Changelog

## [0.2.1] - 2026-05-01

### Changed
- `<c-d>` on a preview line now sets/replaces the inline date instead of toggling; `<c-x>` in the date dialog clears all dates on the line.

### Fixed
- Preview pane no longer breaks words mid-letter when wrapping lines containing hyphens.

## [0.2.0] - 2026-05-01 

### Added
- Experimental new side panel modeled off of expected Mac side pane with a more date-focused UX. Opt-in via `notes_pane.sections_mode: true`
- System for handling migrations via `ruin doctor`, when necessary.
- Ruin 0.4.0 support w/ new tag format
- Configurable sidebar width via `sidebar_width`
- Configurable preview pane padding via `preview_padding`

### Changed
- Rename 'Inbox' feature to 'Scratchpad' (`<c-j>` or `/scratchpad`), Inbox is now used in new experimental side panel.

## [0.1.0] - 2026-04-21

First tagged release. See [README.md](README.md) for installation and quick-start. Min ruin-cli version is v0.3.0
