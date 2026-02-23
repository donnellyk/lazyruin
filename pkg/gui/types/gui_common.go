package types

import (
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
	"github.com/muesli/reflow/wordwrap"
)

// SourceLine is the atomic unit of rendered preview content. It pairs the
// displayed text with the source file location it came from, enforcing a
// 1:1 correspondence by construction.
type SourceLine struct {
	Text    string // rendered line for display
	UUID    string // note UUID (empty for non-content lines like separators)
	LineNum int    // 1-indexed content line (matches ruin CLI --line), 0 for non-content
	Path    string // source file path
}

// PreviewContentWidth returns the usable content width for the preview view,
// accounting for the 1-character padding on each side.
func PreviewContentWidth(v *gocui.View) int {
	width, _ := v.InnerSize()
	if width < 10 {
		width = 40
	}
	return max(width-2, 10)
}

// PickMatchVisualLines returns the number of visual lines a pick match
// occupies when rendered at the given content width. Used by both the
// renderer and the resolver to ensure identical line counts.
func PickMatchVisualLines(match models.PickMatch, contentWidth int) int {
	lineNum := fmt.Sprintf("%02d", match.Line)
	prefix := fmt.Sprintf("  L%s: ", lineNum)
	prefixLen := len(prefix)
	wrapped := wordwrap.String(match.Content, contentWidth-prefixLen)
	return len(strings.Split(strings.TrimRight(wrapped, "\n"), "\n"))
}

// IGuiCommon is the authoritative interface for GUI operations.
// Controllers, helpers, and handler code interact with the GUI through
// this interface instead of depending on the full Gui struct.
//
// The helpers package extends this with Contexts() *context.ContextTree
// (which can't live here due to the types↔context import cycle).
type IGuiCommon interface {
	// Rendering
	Update(func() error)
	RenderNotes()
	RenderTags()
	RenderQueries()
	RenderPreview()
	RenderAll()
	UpdateNotesTab()
	UpdateTagsTab()
	UpdateQueriesTab()
	UpdateStatusBar()

	// Navigation
	CurrentContext() Context
	CurrentContextKey() ContextKey
	PushContext(ctx Context, opts OnFocusOpts)
	PushContextByKey(key ContextKey)
	PopContext()
	ReplaceContext(ctx Context)
	ReplaceContextByKey(key ContextKey)
	ContextByKey(key ContextKey) Context
	PopupActive() bool
	SearchQueryActive() bool

	// Dialogs
	ShowConfirm(title, message string, onConfirm func() error)
	ShowInput(title, message string, onConfirm func(string) error)
	ShowError(err error)
	ShowMenuDialog(title string, items []MenuItem)

	// Search
	SetCursorEnabled(enabled bool)

	// Editor
	Suspend() error
	Resume() error

	// View access
	GetView(name string) *gocui.View
	DeleteView(name string)

	// Preview rendering
	BuildCardContent(note models.Note, width int) []SourceLine
	RenderPickDialog()

	// Context state
	PreviousContextKey() ContextKey
}
