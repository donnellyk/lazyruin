# lazyruin

[![CI](https://github.com/donnellyk/lazyruin/actions/workflows/ci.yml/badge.svg)](https://github.com/donnellyk/lazyruin/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/donnellyk/lazyruin)](https://github.com/donnellyk/lazyruin/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/donnellyk/lazyruin)](go.mod)
[![License](https://img.shields.io/github/license/donnellyk/lazyruin)](LICENSE)

A terminal UI for the [`ruin`](https://github.com/donnellyk/ruin-note-cli)
note-taking CLI. 

Don't organize; compose. Write small, atomic notes saved as simple markdown files. Later, compose them into different "documents", depending on your needs. This probably has an audience of one (me) or zero (turns out, not me); time will tell.

<!-- TODO: add screenshot or asciinema demo -->

## Installation

### Homebrew (recommended)

```
brew install donnellyk/ruin/lazyruin
```

This installs both `lazyruin` and the `ruin` CLI it depends on.

### `go install`

Requires Go 1.26+:

```
go install github.com/donnellyk/lazyruin@latest
```

You'll need the `ruin` CLI on `PATH` separately.

### From a local checkout

Requires Go 1.26+ and [mise](https://mise.jdx.dev):

```
git clone https://github.com/donnellyk/lazyruin
cd lazyruin
mise run install
```

## Quick start

```
$ lazyruin
```

First launch opens the TUI against the vault configured via `ruin config
vault_path`. Override with `--vault /path/to/vault` or the `LAZYRUIN_VAULT`
environment variable.

Core keys:

| Key     | Action               |
| ------- | -------------------- |
| `q`     | Quit                 |
| `S`     | Search notes         |
| `n`     | New note             |
| `<c-l>` | New link             |
| `p`     | Pick (tag filter)    |
| `:`     | Command palette      |
| '::'    | Quick Open           |
| `?`     | Keybindings help     |
| `Tab`   | Next panel           |

Quick direct-launch modes, which exit on save or close:

```
lazyruin --new                 # straight into new-note capture
lazyruin --link                # open the new-link input popup
lazyruin --link=https://...    # resolve the URL directly
```

See [`docs/keybindings.md`](docs/keybindings.md) for the full reference.

## Documentation

- [`docs/keybindings.md`](docs/keybindings.md) — full keybinding reference
- [`docs/architecture.md`](docs/architecture.md) — architecture and layer overview
- [`docs/abstractions.md`](docs/abstractions.md) — reusable abstraction patterns
- [`CHANGELOG.md`](CHANGELOG.md) — release notes

## License

MIT. See [LICENSE](LICENSE).

## AI Usage
Claude Code was used extensively on this project. All code was read, tested, reviewed, and committed by a human.
