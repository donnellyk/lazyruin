package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

// errEditConflict signals that the note's file was modified externally
// between OpenCaptureForEdit and the save attempt. The caller should keep
// the popup open so the user can recover their in-progress text.
var errEditConflict = errors.New("note was modified externally since edit began")

// CaptureHelper encapsulates the capture popup logic.
type CaptureHelper struct {
	c *HelperCommon
}

func NewCaptureHelper(c *HelperCommon) *CaptureHelper {
	return &CaptureHelper{c: c}
}

// resetCaptureState clears all mode-like fields on the capture context so
// Open* methods start from a known blank slate. Defensive against any path
// that pushes the capture context without going through CloseCapture first.
func resetCaptureState(ctx *context.CaptureContext) {
	ctx.Parent = nil
	ctx.Completion = types.NewCompletionState()
	ctx.LinkURL = ""
	ctx.LinkTitle = ""
	ctx.LinkTags = nil
	ctx.LinkExistingUUID = ""
	ctx.LinkParent = ""
	ctx.PrefillContent = ""
	ctx.EditingPath = ""
	ctx.EditingTitle = ""
	ctx.EditingMtime = time.Time{}
	ctx.ResolveState = context.ResolveIdle
	ctx.ResolveResult = nil
	ctx.ResolveDone = nil
}

// OpenCapture opens the capture popup, resetting state.
func (self *CaptureHelper) OpenCapture() error {
	return self.OpenCaptureWithParent("", "")
}

// OpenCaptureWithParent opens the capture popup with a pre-set parent.
// Pass empty uuid/title for no parent.
func (self *CaptureHelper) OpenCaptureWithParent(uuid, title string) error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	ctx := gui.Contexts().Capture
	resetCaptureState(ctx)
	if uuid != "" {
		ctx.Parent = &context.CaptureParentInfo{UUID: uuid, Title: title}
	}
	gui.PushContextByKey("capture")
	return nil
}

// OpenCaptureWithContent opens the capture popup with pre-filled text content.
func (self *CaptureHelper) OpenCaptureWithContent(text string) error {
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	ctx := gui.Contexts().Capture
	resetCaptureState(ctx)
	ctx.PrefillContent = text
	gui.PushContextByKey("capture")
	return nil
}

// OpenCaptureForEdit opens the capture popup populated with the note's
// current content (excluding frontmatter). The popup title becomes the
// note's title instead of "New Note". On Ctrl+S the file is rewritten and
// reindexed via `ruin doctor`; Esc closes without saving.
func (self *CaptureHelper) OpenCaptureForEdit(note *models.Note) error {
	if note == nil || note.Path == "" {
		return nil
	}
	gui := self.c.GuiCommon()
	if gui.PopupActive() {
		return nil
	}
	content, mtime, err := readNoteBodyAndMtime(note.Path)
	if err != nil {
		gui.ShowError(fmt.Errorf("failed to read note: %w", err))
		return nil
	}
	ctx := gui.Contexts().Capture
	resetCaptureState(ctx)
	ctx.PrefillContent = content
	ctx.EditingPath = note.Path
	ctx.EditingTitle = note.Title
	ctx.EditingMtime = mtime
	gui.PushContextByKey("capture")
	return nil
}

// saveEdit writes newContent back to path, preserving the original
// frontmatter block, then reindexes via `ruin doctor`.
//
// The written return distinguishes failure modes:
//   - written=false: no change on disk (read / conflict / write-to-temp
//     failure). Caller should keep the popup open so the user can recover.
//   - written=true: file was updated atomically. If err is non-nil, the
//     failure was in the Doctor reindex — the file is correct, but the
//     search index may be stale until the next refresh.
func (self *CaptureHelper) saveEdit(path string, expectedMtime time.Time, newContent string) (written bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	// External-edit conflict check: the mtime captured at popup open time
	// must still match. Any change (other process, another ruin run, sync
	// client) causes us to abort rather than silently clobber.
	if !expectedMtime.IsZero() {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return false, fmt.Errorf("stat %s: %w", path, statErr)
		}
		if !info.ModTime().Equal(expectedMtime) {
			return false, errEditConflict
		}
	}

	frontmatter := extractFrontmatter(data)
	body := newContent
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	merged := append([]byte{}, frontmatter...)
	merged = append(merged, []byte(body)...)

	// Atomic write: temp file in the same directory, then rename. Rename
	// within a filesystem is atomic on POSIX — either the old file or the
	// fully-written new file is visible, never a half-truncated version.
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return false, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(merged); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return false, fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return false, fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return false, fmt.Errorf("rename %s: %w", path, err)
	}

	// File is committed. Reindex errors are surfaced but don't undo the
	// write — the caller distinguishes via the written return.
	if err := self.c.RuinCmd().Doctor(path); err != nil {
		return true, fmt.Errorf("reindex: %w", err)
	}
	return true, nil
}

// readNoteBodyAndMtime reads a note file and returns the body (after the
// frontmatter, with leading blank lines trimmed) and the file's mtime for
// later conflict detection.
func readNoteBodyAndMtime(path string) (string, time.Time, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", time.Time{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", time.Time{}, err
	}
	frontmatter := extractFrontmatter(data)
	body := data[len(frontmatter):]
	// Strip leading blank lines whether LF- or CRLF-encoded.
	for len(body) > 0 {
		if body[0] == '\n' {
			body = body[1:]
		} else if len(body) >= 2 && body[0] == '\r' && body[1] == '\n' {
			body = body[2:]
		} else {
			break
		}
	}
	return string(body), info.ModTime(), nil
}

