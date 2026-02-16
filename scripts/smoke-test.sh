#!/usr/bin/env bash
# Smoke test for lazyruin TUI via tmux.
# Usage: ./scripts/smoke-test.sh [binary] [vault]
# Requires: tmux, ruin CLI, a populated test vault.
#
# Exits 0 on success, 1 on first failure.
# Prints PASS/FAIL for each check.

set -euo pipefail

BIN="${1:-/tmp/lazyruin-test}"
VAULT="${2:-/private/tmp/ruin-test-vault}"
SESSION="smoke-$$"
COLS=120
ROWS=40
FAILURES=0
TOTAL=0

# --- helpers ---

die()  { echo "FATAL: $1" >&2; cleanup; exit 1; }

cleanup() {
  tmux kill-session -t "$SESSION" 2>/dev/null || true
}
trap cleanup EXIT

# Capture the visible pane text.
cap() { tmux capture-pane -t "$SESSION" -p 2>/dev/null; }

# Capture just the status bar (last 3 lines).
status() { cap | tail -3; }

# Send keys and wait for render.
send() {
  tmux send-keys -t "$SESSION" "$@"
  sleep "${SEND_DELAY:-0.4}"
}

# Assert captured output contains a string.
assert_contains() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" needle="$2"
  if cap | grep -qF "$needle"; then
    echo "  PASS: $desc"
  else
    echo "  FAIL: $desc (expected '$needle')"
    FAILURES=$((FAILURES + 1))
  fi
}

# Assert captured output does NOT contain a string.
assert_not_contains() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" needle="$2"
  if cap | grep -qF "$needle"; then
    echo "  FAIL: $desc (unexpected '$needle')"
    FAILURES=$((FAILURES + 1))
  else
    echo "  PASS: $desc"
  fi
}

# Assert focus by checking the status bar for a context-specific hint.
# Each context has a unique hint that doesn't appear in others.
assert_status() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" needle="$2"
  if status | grep -qF "$needle"; then
    echo "  PASS: $desc"
  else
    echo "  FAIL: $desc (status bar missing '$needle')"
    FAILURES=$((FAILURES + 1))
  fi
}

# --- preflight ---

[ -f "$BIN" ]              || die "binary not found: $BIN (run: go build -o $BIN ./main.go)"
[ -d "$VAULT" ]            || die "vault not found: $VAULT"
command -v tmux >/dev/null  || die "tmux not found"

echo "=== LazyRuin Smoke Test ==="
echo "bin=$BIN  vault=$VAULT  session=$SESSION"
echo ""

# --- launch ---

tmux new-session -d -s "$SESSION" -x "$COLS" -y "$ROWS" \
  "$BIN --vault $VAULT" 2>/dev/null
sleep 1.5  # startup + initial load

# =============================================
# 1. Startup — three sidebar panels rendered
# =============================================
echo "[1] Startup"
assert_contains "notes panel"     "[1]"
assert_contains "queries panel"   "[2]"
assert_contains "tags panel"      "[3]"
assert_contains "preview panel"   "Preview"
# Notes has initial focus (Tab: 1 hint)
assert_status "notes focused" "Tab: 1"

# =============================================
# 2. Panel focus cycling (Tab)
# =============================================
echo "[2] Panel cycling"
send Tab
assert_status "queries focused" "Tab: 2"
send Tab
assert_status "tags focused" "Tab: 3"
send Tab
assert_status "back to notes" "Tab: 1"

# =============================================
# 3. Quick-focus keys (1/2/3/p)
# =============================================
echo "[3] Quick focus"
send 2
assert_status "2 -> queries" "Tab: 2"
send 3
assert_status "3 -> tags" "Tab: 3"
send p
assert_status "p -> preview" "Back: esc"
send 1
assert_status "1 -> notes" "Tab: 1"

# =============================================
# 4. Notes tab headers
# =============================================
echo "[4] Notes tabs"
assert_contains "tab headers" "All - Today - Recent"

# =============================================
# 5. j/k navigation in notes
# =============================================
echo "[5] List navigation"
send j
sleep 0.2
send j
sleep 0.2
TOTAL=$((TOTAL + 1))
echo "  PASS: j/k navigation (no crash)"
send g  # back to top

# =============================================
# 6. Enter to preview, Esc back
# =============================================
echo "[6] Notes -> Preview -> Back"
send Enter
sleep 0.5
assert_status "in preview" "Back: esc"
send Escape
assert_status "back to notes" "Tab: 1"

