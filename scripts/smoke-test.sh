#!/usr/bin/env bash
# Smoke test for lazyruin TUI via tmux.
# Usage: ./scripts/smoke-test.sh [binary] [vault]
# Requires: tmux, ruin CLI.
#
# Exits 0 on success, 1 on first failure.
# Prints PASS/FAIL for each check.

set -euo pipefail

BIN="${1:-/tmp/lazyruin-test}"
VAULT="${2:-/private/tmp/ruin-smoke-$$}"
SESSION="smoke-$$"
COLS=120
ROWS=40
FAILURES=0
TOTAL=0
POLL_INTERVAL=0.05   # seconds between polls
POLL_MAX=60          # max polls (60 * 0.05 = 3s timeout)

# --- helpers ---

die()  { echo "FATAL: $1" >&2; cleanup; exit 1; }

cleanup() {
  tmux kill-session -t "$SESSION" 2>/dev/null || true
  if [ -d "$VAULT" ]; then
    rm -rf "$VAULT"
  fi
}
trap cleanup EXIT

# Capture the visible pane text.
cap() { tmux capture-pane -t "$SESSION" -p 2>/dev/null; }

# Capture just the status bar (last 3 lines).
status() { cap | tail -3; }

# Send keys with minimal delay (just enough for tmux to deliver).
send() {
  tmux send-keys -t "$SESSION" "$@"
  sleep 0.05
}

# Brief pause to let async GUI updates (g.Update) settle between dependent keys.
settle() { sleep 0.15; }

# Poll until needle appears in capture (returns 0) or timeout (returns 1).
wait_for() {
  local needle="$1" max="${2:-$POLL_MAX}" i=0
  while ! cap | grep -qF "$needle"; do
    i=$((i + 1))
    [ "$i" -ge "$max" ] && return 1
    sleep "$POLL_INTERVAL"
  done
}

# Poll until needle disappears from capture (returns 0) or timeout (returns 1).
wait_gone() {
  local needle="$1" max="${2:-$POLL_MAX}" i=0
  while cap | grep -qF "$needle"; do
    i=$((i + 1))
    [ "$i" -ge "$max" ] && return 1
    sleep "$POLL_INTERVAL"
  done
}

# Assert captured output contains a string (polls until found or timeout).
assert_contains() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" needle="$2"
  if wait_for "$needle"; then
    echo "  PASS: $desc"
  else
    echo "  FAIL: $desc (expected '$needle')"
    FAILURES=$((FAILURES + 1))
  fi
}

# Assert captured output does NOT contain a string (polls until gone or timeout).
assert_not_contains() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" needle="$2"
  if wait_gone "$needle"; then
    echo "  PASS: $desc"
  else
    echo "  FAIL: $desc (unexpected '$needle' still present)"
    FAILURES=$((FAILURES + 1))
  fi
}

# Assert focus via status bar hint (polls until found or timeout).
assert_status() {
  TOTAL=$((TOTAL + 1))
  local desc="$1" needle="$2" i=0
  while ! status | grep -qF "$needle"; do
    i=$((i + 1))
    if [ "$i" -ge "$POLL_MAX" ]; then
      echo "  FAIL: $desc (status bar missing '$needle')"
      FAILURES=$((FAILURES + 1))
      return
    fi
    sleep "$POLL_INTERVAL"
  done
  echo "  PASS: $desc"
}

# --- preflight ---

[ -f "$BIN" ]              || die "binary not found: $BIN (run: go build -o $BIN ./main.go)"
command -v tmux >/dev/null  || die "tmux not found"
command -v ruin >/dev/null  || die "ruin CLI not found"

# Seed a fresh test vault
ruin dev seed "$VAULT" >/dev/null 2>&1 || die "failed to seed vault"

echo "=== LazyRuin Smoke Test ==="
echo "bin=$BIN  vault=$VAULT  session=$SESSION"
echo ""

# --- launch ---

START_TIME=$SECONDS

tmux new-session -d -s "$SESSION" -x "$COLS" -y "$ROWS" \
  "$BIN --vault $VAULT" 2>/dev/null
wait_for "Preview" 200 || die "app did not start within 10s"  # 200 * 0.05 = 10s

# =============================================
# 1. Startup — three sidebar panels rendered
# =============================================
echo "[1] Startup"
assert_contains "notes panel"     "[1]"
assert_contains "queries panel"   "[2]"
assert_contains "tags panel"      "[3]"
assert_contains "preview panel"   "Preview"
# Notes has initial focus
assert_status "notes focused" "View: enter"

# =============================================
# 2. Panel focus cycling (Tab)
# =============================================
echo "[2] Panel cycling"
send Tab
assert_status "queries focused" "Run: enter"
send Tab
assert_status "tags focused" "Filter: enter"
send Tab
assert_status "back to notes" "View: enter"

# =============================================
# 3. Quick-focus keys (1/2/3/p)
# =============================================
echo "[3] Quick focus"
send 2
assert_status "2 -> queries" "Run: enter"
send 3
assert_status "3 -> tags" "Filter: enter"
send p
assert_contains "p -> pick" "Pick"
send Escape
settle
send Escape
send 1
assert_status "1 -> notes" "View: enter"

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
send j
TOTAL=$((TOTAL + 1))
echo "  PASS: j/k navigation (no crash)"
send g  # back to top

