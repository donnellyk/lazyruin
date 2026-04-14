# Changelog

All notable changes to lazyruin are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Pre-1.0 breaking changes bump MINOR. Breaking CLI contract changes will be
called out here.

## [0.2.0] - 2026-04-14

### Added
- Type-ahead completion for dynamic embeds in the New Note capture popup.
  - Type `![[` to see search / pick / query / compose type prefixes

### Changed
- Updated minimum ruin-cli version to v0.2.0

## [0.1.0] - 2026-04-09

First tagged release. See [README.md](README.md) for installation and quick-start.

### Added

- TUI wrapper around the [ruin CLI](https://github.com/donnellyk/ruin-note-cli)
  for browsing, searching, and editing markdown notes
- Notes / Tags / Queries sidebar with tabbed views
- Preview pane with syntax highlighting and per-card layout
- Full-text search with tag / inline-tag completion
- Capture popup for new notes with optional parent selection
- Link capture flow: paste a URL, auto-resolve title / summary / tags
- Calendar and contributions heatmap views
- Quick jot inbox for ephemeral capture
- Command palette (`:`) with fuzzy search over all actions
- Keybinding reference popup (`?`)
- `--vault`, `--ruin`, `--new`, `--link`, `--link=<url>`, `--open`,
  `--version`, `--debug-bindings` CLI flags
- Runtime check that warns in the status bar if the installed `ruin` binary
  is below the minimum supported version

### Known limitations

- macOS and Linux only (no Windows)
- Requires the `ruin` CLI on `PATH` (installed automatically via Homebrew)
