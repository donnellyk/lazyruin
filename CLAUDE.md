# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Important
- Always use `--vault /tmp/ruin-test-vault` (create it first with `ruin dev seed /tmp/ruin-test-vault` if needed). Never run against the user's real vault.
- Never modify `~/.config/ruin`. Use `--vault`. Do not run `config vault_path ~/path`. Stop before changing vault path.

## Project Overview

LazyRuin is a TUI (Terminal User Interface) for the `ruin` notes CLI, heavily inspired by lazygit's architecture and UX patterns. It uses the jesseduffield/gocui framework.

## Key Documentation

- `docs/ARCHITECTURE.md` - Layer architecture, package structure, data flow
- `docs/UI_MOCKUPS.md` - Visual mockups and responsive layouts
- `docs/KEYBINDINGS.md` - Complete keybinding reference
- `docs/PROJECT_PLAN.md` - Phased implementation plan

## Module

`kvnd/lazyruin`

# Build Commands

```bash
# Build
go build -o lazyruin ./main.go

# Run
./lazyruin
./lazyruin --vault /path/to/vault

# Test
go test ./...

# Lint
golangci-lint run
```

## Architecture Pattern

Follow lazygit's layered architecture:

```
App → Gui → Controllers → Helpers → Commands → Models
```

- **Controllers** handle user input and keybindings
- **Helpers** encapsulate domain logic (reusable operations)
- **Commands** wrap ruin CLI execution with typed responses
- **Models** are data structures (Note, Tag, Query)

## Key Patterns

1. **Context System**: Each panel has a Context managing state and keybindings
2. **Null Object Controllers**: Base controller returns nil, children override selectively
3. **Trait Composition**: ListContextTrait for common list behavior
4. **Thread-Safe Updates**: Use `gui.Update()` for goroutine GUI updates
5. **JSON Mode**: All ruin commands use `--json` for reliable parsing

## ruin CLI Integration

The TUI wraps these ruin commands:

```bash
ruin search "<query>" --json          # Search notes
ruin today --json                     # Today's notes
ruin log "<content>" --json           # Create note
ruin tags list --json                 # List tags
ruin tags rename <old> <new>          # Rename tag
ruin query list --json                # List saved queries
ruin query run <name> --json          # Run saved query
```

## Manual Testing

Use tmux to manually test the TUI:

```bash
go build -o /tmp/lazyruin-test ./main.go
tmux new-session -d -s test -x 120 -y 40 '/tmp/lazyruin-test --vault /private/tmp/ruin-test-vault'
tmux capture-pane -t test -p          # screenshot
tmux send-keys -t test <key>          # send keystrokes
tmux kill-session -t test             # cleanup
```

## Smoke Test

`scripts/smoke-test.sh` runs an automated TUI regression check via tmux (45 assertions across 18 categories). Keep this script up to date when adding/changing UI flows, keybindings, panel titles, status bar hints, or popup dialogs.

```bash
go build -o /tmp/lazyruin-test ./main.go && ./scripts/smoke-test.sh
```

## Coding Guidelines

- Use jesseduffield/gocui (not jroimartin/gocui)
- Use `v.InnerSize()` instead of `v.Size()` when calculating renderable width/height — `Size()` includes the frame borders and will cause off-by-2 clipping
- Support both keyboard and mouse navigation
- All list panels support j/k navigation, as well as up / down arrows
- Keybindings should be configurable via YAML
- Handle terminal resize gracefully
- Use `--json` for all ruin command parsing