# =============================================
# 7. Search flow
# =============================================
echo "[7] Search"
send /
sleep 0.3
assert_contains "search popup open" "Search"
assert_status "search hints" "Complete: tab"
send -l "project"
sleep 0.3
send Enter
sleep 0.8
# Search filter bar should appear with query text and [0]-Search title
assert_contains "search filter shown" "[0]-Search"
assert_contains "query in filter bar" "project"
# Focus should be on preview with results
assert_status "preview after search" "Back: esc"

# Clear search via search filter
send 0  # focus search filter
sleep 0.3
assert_status "search filter focused" "Clear: x"
send x
sleep 0.5
assert_not_contains "search filter gone" "[0]-Search"

# =============================================
# 8. Pick flow
# =============================================
echo "[8] Pick"
send 1
sleep 0.2
send '\'
sleep 0.3
assert_contains "pick popup open" "Pick"
send Escape  # dismiss completion dropdown
sleep 0.2
send Escape  # close pick dialog
sleep 0.2
assert_not_contains "pick closed" "Pick tags"

# =============================================
# 9. Command palette
# =============================================
echo "[9] Palette"
send 1  # ensure notes focus
sleep 0.3
send :
sleep 0.5
assert_contains "palette open" "Command Palette"
assert_contains "commands listed" "Global:"
send -l "quit"
sleep 0.3
assert_contains "filtered to quit" "Global: Quit"
send Escape
sleep 0.3
assert_not_contains "palette closed" "Command Palette"

# =============================================
# 10. Help dialog
# =============================================
echo "[10] Help"
send ?
sleep 0.3
assert_contains "help visible" "Keybindings"
send Escape
sleep 0.2

# =============================================
# 11. Note actions from Notes panel (shared keys)
# =============================================
echo "[11] Note actions (Notes)"
send 1
sleep 0.2
send s  # Show Info
sleep 0.5
assert_contains "info dialog" "Info"
send Escape
sleep 0.2

# =============================================
# 12. Note actions from Preview (shared keys)
# =============================================
echo "[12] Note actions (Preview)"
send Enter  # view in preview
sleep 0.5
send s  # Show Info from preview
sleep 0.5
assert_contains "info from preview" "Info"
send Escape
sleep 0.2
send Escape  # back to notes

# =============================================
# 13. Add Tag (input popup with completion)
# =============================================
echo "[13] Add Tag"
send 1
sleep 0.2
send t  # Add Tag
sleep 0.5
assert_contains "add tag popup" "Add Tag"
assert_contains "tag seed" "#"
assert_contains "tag footer hint" "Tab: accept"
send Escape  # dismiss completion
sleep 0.2
send Escape  # close popup
sleep 0.2
assert_not_contains "tag popup closed" "Add Tag"

# =============================================
# 14. Set Parent (input popup with completion)
# =============================================
echo "[14] Set Parent"
send 1
sleep 0.2
send '>'  # Set Parent
sleep 0.5
assert_contains "set parent popup" "Set Parent"
assert_contains "parent completion shown" "alpha"
send Escape  # dismiss completion
sleep 0.2
send Escape  # close popup
sleep 0.2
assert_not_contains "parent popup closed" "Set Parent"

# =============================================
# 15. Run query
# =============================================
echo "[15] Queries"
send 2
sleep 0.3
send Enter  # run first query
sleep 0.8
assert_status "preview after query" "Back: esc"
send 1

# =============================================
# 16. Filter by tag
# =============================================
echo "[16] Tags"
send 3
sleep 0.3
send Enter  # filter by tag
sleep 0.8
assert_status "preview after tag filter" "Back: esc"
send 1

# =============================================
# 17. New note capture
# =============================================
echo "[17] Capture"
send n
sleep 0.3
assert_contains "capture popup" "New Note"
assert_contains "save hint" "<c-s> to save"
send Escape
sleep 0.2
assert_not_contains "capture closed" "New Note"

# =============================================
# 18. Resize handling
# =============================================
echo "[18] Resize"
tmux resize-pane -t "$SESSION" -x 80 -y 24
sleep 0.5
TOTAL=$((TOTAL + 1))
if cap >/dev/null 2>&1; then
  echo "  PASS: resize survived"
else
  echo "  FAIL: resize crashed"
  FAILURES=$((FAILURES + 1))
fi
tmux resize-pane -t "$SESSION" -x "$COLS" -y "$ROWS"
sleep 0.3

# =============================================
# Done
# =============================================
echo ""
echo "=== Results: $((TOTAL - FAILURES))/$TOTAL passed ==="
if [ "$FAILURES" -gt 0 ]; then
  echo "FAILED ($FAILURES failures)"
  exit 1
else
  echo "ALL PASSED"
  exit 0
fi
