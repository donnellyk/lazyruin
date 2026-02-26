# Test Coverage Analysis

## Current State

The codebase has **25 test files** covering **~127 non-test source files** across three testing layers:

| Layer | Mechanism | Count |
|-------|-----------|-------|
| Unit tests | `go test` with mocks | 25 `_test.go` files |
| Integration tests | `go test -tags=integration` against real `ruin` CLI | 3 files (build-tagged) |
| E2E smoke tests | tmux-based TUI assertions | 2 shell scripts (~150 assertions) |

### Per-Package Coverage

| Package | Source Files | Test Files | Coverage |
|---------|-------------|------------|----------|
| `pkg/commands` | 8 | 9 | Strong — both unit (mock) and integration tests |
| `pkg/gui` (top-level) | ~37 | 7 | Moderate — headless GUI tests cover major workflows |
| `pkg/gui/context` | ~15 contexts + ~9 utilities | 3 | Low — only list primitives and datepreview tested |
| `pkg/gui/controllers` | ~12 controllers + ~7 utilities | 1 | Very low — only `ListControllerTrait` tested |
| `pkg/gui/helpers` | ~27 helpers | 4 | Low — only completion, datepreview, pick, refresh |
| `pkg/gui/types` | 10 | 0 | N/A (pure interfaces) |
| `pkg/models` | 5 | 1 | Low — only `Note` methods partially tested |
| `pkg/config` | 1 | 0 | None |
| `pkg/app` | 1 | 0 | None |

---

## Recommended Improvements

### Priority 1: High-Value, Low-Effort Unit Tests

These are pure functions or simple structs with no GUI dependencies — easy to test in isolation.

#### 1.1 `pkg/config` — Config loading, migration, and abbreviations

`config.go` has non-trivial logic that is completely untested:

- **`VaultAbbreviations()`** — vault-scoped lookup with legacy fallback
- **`SetVaultAbbreviation()` / `DeleteVaultAbbreviation()`** — map mutation with cleanup
- **`Load()`** — YAML parsing with automatic migration from flat to nested abbreviation format
- **`getConfigPath()`** — XDG_CONFIG_HOME resolution

Suggested tests:
- Table-driven tests for `VaultAbbreviations` (vault-specific, legacy fallback, nil map, empty map)
- Round-trip test: `SetVaultAbbreviation` then `VaultAbbreviations` returns it
- `DeleteVaultAbbreviation` removes entry and cleans up empty vault maps
- `Load` with old flat format triggers migration
- `Load` with new nested format parses correctly
- `Load` with missing file returns defaults
- `getConfigPath` respects `XDG_CONFIG_HOME` vs default `~/.config`

#### 1.2 `pkg/models` — Untested model methods

Only `ShortDate()` and `TagsString()` have tests. Missing:

- **`FirstLine()`** — empty content, single line, multiline with leading blanks
- **`JoinDot()`** — empty parts, single part, multiple parts, mixed empty/non-empty
- **`GlobalTagsString()`** — tags with/without `#` prefix, empty tags

These are trivial table-driven tests.

#### 1.3 `pkg/gui/helpers/search_helper.go` — `extractSort()`

The `extractSort()` function is a pure string parser that is not tested anywhere:

- Query with `sort:created` token
- Query with no sort token
- Query with sort token at beginning/middle/end
- Multiple sort tokens (last wins? all extracted?)

#### 1.4 `pkg/commands/note.go` — NoteCommand argument construction

Similar to the existing `pick_unit_test.go` pattern, `NoteCommand` has 14 methods that construct CLI arguments. None are tested. Key methods to cover:

- `Append()` — conditional `--line` and `--suffix` flags
- `Merge()` — conditional `--delete-source` and `--strip-title` flags, plus JSON response parsing
- `ToggleTodo()` — `--toggle-todo --line N --sink` argument ordering
- `AddTagToLine()` / `RemoveTagFromLine()` — `--line` interaction with `--add-tag`/`--remove-tag`

Could reuse the `argCapture` spy pattern from `pick_unit_test.go`.

---

### Priority 2: Medium-Effort Tests for Critical Business Logic

#### 2.1 `pkg/gui/helpers/` — Untested helpers with domain logic

Several helpers contain testable logic that could be extracted and unit-tested:

