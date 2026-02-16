# LazyRuin Project Plan

A comprehensive implementation plan for LazyRuin, a TUI for the `ruin` notes CLI.

## Project Overview

**Goal**: Build a terminal user interface for the `ruin` note-taking CLI, following lazygit's architecture and UX patterns.

**Key Technologies**:
- Go 1.21+
- github.com/jesseduffield/gocui (TUI framework)
- ruin CLI (underlying note management)

## Phase 1: Foundation ✅

### 1.1 Project Setup

- [x] Initialize Go module (`go mod init kvnd/lazyruin`)
- [x] Set up directory structure per ARCHITECTURE.md
- [x] Add gocui dependency
- [x] Create basic main.go entry point
- [x] Set up mise.toml with build/test/run targets

### 1.2 Configuration System

- [x] Define config struct (`pkg/config/config.go`)
- [x] Load from `~/.config/lazyruin/config.yml`
- [x] Implement default configuration
- [x] Add vault path resolution (from ruin config or CLI flag)
- [x] Support command-line flag overrides (`--vault`, `--ruin`, `--new`)

### 1.3 Command Layer

- [x] Create base RuinCommand struct (`pkg/commands/ruin.go`)
- [x] Implement command execution with JSON parsing
- [x] Add error handling and timeout support
- [x] Create typed command wrappers:
  - [x] SearchCommand (search, today, yesterday)
  - [x] NoteCommand (set, append, merge)
  - [x] TagsCommand (list, rename, delete)
  - [x] QueryCommand (save, list, run, delete)
  - [x] ParentCommand (children, tree, save, delete, list)
  - [x] PickCommand (pick)
  - [ ] VaultCommand (vault, config, doctor)

### 1.4 Models

- [x] Note model with frontmatter fields
- [x] Tag model with count
- [x] Query model (name + query string)
- [x] SearchResult model

## Phase 2: Basic GUI ✅

### 2.1 Core GUI Setup

- [x] Create Gui struct wrapping gocui.Gui
- [x] Implement Run() with Init and Close
- [x] Set up main event loop
- [x] Configure mouse and cursor support
- [x] Handle terminal resize

### 2.2 View Creation

- [x] Define view constants and types
- [x] Implement view creation functions
- [x] Configure view properties (frame, title, wrap, etc.)
- [x] Set up view positioning helpers

### 2.3 Basic Layout

- [x] Implement layout manager
- [x] Create sidebar column (Notes, Queries, Tags)
- [x] Create main content area (Preview)
- [x] Handle responsive sizing
- [x] Implement panel focus indicators

### 2.4 Initial Rendering

- [x] Render dynamic content from ruin CLI
- [x] Verify panel switching with Tab
- [x] Verify quit with 'q'
- [x] Verify help display with '?'

## Phase 3: Context System ✅

### 3.1 Context Infrastructure

- [x] Define ContextKey type
- [x] Implement context key constants
- [x] Create state management structs

### 3.2 Panel Contexts

- [x] NotesContext with selection state
- [x] TagsContext with selection state
- [x] QueriesContext with selection state
- [x] PreviewContext with scroll state
- [x] SearchContext with input state
- [x] CaptureContext
- [x] PickContext
- [x] PaletteContext
- [x] SearchFilterContext

### 3.3 Context Management

- [x] Context switching logic
- [x] Focus handling via setContext()
- [x] Keybinding attachment per context

## Phase 4: Controllers ✅

### 4.1 Controller Infrastructure

- [x] Keybindings in dedicated file
- [x] Handlers in dedicated file
- [x] Per-view keybinding setup

### 4.2 Global Controller

- [x] Quit handling (q, Ctrl+C)
- [x] Help display (?)
- [x] Panel navigation (Tab, 1-3, p)
- [x] Search activation (/)
- [x] Refresh all (Ctrl+R)

### 4.3 Notes Controller

- [x] Navigation (j/k, g/G)
- [ ] Half-page scroll (Ctrl+d/u)
- [x] Edit note (Enter, E)
- [x] New note (n)
- [x] Delete note (d)
- [x] Copy path (y)
- [x] Add/Remove tag (t/T)
- [x] Set/Remove parent (>/P)
- [x] Toggle bookmark (b)
- [x] Show info (s)

### 4.4 Tags Controller

- [x] Navigation (j/k)
- [x] Filter by tag (Enter)
- [x] Rename tag (r)
- [x] Delete tag (d)

### 4.5 Queries Controller

- [x] Navigation (j/k)
- [x] Run query (Enter)
- [ ] Edit query (e)
- [x] Delete query (d)
- [ ] New query (n)

### 4.6 Preview Controller

- [x] Scroll navigation (j/k, J/K cards, ]/[ headers)
- [x] Edit from preview (E)
- [x] Toggle frontmatter (f)
- [x] Focus note from card (Enter)
- [x] Back to previous (Esc)
- [x] Delete card (d)
- [x] Append #done (D)
- [x] Move card (m)
- [x] Merge notes (M)
- [x] Add/Remove tag (t/T)
- [x] Toggle inline tag (<c-t>)
- [x] Set/Remove parent (>/P)
- [x] Toggle bookmark (b)
- [x] Show info (s)
- [x] Open link (o)
- [x] Toggle todo (x)
- [x] View options (v)
- [x] Link highlight navigation (l/L)

### 4.7 Search Controller

- [x] Input handling
- [x] Execute search (Enter)
- [x] Cancel search (Esc)
- [x] Autocomplete (tags, dates, sort, filters, abbreviations)
- [ ] Save as query
- [ ] History navigation

## Phase 5: Helpers ✅

### 5.1 Core Helpers

