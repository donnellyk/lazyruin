# LazyRuin Architecture

This document describes the architecture for LazyRuin, a TUI for the `ruin` notes CLI, heavily inspired by lazygit.

## Overview

LazyRuin provides a terminal-based visual interface for managing markdown notes with the ruin CLI. It follows lazygit's architectural patterns while adapting them for note management workflows.

## Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
│  app.go - Bootstrap, lifecycle, dependency injection            │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                          GUI Layer                               │
│  gui.go - gocui wrapper, layout management, event routing       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                       Controller Layer                           │
│  Controllers handle user input and orchestrate operations        │
│  - notes_controller.go      - tags_controller.go                │
│  - search_controller.go     - queries_controller.go             │
│  - preview_controller.go    - global_controller.go              │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                        Helper Layer                              │
│  Helpers encapsulate domain logic and complex operations         │
│  - refresh_helper.go        - search_helper.go                  │
│  - editor_helper.go         - confirmation_helper.go            │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                       Commands Layer                             │
│  Wraps ruin CLI execution with typed responses                   │
│  - ruin_commands.go         - search_commands.go                │
│  - tags_commands.go         - note_commands.go                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                        Models Layer                              │
│  Data structures representing domain entities                    │
│  - note.go, tag.go, query.go, search_result.go                  │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
lazyruin/
├── main.go                          # Entry point
├── pkg/
│   ├── app/
│   │   └── app.go                   # Application bootstrap, CLI flags
│   │
│   ├── commands/                    # ruin CLI wrappers (typed Go interfaces)
│   │   ├── ruin.go                  # Base command execution, JSON parsing
│   │   ├── executor.go              # Command executor interface (for testing)
│   │   ├── search.go                # Search operations
│   │   ├── note.go                  # Note mutations (set, append, merge)
│   │   ├── tags.go                  # Tag operations
│   │   ├── queries.go               # Saved query operations
│   │   ├── parent.go                # Parent/bookmark operations
│   │   └── pick.go                  # Pick (tag intersection) operations
│   │
│   ├── models/                      # Data structures
│   │   ├── note.go                  # Note with frontmatter fields
│   │   ├── tag.go                   # Tag with count
│   │   ├── query.go                 # Saved query
│   │   ├── parent.go                # Parent bookmark
│   │   └── pick.go                  # Pick result
│   │
│   ├── config/
│   │   └── config.go                # Configuration loading and saving
│   │
│   ├── gui/                         # All GUI code in a flat package
│   │   ├── gui.go                   # Main Gui struct, Run(), refresh
│   │   ├── state.go                 # GuiState and sub-state structs
│   │   ├── views.go                 # View name constants
│   │   ├── layout.go                # View creation and positioning
│   │   ├── commands.go              # Command table (keys, handlers, palette)
│   │   ├── keybindings.go           # Keybinding registration
│   │   ├── hints.go                 # Context-sensitive status bar hints
│   │   ├── statusbar.go             # Status bar rendering
│   │   ├── colors.go                # Color/style constants
│   │   │
│   │   ├── handlers.go              # Global + search handlers
│   │   ├── handlers_notes.go        # Notes panel handlers
│   │   ├── handlers_preview.go      # Preview panel + input popup handlers
│   │   ├── handlers_tags.go         # Tags panel handlers
│   │   ├── handlers_queries.go      # Queries panel handlers
│   │   ├── handlers_parents.go      # Parent bookmark handlers
│   │   ├── handlers_capture.go      # Capture (new note) handlers
│   │   ├── handlers_pick.go         # Pick popup handlers
│   │   ├── handlers_snippets.go     # Snippet editor handlers
│   │   ├── handlers_palette.go      # Command palette handlers
│   │   │
│   │   ├── completion.go            # Completion engine, state, accept logic
│   │   ├── completion_triggers.go   # Trigger definitions per context
│   │   ├── completion_candidates.go # Candidate provider functions
│   │   ├── completion_abbreviation.go # Abbreviation expansion
│   │   ├── completion_render.go     # Suggestion dropdown rendering
│   │   ├── editor_completion.go     # Configurable completion editor
│   │   ├── editor_capture.go        # Capture-specific editor (markdown)
│   │   ├── editor_palette.go        # Palette filter editor
│   │   │
│   │   ├── render.go                # List rendering (notes, tags, queries)
│   │   ├── render_preview.go        # Preview pane rendering (cards, content)
│   │   ├── highlight.go             # Link highlighting in preview
│   │   ├── markdown.go              # Markdown continuation helpers
│   │   └── dialogs.go               # Confirmation, menu, info dialogs
│   │
│   └── testutil/                    # Shared test helpers
│
├── scripts/
│   └── smoke-test.sh                # Automated TUI regression via tmux
│
├── docs/
│   ├── ARCHITECTURE.md              # This file
│   ├── KEYBINDINGS.md               # Keybinding reference
│   ├── CLI_CHANGES.md               # ruin CLI command reference
│   └── UI_MOCKUPS.md                # Visual mockups
│
└── go.mod
```

## Core Components

### 1. Application Bootstrap (`pkg/app/`)

`app.go` handles CLI flag parsing, vault resolution, config loading, and launches the GUI.

### 2. Commands Layer (`pkg/commands/`)

Wraps ruin CLI with typed Go interfaces:

```go
type RuinCommand struct {
    vaultPath string
    binPath   string
    executor  Executor
}

