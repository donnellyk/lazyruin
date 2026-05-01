package context

import (
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

// CaptureParentInfo tracks the parent selected via > completion in the capture dialog.
type CaptureParentInfo struct {
	UUID  string
	Title string // display title for footer (e.g. "Parent / Child")
}

type LinkResolveState int

const (
	ResolveIdle LinkResolveState = iota
	ResolveInFlight
	ResolveComplete
	ResolveFailed
)

type LinkResolveResult struct {
	Title   string
	Summary string
}

// CaptureContext owns the capture popup panel.
type CaptureContext struct {
	BaseContext
	Parent           *CaptureParentInfo
	Completion       *types.CompletionState
	LinkURL          string
	LinkTitle        string
	LinkTags         []string
	LinkExistingUUID string    // non-empty when re-resolving an existing link note
	LinkParent       string    // parent UUID to preserve on re-resolve
	PrefillContent   string    // pre-filled text (e.g. from scratchpad promote)
	CursorLine       int       // 0-indexed line in PrefillContent to position cursor on; 0 = end (default)
	EditingPath      string    // non-empty when editing an existing note; overrides save flow
	EditingUUID      string    // UUID of the note being edited; used to apply frontmatter mutations (parent, etc.) on save
	EditingTitle     string    // title of the note being edited, shown as popup title
	EditingMtime     time.Time // modification time of the file when the popup opened; used to detect external edits
	ResolveState     LinkResolveState
	ResolveResult    *LinkResolveResult
	ResolveDone      chan struct{}
}

// NewCaptureContext creates a CaptureContext.
func NewCaptureContext() *CaptureContext {
	return &CaptureContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.PERSISTENT_POPUP,
			Key:       "capture",
			ViewName:  "capture",
			Focusable: true,
			Title:     "Capture",
		}),
		Completion: types.NewCompletionState(),
	}
}

var _ types.Context = &CaptureContext{}