| Helper | Testable Logic |
|--------|---------------|
| `search_helper.go` | `extractSort()` (pure function — see 1.3) |
| `calendar_helper.go` | Date calculation, week boundaries, month navigation |
| `snippet_helper.go` | Abbreviation expansion, cursor positioning |
| `preview_line_ops_helper.go` | Line-level operations (toggle todo, tag manipulation) |
| `preview_links_helper.go` | Link extraction and navigation logic |
| `preview_info_helper.go` | Note info formatting |
| `date_candidates.go` | `AmbientDateCandidates()` — relative date string generation |

#### 2.2 `pkg/gui/context/` — Context initialization and state

Most context types follow a pattern of initialization + state management. Tests should verify:

- Default state after construction (current tab, selected index, empty lists)
- Tab cycling behavior (e.g., `NotesContext` has Today/All/Search tabs)
- `GetViewNames()` / `GetPrimaryViewName()` return correct view identifiers
- List adapter delegates properly to underlying data

#### 2.3 `pkg/gui/` — Completion subsystem gaps

The completion system (`completion_test.go`) is well-tested for cursor/token mechanics, but the following are untested:

- **`completion_abbreviation.go`** — snippet expansion trigger and replacement
- **`completion_candidates.go`** — candidate generation for different trigger types
- **`completion_triggers.go`** — trigger registration and matching logic

---

### Priority 3: Structural / Architectural Test Improvements

#### 3.1 Error path coverage

The codebase has extensive error handling (CLI failures, file-not-found, parse errors), but tests mostly cover the happy path. Add:

- Mock executor returning errors — verify error propagation and user-facing messages
- `CheckVault()` with unreachable vault
- JSON parse failures from malformed CLI output
- File I/O errors in `loadNoteContent`

#### 3.2 `pkg/app` — Vault resolution logic

`resolveVaultPath()` in `app.go` has a 4-step priority chain (CLI flag → config → env → ruin CLI) that is untested. This could be tested by:

- Extracting `resolveVaultPath` to accept its dependencies as parameters
- Testing each priority level and the fallback chain

#### 3.3 Mouse interaction tests

The codebase supports mouse navigation (`list_mouse_trait.go`, `GetMouseKeybindings`, `GetOnClick`), but no tests exercise mouse handlers. The headless GUI infrastructure (`testgui_test.go`) could be extended to simulate mouse events.

#### 3.4 Thread safety

Several GUI operations use `gui.g.Update()` for goroutine-safe updates. Consider adding tests that exercise concurrent access patterns to verify there are no race conditions. Run with `go test -race`.

---

### Priority 4: Testing Infrastructure Improvements

#### 4.1 Make `MockExecutor` support `NoteCommand` operations

The current `MockExecutor` handles `today`, `search`, `tags`, `query`, `parent`, and `compose` but does not handle `note` subcommands (delete, set, append, merge). Extending it would enable unit testing of `NoteCommand` and the helpers that call it (e.g., `NoteActionsHelper`).

#### 4.2 Add fuzz tests for user-input parsing

Functions that parse arbitrary user input are good fuzz targets:

- `extractSort()` — search query parsing
- `ParsePickQuery()` — pick dialog query parsing
- Completion token extraction
- Markdown continuation detection

#### 4.3 Add benchmark tests for hot paths

No benchmarks exist. Candidates:

- `loadNoteContent()` with large files
- Completion candidate filtering with many notes
- List rendering with hundreds of items
- Markdown continuation matching

---

## Summary Matrix

| Area | Current | Recommended | Effort |
|------|---------|-------------|--------|
| Config (load, migrate, abbreviations) | 0% | Unit tests | Low |
| Models (FirstLine, JoinDot, GlobalTagsString) | ~30% | Table-driven unit tests | Low |
| NoteCommand (14 methods) | 0% | Arg-capture unit tests | Low |
| `extractSort` | 0% | Table-driven unit tests | Low |
| Calendar/snippet/preview helpers | 0% | Extract + unit test pure logic | Medium |
| Context initialization/state | ~20% | State verification tests | Medium |
| Completion subsystem (abbreviation, candidates, triggers) | ~50% | Unit tests | Medium |
| Error paths across all layers | ~10% | Mock error injection | Medium |
| App vault resolution | 0% | Dependency injection + test | Medium |
| Mouse interactions | 0% | Extend headless GUI tests | High |
| Thread safety (race detection) | 0% | `-race` flag + concurrent tests | High |
| Fuzz tests (input parsing) | 0% | `testing.F` fuzz targets | Medium |
| Benchmarks | 0% | `testing.B` benchmarks | Low |
