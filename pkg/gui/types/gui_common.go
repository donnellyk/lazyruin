package types

import (
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
)

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
	BuildCardContent(note models.Note, width int) []string

	// Context state
	PreviousContextKey() ContextKey
}