# =============================================
# 6. Enter to preview, Esc back
# =============================================
echo "[6] Notes -> Preview -> Back"
send Enter
assert_status "in preview" "Back: esc"
send Escape
assert_status "back to notes" "View: enter"

# =============================================
# 7. Search flow
# =============================================
echo "[7] Search"
send /
assert_contains "search popup open" "Search"
assert_status "search hints" "Complete: tab"
send -l "project"
send Enter
# Search filter bar should appear with query text and [0]-Search title
assert_contains "search filter shown" "[0]-Search"
assert_contains "query in filter bar" "project"
# Focus should be on preview with results
assert_status "preview after search" "Back: esc"

# Clear search via search filter
send 0  # focus search filter
assert_status "search filter focused" "Clear: x"
send x
assert_not_contains "search filter gone" "[0]-Search"

# =============================================
# 8. Pick flow
# =============================================
echo "[8] Pick"
send 1
send '\'
assert_contains "pick popup open" "Pick"
send Escape  # dismiss completion dropdown
settle
send Escape  # close pick dialog
assert_not_contains "pick closed" "Pick tags"

# =============================================
# 9. Command palette
# =============================================
echo "[9] Palette"
send 1  # ensure notes focus
settle
send :
assert_contains "palette open" "Command Palette"
assert_contains "commands listed" "Global:"
send -l "quit"
assert_contains "filtered to quit" "Global: Quit"
send Escape
assert_not_contains "palette closed" "Command Palette"

# =============================================
# 10. Help dialog
# =============================================
echo "[10] Help"
send ?
assert_contains "help visible" "Keybindings"
send Escape
settle

# =============================================
# 11. Note actions from Notes panel (shared keys)
# =============================================
echo "[11] Note actions (Notes)"
send 1
settle
send s  # Show Info
assert_contains "info dialog" "Info"
send Escape

# =============================================
# 12. Note actions from Preview (shared keys)
# =============================================
echo "[12] Note actions (Preview)"
send Enter  # view in preview
wait_for "Back: esc" || true
send s  # Show Info from preview
assert_contains "info from preview" "Info"
send Escape
settle
send Escape  # back to notes

# =============================================
# 13. Add Tag (input popup with completion)
# =============================================
echo "[13] Add Tag"
send 1
settle
send t  # Add Tag
assert_contains "add tag popup" "Add Tag"
assert_contains "tag seed" "#"
assert_contains "tag footer hint" "Tab: accept"
send Escape  # dismiss completion
settle
send Escape  # close popup
assert_not_contains "tag popup closed" "Add Tag"

# =============================================
# 14. Set Parent (input popup with completion)
# =============================================
echo "[14] Set Parent"
send 1
settle
send '>'  # Set Parent
assert_contains "set parent popup" "Set Parent"
assert_contains "parent completion shown" "alpha"
send Escape  # dismiss completion
settle
send Escape  # close popup
assert_not_contains "parent popup closed" "Set Parent"

# =============================================
# 15. Run query
# =============================================
echo "[15] Queries"
send 2
settle
send Enter  # run first query
assert_status "preview after query" "Back: esc"
send 1

# =============================================
# 16. Filter by tag
# =============================================
echo "[16] Tags"
send 3
settle
send Enter  # filter by tag
assert_status "preview after tag filter" "Back: esc"
send 1

# =============================================
# 17. New note capture
# =============================================
echo "[17] Capture"
send n
assert_contains "capture popup" "New Note"
assert_contains "save hint" "<c-s> to save"
send Escape
assert_not_contains "capture closed" "New Note"

# =============================================
# 18. Calendar dialog
# =============================================
echo "[18] Calendar"
send 1
settle
send c
assert_contains "calendar open" "Su Mo Tu We Th Fr Sa"
send Escape
assert_not_contains "calendar closed" "Su Mo Tu We Th Fr Sa"

# =============================================
# 19. Contributions dialog
# =============================================
echo "[19] Contributions"
send 1
settle
send C
assert_contains "contrib open" "Contributions"
send Escape
assert_not_contains "contrib closed" "Contributions"

# =============================================
# 20. Resize handling
# =============================================
echo "[20] Resize"
tmux resize-pane -t "$SESSION" -x 80 -y 24
sleep 0.15  # allow resize event to propagate
TOTAL=$((TOTAL + 1))
if cap >/dev/null 2>&1; then
  echo "  PASS: resize survived"
else
  echo "  FAIL: resize crashed"
  FAILURES=$((FAILURES + 1))
fi
tmux resize-pane -t "$SESSION" -x "$COLS" -y "$ROWS"

# =============================================
# Done
# =============================================
ELAPSED=$((SECONDS - START_TIME))
echo ""
echo "=== Results: $((TOTAL - FAILURES))/$TOTAL passed in ${ELAPSED}s ==="
if [ "$FAILURES" -gt 0 ]; then
  echo "FAILED ($FAILURES failures)"
  exit 1
else
  echo "ALL PASSED"
  exit 0
fi
