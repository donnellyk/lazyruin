package helpers

import (
	"strings"

	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

// Navigator is the single entry point for changing the preview pane.
//
// All preview mutation routes through Navigator. NavigateTo produces a
// committed view (recorded in history); ShowHover produces a hover view
// (not recorded; title italicized). Back / Forward restore adjacent
// entries, re-running each context's query via RestoreSnapshot.
//
// Navigator applies the capture-on-departure rule: every method, before
// doing anything else, snapshots the current preview (if the current view
// is committed) into the current entry, then performs the operation. This
// centralises snapshot updates so toggle/scroll/filter handlers never need
// to know about the Navigator.
type Navigator struct {
	c                  *HelperCommon
	mgr                *context.NavigationManager
	currentIsCommitted bool
}

// NewNavigator creates a Navigator backed by the given NavigationManager.
func NewNavigator(c *HelperCommon, mgr *context.NavigationManager) *Navigator {
	return &Navigator{c: c, mgr: mgr}
}

// Manager returns the underlying NavigationManager.
func (n *Navigator) Manager() *context.NavigationManager { return n.mgr }

// IsCurrentCommitted reports whether the current preview was the result of
// a committed navigation (as opposed to a hover).
func (n *Navigator) IsCurrentCommitted() bool { return n.currentIsCommitted }

// NavigateTo performs a committed navigation. Capture-on-departure saves
// the outgoing state, load runs to mutate preview context state, focus
// shifts to the destination preview, and a new history entry is recorded.
func (n *Navigator) NavigateTo(destination types.ContextKey, title string, load func() error) error {
	n.captureOnDeparture()

	if load != nil {
		if err := load(); err != nil {
			return err
		}
	}

	n.c.GuiCommon().Contexts().ActivePreviewKey = destination
	if context.IsPreviewContextKey(destination) {
		n.c.GuiCommon().PushContextByKey(destination)
	}

	n.recordCurrent(destination, title)
	n.currentIsCommitted = true
	return nil
}

// ReplaceCurrent performs a committed navigation that replaces the current
// focused preview context without pushing a new context-stack entry. Used
// when the caller is already in a preview context and wants to swap the
// data (e.g. search submitted from card list).
func (n *Navigator) ReplaceCurrent(destination types.ContextKey, title string, load func() error) error {
	n.captureOnDeparture()

	if load != nil {
		if err := load(); err != nil {
			return err
		}
	}

	gui := n.c.GuiCommon()
	gui.Contexts().ActivePreviewKey = destination
	if context.IsPreviewContextKey(destination) && gui.CurrentContextKey() != destination {
		gui.ReplaceContextByKey(destination)
	}

	n.recordCurrent(destination, title)
	n.currentIsCommitted = true
	return nil
}

// ShowHover performs a hover navigation: capture-on-departure saves the
// outgoing committed state (if any), load runs to mutate preview context
// state, no history entry is recorded, and the title is decorated with
// italics to signal the view is not committed.
func (n *Navigator) ShowHover(destination types.ContextKey, title string, load func() error) error {
	n.captureOnDeparture()

	if load != nil {
		if err := load(); err != nil {
			return err
		}
	}

	n.c.GuiCommon().Contexts().ActivePreviewKey = destination
	n.decorateTitle(HoverTitle(title))
	n.currentIsCommitted = false
	return nil
}

// Back restores the previous committed view; no-op when there is no
// previous entry.
func (n *Navigator) Back() error {
	n.captureOnDeparture()

	evt, ok := n.mgr.Back()
	if !ok {
		return nil
	}
	if err := n.restore(evt); err != nil {
		// Partial restore: mark not-committed so the next capture-on-departure
		// does not overwrite the entry with half-applied state.
		n.currentIsCommitted = false
		return err
	}
	n.currentIsCommitted = true
	return nil
}

// Forward restores the next committed view; no-op when there is no next
// entry.
func (n *Navigator) Forward() error {
	n.captureOnDeparture()

	evt, ok := n.mgr.Forward()
	if !ok {
		return nil
	}
	if err := n.restore(evt); err != nil {
		n.currentIsCommitted = false
		return err
	}
	n.currentIsCommitted = true
	return nil
}

// CommitHover promotes the current hover view to a committed history entry.
// Strips the hover-title decoration and records a new entry at the current
// active preview context. No-op if the current view is already committed or
// no preview context is active.
func (n *Navigator) CommitHover() {
	if n.currentIsCommitted {
		return
	}
	ctx := n.c.GuiCommon().Contexts().ActivePreview()
	if ctx == nil {
		return
	}
	title := strings.TrimPrefix(ctx.Title(), "~ ")
	ctx.SetTitle(title)
	destination := n.c.GuiCommon().Contexts().ActivePreviewKey
	n.recordCurrent(destination, title)
	n.currentIsCommitted = true
}

// HoverTitle decorates a title to signal the view is a hover (not committed
// to history). gocui's view-title rendering does not interpret ANSI escapes,
// so a subtle typographic prefix is used instead of italics: a leading
// tilde-space stands in visually for the "italics" convention in the
// design spec.
func HoverTitle(title string) string {
	if title == "" {
		return ""
	}
	return "~ " + title
}

func (n *Navigator) captureOnDeparture() {
	if !n.currentIsCommitted {
		return
	}
	snap, ok := n.currentSnapshot()
	if !ok {
		return
	}
	n.mgr.UpdateCurrent(snap)
}

func (n *Navigator) recordCurrent(destination types.ContextKey, title string) {
	snap, _ := n.currentSnapshot()
	n.mgr.Record(context.NavigationEvent{
		ContextKey: destination,
		Title:      title,
		Snapshot:   snap,
	})
}

func (n *Navigator) currentSnapshot() (types.Snapshot, bool) {
	ctx := n.c.GuiCommon().Contexts().ActivePreview()
	if ctx == nil {
		return nil, false
	}
	s, ok := ctx.(types.Snapshotter)
	if !ok {
		return nil, false
	}
	return s.CaptureSnapshot(), true
}

func (n *Navigator) restore(evt context.NavigationEvent) error {
	gui := n.c.GuiCommon()
	contexts := gui.Contexts()

	target := evt.ContextKey
	if target == "" {
		target = "cardList"
	}
	contexts.ActivePreviewKey = target

	ctx := contexts.ActivePreview()
	if ctx != nil && evt.Title != "" {
		ctx.SetTitle(evt.Title)
	}

	if s, ok := ctx.(types.Snapshotter); ok && evt.Snapshot != nil {
		if err := s.RestoreSnapshot(evt.Snapshot); err != nil {
			return err
		}
	}

	if context.IsPreviewContextKey(target) {
		if gui.CurrentContextKey() != target {
			gui.ReplaceContextByKey(target)
		}
	}
	n.syncSidePaneCursor()
	gui.RenderPreview()
	return nil
}

// syncSidePaneCursor best-effort aligns the Notes panel cursor with the
// currently-selected card's note so the side-pane selection stays visually
// consistent with the preview after Back/Forward (Q2).
func (n *Navigator) syncSidePaneCursor() {
	card := n.c.Helpers().Preview().CurrentPreviewCard()
	if card == nil {
		return
	}
	notes := n.c.GuiCommon().Contexts().Notes
	if notes == nil {
		return
	}
	if notes.SelectByUUID(card.UUID) {
		n.c.GuiCommon().RenderNotes()
	}
}

func (n *Navigator) decorateTitle(title string) {
	ctx := n.c.GuiCommon().Contexts().ActivePreview()
	if ctx == nil {
		return
	}
	ctx.SetTitle(title)
}
