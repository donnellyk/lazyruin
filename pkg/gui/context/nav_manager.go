package context

import (
	"time"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

// NavigationEvent is a single entry in the preview navigation history.
type NavigationEvent struct {
	ContextKey types.ContextKey
	Title      string
	Snapshot   types.Snapshot
	Timestamp  time.Time
}

// DefaultNavHistoryCap is the default cap on the number of history entries.
const DefaultNavHistoryCap = 50

// NavigationManager is the preview-scoped history stack. It owns the ordered
// list of NavigationEvents and the current index. It does not know anything
// about snapshot internals — it treats the Snapshot field as opaque.
type NavigationManager struct {
	entries []NavigationEvent
	index   int
	cap     int
}

// NewNavigationManager creates an empty NavigationManager with the default cap.
func NewNavigationManager() *NavigationManager {
	return &NavigationManager{index: -1, cap: DefaultNavHistoryCap}
}

// Record appends a new entry, truncating any forward entries and applying
// the ring-buffer cap.
func (m *NavigationManager) Record(evt NavigationEvent) {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now()
	}
	if m.index >= 0 && m.index < len(m.entries)-1 {
		m.entries = m.entries[:m.index+1]
	}
	m.entries = append(m.entries, evt)
	m.index = len(m.entries) - 1
	if m.cap > 0 && len(m.entries) > m.cap {
		drop := len(m.entries) - m.cap
		m.entries = m.entries[drop:]
		m.index -= drop
	}
}

// UpdateCurrent overwrites the current entry's snapshot in place.
// Called by capture-on-departure inside Navigator.
func (m *NavigationManager) UpdateCurrent(s types.Snapshot) {
	if m.index < 0 || m.index >= len(m.entries) {
		return
	}
	m.entries[m.index].Snapshot = s
}

// UpdateCurrentTitle overwrites the current entry's title in place.
func (m *NavigationManager) UpdateCurrentTitle(title string) {
	if m.index < 0 || m.index >= len(m.entries) {
		return
	}
	m.entries[m.index].Title = title
}

// Back decrements the index and returns the new current entry, or
// (zero, false) if there is no previous entry.
func (m *NavigationManager) Back() (NavigationEvent, bool) {
	if m.index <= 0 {
		return NavigationEvent{}, false
	}
	m.index--
	return m.entries[m.index], true
}

// Forward increments the index and returns the new current entry, or
// (zero, false) if there is no next entry.
func (m *NavigationManager) Forward() (NavigationEvent, bool) {
	if m.index < 0 || m.index >= len(m.entries)-1 {
		return NavigationEvent{}, false
	}
	m.index++
	return m.entries[m.index], true
}

// Current returns the current entry, or (zero, false) if empty.
func (m *NavigationManager) Current() (NavigationEvent, bool) {
	if m.index < 0 || m.index >= len(m.entries) {
		return NavigationEvent{}, false
	}
	return m.entries[m.index], true
}

// JumpTo sets the index to idx and returns the new current entry, or
// (zero, false) if idx is out of range. Used by the navigation-history
// menu to let the user jump to an arbitrary past entry.
func (m *NavigationManager) JumpTo(idx int) (NavigationEvent, bool) {
	if idx < 0 || idx >= len(m.entries) {
		return NavigationEvent{}, false
	}
	m.index = idx
	return m.entries[idx], true
}

// Clear empties the history.
func (m *NavigationManager) Clear() {
	m.entries = nil
	m.index = -1
}

// Len returns the number of entries currently in history.
func (m *NavigationManager) Len() int { return len(m.entries) }

// Index returns the current position (-1 when empty).
func (m *NavigationManager) Index() int { return m.index }

// Entries returns a read-only view of the entries. Callers must not mutate.
func (m *NavigationManager) Entries() []NavigationEvent { return m.entries }

// SetCap changes the entry cap. Used in tests; also re-applies the cap if the
// current slice exceeds the new cap.
func (m *NavigationManager) SetCap(cap int) {
	m.cap = cap
	if cap > 0 && len(m.entries) > cap {
		drop := len(m.entries) - cap
		m.entries = m.entries[drop:]
		m.index -= drop
		if m.index < -1 {
			m.index = -1
		}
	}
}