- [x] Refresh functions in gui.go
- [x] Layout calculations in layout.go
- [x] ConfirmationHelper - dialog management

### 5.2 Domain Helpers

- [x] Search execution in handlers
- [x] EditorHelper - $EDITOR integration
- [x] Tag filter operations
- [x] Note operations

### 5.3 UI Helpers

- [x] StatusHelper - status bar updates
- [x] ClipboardHelper - copy operations (pbcopy/xclip/xsel)

## Phase 6: Presentation ✅

### 6.1 List Rendering

- [x] Notes list with date, title, tags
- [x] Tags list with counts
- [x] Queries list with query preview

### 6.2 Preview Rendering

- [x] Basic content display
- [x] Card list mode for search results
- [x] Frontmatter display toggle
- [x] Header highlighting
- [x] Syntax highlighting (tags, code blocks, markdown via chroma)

### 6.3 Status Bar

- [x] Context-sensitive hints
- [x] Error display (showError dialog)

## Phase 7: Dialogs ✅

### 7.1 Dialog Infrastructure

- [x] Modal overlay rendering
- [x] Input capture
- [x] Focus management
- [x] Keyboard handling

### 7.2 Dialog Types

- [x] Confirmation dialog (y/n)
- [x] Text input dialog (single line, with completion)
- [x] Help overlay
- [x] Menu dialog (move, merge, view options, info)
- [x] Command palette

### 7.3 Specific Dialogs

- [x] New note dialog (full capture popup with markdown editing)
- [x] Tag rename dialog
- [x] Delete confirmation
- [x] Set parent dialog
- [x] Add tag dialog
- [x] Toggle inline tag dialog
- [x] Merge notes dialog
- [x] Move card dialog
- [x] View options dialog
- [x] Info dialog (parent tree, children, TOC)
- [x] Snippet editor (two-field stacked)

## Phase 8: Polish

### 8.1 Theming

- [ ] Define color constants
- [ ] Apply theme to all views
- [ ] Support custom themes in config

### 8.2 Responsive Design

- [x] Handle narrow terminals (min size check)
- [ ] Accordion mode for small screens
- [ ] Graceful degradation

### 8.3 Performance

- [ ] Lazy loading for large vaults
- [ ] Viewport-only rendering
- [ ] Debounced search
- [ ] Async data refresh

### 8.4 Error Handling

- [ ] Graceful error display
- [ ] Recovery from ruin failures
- [x] Vault not found handling
- [ ] Permission errors

## Phase 9: Advanced Features

### 9.1 Bulk Operations

- [ ] Multi-select with Space
- [ ] Bulk delete
- [ ] Bulk tag add/remove

### 9.2 Search Enhancements

- [x] Tag autocomplete dropdown
- [x] Filter autocomplete (dates, sort, path, title, parent)
- [x] Abbreviation snippet expansion
- [ ] Search history
- [ ] Recent searches display

### 9.3 Editor Integration

- [x] Suspend/resume TUI for $EDITOR
- [x] Auto-refresh after edit

### 9.4 Clipboard

- [x] Copy note path (y in Notes panel)
- [ ] Copy note content

## Phase 10: Testing & Documentation

### 10.1 Testing

- [x] Unit tests for commands layer
- [x] Unit tests for models
- [x] Unit tests for gui state
- [x] Unit tests for render helpers
- [x] Automated smoke test via tmux (scripts/smoke-test.sh)
- [ ] Integration tests with mock ruin

### 10.2 Documentation

- [ ] README.md with installation/usage
- [x] CLAUDE.md for development guidance
- [ ] Config file documentation
- [ ] Keybindings reference

### 10.3 Release

- [ ] Version command
- [ ] Release binaries (goreleaser)
- [ ] Homebrew formula
- [ ] Installation script

## Milestones

### MVP (Phases 1-4) ✅
- Basic TUI with panel navigation
- Note list viewing
- Note editing in $EDITOR
- Basic search

### Beta (Phases 5-7) ✅
- Full keybinding support
- All dialogs functional
- Tag management
- Query management
- Completion system (search, capture, pick, snippets)
- Parent/bookmark management
- Preview editing actions

### 1.0 (Phases 8-10)
- Polished UI
- Performance optimized
- Fully documented
- Release artifacts

## Technical Decisions

### Why jesseduffield/gocui?

- Actively maintained fork used by lazygit
- Better mouse support than original
- More features (overlapping views, etc.)
- Battle-tested in production

### JSON Mode for Commands

All ruin commands use `--json` for:
- Reliable parsing
- Structured data
- Forward compatibility
- Error handling

### Configuration Location

- Config: `~/.config/lazyruin/config.yml`

### Vault Resolution

Priority order:
1. `--vault` CLI flag
2. Config file `vault_path`
3. `ruin config vault_path` output
4. Error if none found

### gocui Error Handling

The jesseduffield/gocui fork wraps errors using `github.com/go-errors/errors`, which breaks `errors.Is(err, gocui.ErrUnknownView)`. We use string comparison instead:

```go
if err != nil && err.Error() != "unknown view" {
    return err
}
```

**Follow-up**: Investigate why lazygit's `errors.Is` works but ours doesn't. May be a gocui version difference or go-errors version mismatch.

## Dependencies

- `github.com/jesseduffield/gocui` — TUI framework (lazygit's fork)
- `gopkg.in/yaml.v3` — Configuration file parsing
- `github.com/alecthomas/chroma/v2` — Syntax highlighting

## Success Criteria

1. **Functional**: All ruin CLI operations accessible via TUI
2. **Familiar**: lazygit users feel at home
3. **Fast**: Responsive with large vaults (1000+ notes)
4. **Stable**: No crashes, graceful error handling
5. **Configurable**: Keybindings and themes customizable