// Each domain has its own command struct with typed methods:
type SearchCommand struct { ruin *RuinCommand }
type NoteCommand struct { ruin *RuinCommand }
type TagsCommand struct { ruin *RuinCommand }
type ParentCommand struct { ruin *RuinCommand }
```

All commands use `--json` output for reliable parsing. The `Executor` interface enables test mocking.

### 3. Context System

Contexts are simple `ContextKey` constants (not interfaces). The `Gui` tracks `CurrentContext` and `PreviousContext` in `GuiState`, and `setContext()` handles focus switching:

```go
type ContextKey string

const (
    NotesContext        ContextKey = "notes"
    QueriesContext      ContextKey = "queries"
    TagsContext         ContextKey = "tags"
    PreviewContext      ContextKey = "preview"
    SearchContext       ContextKey = "search"
    CaptureContext      ContextKey = "capture"
    PickContext         ContextKey = "pick"
    PaletteContext      ContextKey = "palette"
    SearchFilterContext ContextKey = "searchFilter"
)
```

### 4. Command Table (`pkg/gui/commands.go`)

A single `commands()` method returns all user-facing actions. Each `Command` binds keys to handlers and drives both keybinding registration and the command palette:

```go
type Command struct {
    Name      string       // palette display name
    Category  string       // palette grouping
    Keys      []any        // gocui keys; nil = palette-only
    Views     []string     // scoped views; nil = global
    Handler   func(...)    // keybinding handler
    OnRun     func() error // palette-only runner
    Contexts  []ContextKey // palette context filter
}
```

### 5. Handlers

Handlers are methods on `*Gui` organized by domain in separate files (`handlers_preview.go`, `handlers_notes.go`, etc.). There is no controller or helper abstraction — handlers directly call command wrappers and update state.

### 6. State Management

State is organized hierarchically in `GuiState` (see `pkg/gui/state.go`):

```go
type GuiState struct {
    // Panel states
    Notes      *NotesState
    Tags       *TagsState
    Queries    *QueriesState
    Parents    *ParentsState
    Preview    *PreviewState

    // Context tracking
    CurrentContext  ContextKey
    PreviousContext ContextKey

    // Modal modes (each has its own completion state)
    SearchMode     bool
    CaptureMode    bool
    PickMode       bool
    InputPopupMode bool
    SnippetEditorMode bool
    // ... corresponding CompletionState pointers
}
```

### 7. Keybinding System

Bindings come from two sources:

1. **Command table** (`commands()`) — user-facing actions bound to views/keys, also drives the command palette
2. **Navigation bindings** (`keybindings.go`) — per-view j/k/arrows/mouse, popup Enter/Esc/Tab

Global bindings are suppressed during dialogs via `suppressDuringDialog()`.

## Data Flow

### Search Flow

```
User presses / → openSearch() → SearchContext active
         │
User types query (completion triggers fire via completionEditor)
         │
User presses Enter → executeSearch()
         │
         ▼
ruinCmd.Search.Search(query, opts)  →  exec: ruin search "<query>" --json
         │
         ▼
Parse JSON → []models.Note
         │
         ▼
state.Preview.Cards = notes
state.Preview.Mode = PreviewModeCardList
         │
         ▼
renderPreview() → setContext(PreviewContext)
```

### Edit Flow

```
User presses E → openInEditor(path)
         │
         ▼
gui.Suspend() → exec $EDITOR <path> → gui.Resume()
         │
         ▼
refreshTags() + refreshQueries() + reloadContent() + renderAll()
```

## Concurrency Model

1. **Main Thread**: All GUI updates and ruin CLI calls run on the main thread (synchronous)
2. **gui.Update()**: Used when a goroutine needs to schedule a GUI update on the main loop

## Configuration

Configuration is loaded from `~/.config/lazyruin/config.yml` by `pkg/config/config.go`. It stores abbreviation snippets and the vault path. The `--vault` CLI flag overrides the config file.
