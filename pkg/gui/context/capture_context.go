package context

import "github.com/donnellyk/lazyruin/pkg/gui/types"

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
	LinkExistingUUID string // non-empty when re-resolving an existing link note
	LinkParent       string // parent UUID to preserve on re-resolve
	PrefillContent   string // pre-filled text (e.g. from inbox promote)
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