// readNoteBodyContent reads a note file and returns the body (everything
// after the closing `---` of the frontmatter, if any).
func readNoteBodyContent(path string) (string, error) {
	body, _, err := readNoteBodyAndMtime(path)
	return body, err
}

// extractFrontmatter returns the leading frontmatter block (including the
// trailing `---\n` and following blank line if present). Returns empty
// when the file has no frontmatter. Accepts both LF and CRLF line endings
// on the `---` delimiters.
func extractFrontmatter(data []byte) []byte {
	// Accept `---\n` or `---\r\n` as the opening delimiter.
	var openEnd int
	switch {
	case bytes.HasPrefix(data, []byte("---\n")):
		openEnd = 4
	case bytes.HasPrefix(data, []byte("---\r\n")):
		openEnd = 5
	default:
		return nil
	}
	// Find either `\n---\n` or `\r\n---\r\n` (or the mixed variants) as
	// the closing delimiter. Scanning for the plain `\n---` sequence and
	// then verifying the following bytes is sufficient.
	rest := data[openEnd:]
	for i := range len(rest) {
		if rest[i] != '\n' {
			continue
		}
		// Check `\n---` at position i
		if i+4 > len(rest) {
			break
		}
		if rest[i+1] == '-' && rest[i+2] == '-' && rest[i+3] == '-' {
			// Must be followed by \n or \r\n or end-of-file
			after := i + 4
			if after == len(rest) {
				return data[:openEnd+after]
			}
			if rest[after] == '\n' {
				end := openEnd + after + 1
				// Include a single trailing blank line if present (either `\n` or `\r\n`).
				if end < len(data) && data[end] == '\n' {
					end++
				} else if end+1 < len(data) && data[end] == '\r' && data[end+1] == '\n' {
					end += 2
				}
				return data[:end]
			}
			if rest[after] == '\r' && after+1 < len(rest) && rest[after+1] == '\n' {
				end := openEnd + after + 2
				if end < len(data) && data[end] == '\n' {
					end++
				} else if end+1 < len(data) && data[end] == '\r' && data[end+1] == '\n' {
					end += 2
				}
				return data[:end]
			}
		}
	}
	return nil
}

// SubmitCapture submits the capture content and closes the popup.
//
// Edit mode (ctx.EditingPath set): writes the edited content back to the
// note file and reindexes. Create mode: runs `ruin log`.
func (self *CaptureHelper) SubmitCapture(content string, quickCapture bool) error {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().Capture

	if ctx.EditingPath != "" {
		return self.submitEdit(ctx, content)
	}

	if content == "" {
		if quickCapture {
			return gocui.ErrQuit
		}
		return self.CloseCapture()
	}

	args := []string{"log", content}
	if ctx.Parent != nil {
		args = append(args, "--parent", ctx.Parent.UUID)
	}
	_, err := self.c.RuinCmd().Execute(args...)
	if err != nil {
		if quickCapture {
			return gocui.ErrQuit
		}
		return self.CloseCapture()
	}

	if quickCapture {
		return gocui.ErrQuit
	}

	self.CloseCapture()
	self.c.Helpers().Preview().ReloadActivePreview()
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

// submitEdit handles the Ctrl+S path for edit mode. Branches on saveEdit
// outcome to distinguish recoverable failures (keep popup open) from
// successful-write-but-reindex-failed (refresh UI, surface warning).
func (self *CaptureHelper) submitEdit(ctx *context.CaptureContext, content string) error {
	gui := self.c.GuiCommon()
	// Guard against accidental body wipe: empty/whitespace-only content in
	// edit mode is treated as cancel, not a save. Users who really want to
	// clear a note's body can delete it outright.
	if strings.TrimSpace(content) == "" {
		return self.CloseCapture()
	}
	written, err := self.saveEdit(ctx.EditingPath, ctx.EditingMtime, content)
	if err != nil && !written {
		// Write failed — file is untouched. Keep the popup open so the
		// user doesn't lose their edits; surface the specific error.
		if errors.Is(err, errEditConflict) {
			gui.ShowError(fmt.Errorf("%q was modified externally; your edits are preserved in the popup — close and reopen to see the new version", filepath.Base(ctx.EditingPath)))
		} else {
			gui.ShowError(err)
		}
		return nil
	}
	if err != nil {
		// File was written but Doctor reindex failed. Close + refresh so
		// the UI matches the on-disk state, but warn about index staleness.
		gui.ShowError(fmt.Errorf("saved but reindex failed (%w); run `ruin doctor` to refresh", err))
	}
	self.CloseCapture()
	self.c.Helpers().Preview().ReloadActivePreview()
	self.c.Helpers().Tags().RefreshTags(false)
	return nil
}

// CancelCapture cancels the capture, dismissing completion first if active.
func (self *CaptureHelper) CancelCapture(quickCapture bool) error {
	ctx := self.c.GuiCommon().Contexts().Capture
	if ctx.Completion.Active {
		ctx.Completion.Dismiss()
		return nil
	}
	if quickCapture {
		return gocui.ErrQuit
	}
	return self.CloseCapture()
}

// CloseCapture resets capture state and pops the context.
func (self *CaptureHelper) CloseCapture() error {
	gui := self.c.GuiCommon()
	resetCaptureState(gui.Contexts().Capture)
	gui.SetCursorEnabled(false)
	gui.PopContext()
	return nil
}
