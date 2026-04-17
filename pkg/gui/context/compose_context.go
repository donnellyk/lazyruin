package context

import (
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// ComposeRequery re-runs the parent composition and returns the fresh note
// and source map.
type ComposeRequery func() (models.Note, []models.SourceMapEntry, error)

// ComposeState holds state specific to the compose preview mode.
type ComposeState struct {
	Note      models.Note             // single composed note
	SourceMap []models.SourceMapEntry // maps composed line ranges to source children
	Parent    models.ParentBookmark   // for reload after mutations
	Requery   ComposeRequery          // closure to re-compose the current parent
}

// ComposeContext owns the compose preview mode (parent composition via `ruin compose`).
type ComposeContext struct {
	BaseContext
	PreviewContextTrait
	*ComposeState
}

// NewComposeContext creates a ComposeContext.
func NewComposeContext() *ComposeContext {
	return &ComposeContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:      types.MAIN_CONTEXT,
			Key:       "compose",
			ViewName:  "preview",
			Focusable: true,
			Title:     "Compose",
		}),
		PreviewContextTrait: NewPreviewContextTrait(),
		ComposeState:        &ComposeState{},
	}
}

// IPreviewContext implementation (CardCount varies per context; the rest are
// provided by the embedded PreviewContextTrait).

func (self *ComposeContext) CardCount() int { return 1 }

// composeSnapshot carries view params (Parent + Requery closure) and view
// state for Compose restoration.
type composeSnapshot struct {
	Title           string
	Parent          models.ParentBookmark
	Requery         ComposeRequery
	FrozenNote      models.Note
	FrozenSourceMap []models.SourceMapEntry
	SelectedCardIdx int
	CursorLine      int
	ScrollOffset    int
	Display         PreviewDisplayState
}

func (self *ComposeContext) CaptureSnapshot() types.Snapshot {
	ns := self.NavState()
	return &composeSnapshot{
		Title:           self.Title(),
		Parent:          self.Parent,
		Requery:         self.Requery,
		FrozenNote:      self.Note,
		FrozenSourceMap: append([]models.SourceMapEntry(nil), self.SourceMap...),
		SelectedCardIdx: self.SelectedCardIdx,
		CursorLine:      ns.CursorLine,
		ScrollOffset:    ns.ScrollOffset,
		Display:         *self.DisplayState(),
	}
}

func (self *ComposeContext) RestoreSnapshot(s types.Snapshot) error {
	snap, ok := s.(*composeSnapshot)
	if !ok || snap == nil {
		return nil
	}
	self.SetTitle(snap.Title)
	self.Parent = snap.Parent
	self.Requery = snap.Requery
	*self.DisplayState() = snap.Display

	if snap.Requery != nil {
		note, sm, err := snap.Requery()
		if err == nil {
			self.Note = note
			self.SourceMap = sm
		} else {
			self.Note = snap.FrozenNote
			self.SourceMap = append([]models.SourceMapEntry(nil), snap.FrozenSourceMap...)
		}
	} else {
		self.Note = snap.FrozenNote
		self.SourceMap = append([]models.SourceMapEntry(nil), snap.FrozenSourceMap...)
	}

	self.SelectedCardIdx = snap.SelectedCardIdx
	ns := self.NavState()
	ns.CursorLine = snap.CursorLine
	ns.ScrollOffset = snap.ScrollOffset
	return nil
}

var _ types.Context = &ComposeContext{}
var _ IPreviewContext = &ComposeContext{}
var _ types.Snapshotter = &ComposeContext{}
