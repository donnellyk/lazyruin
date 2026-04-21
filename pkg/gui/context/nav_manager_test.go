package context

import "testing"

func newEvt(title string) NavigationEvent {
	return NavigationEvent{
		ContextKey: "cardList",
		Title:      title,
		Snapshot:   title,
	}
}

func newEvtID(title, id string) NavigationEvent {
	return NavigationEvent{
		ContextKey: "cardList",
		Title:      title,
		Snapshot:   title,
		ID:         id,
	}
}

// TestNavigationManager_RecordDedupsByID verifies the LRU-style dedup:
// recording an entry whose ID matches a prior entry evicts the prior
// entry so the new one is the unique occurrence and sits at the top of
// the stack. Entries with empty ID are never deduped.
func TestNavigationManager_RecordDedupsByID(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvtID("Note A", "n:a"))
	m.Record(newEvtID("Note B", "n:b"))
	m.Record(newEvtID("Note C", "n:c"))
	m.Record(newEvtID("Note A (again)", "n:a"))

	if m.Len() != 3 {
		t.Fatalf("Len() = %d, want 3 (A should have been deduped)", m.Len())
	}
	titles := make([]string, 0, m.Len())
	for _, e := range m.Entries() {
		titles = append(titles, e.Title)
	}
	want := []string{"Note B", "Note C", "Note A (again)"}
	if len(titles) != len(want) {
		t.Fatalf("titles = %v, want %v", titles, want)
	}
	for i, w := range want {
		if titles[i] != w {
			t.Errorf("titles[%d] = %q, want %q", i, titles[i], w)
		}
	}
	if m.Index() != 2 {
		t.Errorf("Index() = %d, want 2 (A-again on top)", m.Index())
	}
}

// TestNavigationManager_RecordEmptyIDDoesNotDedup ensures entries with
// no dedup ID (e.g. multi-card search results) remain distinct even when
// their titles match.
func TestNavigationManager_RecordEmptyIDDoesNotDedup(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("same"))
	m.Record(newEvt("same"))
	if m.Len() != 2 {
		t.Errorf("Len() = %d, want 2 (empty IDs should not dedup)", m.Len())
	}
}

func TestNavigationManager_Empty(t *testing.T) {
	m := NewNavigationManager()
	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}
	if m.Index() != -1 {
		t.Errorf("Index() = %d, want -1", m.Index())
	}
	if _, ok := m.Current(); ok {
		t.Error("Current() ok=true on empty manager, want false")
	}
	if _, ok := m.Back(); ok {
		t.Error("Back() ok=true on empty manager, want false")
	}
	if _, ok := m.Forward(); ok {
		t.Error("Forward() ok=true on empty manager, want false")
	}
}

func TestNavigationManager_RecordAndCurrent(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("a"))
	m.Record(newEvt("b"))
	m.Record(newEvt("c"))

	if m.Len() != 3 {
		t.Errorf("Len() = %d, want 3", m.Len())
	}
	if m.Index() != 2 {
		t.Errorf("Index() = %d, want 2", m.Index())
	}
	cur, ok := m.Current()
	if !ok {
		t.Fatal("Current() ok=false, want true")
	}
	if cur.Title != "c" {
		t.Errorf("Current().Title = %q, want %q", cur.Title, "c")
	}
	if cur.Timestamp.IsZero() {
		t.Error("Record did not set Timestamp")
	}
}

func TestNavigationManager_BackForward(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("a"))
	m.Record(newEvt("b"))
	m.Record(newEvt("c"))

	evt, ok := m.Back()
	if !ok || evt.Title != "b" {
		t.Errorf("Back() = (%v, %v), want (b, true)", evt.Title, ok)
	}
	evt, ok = m.Back()
	if !ok || evt.Title != "a" {
		t.Errorf("Back() = (%v, %v), want (a, true)", evt.Title, ok)
	}
	if _, ok := m.Back(); ok {
		t.Error("Back() from first entry ok=true, want false")
	}
	evt, ok = m.Forward()
	if !ok || evt.Title != "b" {
		t.Errorf("Forward() = (%v, %v), want (b, true)", evt.Title, ok)
	}
	evt, ok = m.Forward()
	if !ok || evt.Title != "c" {
		t.Errorf("Forward() = (%v, %v), want (c, true)", evt.Title, ok)
	}
	if _, ok := m.Forward(); ok {
		t.Error("Forward() from last entry ok=true, want false")
	}
}

func TestNavigationManager_RecordTruncatesForward(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("a"))
	m.Record(newEvt("b"))
	m.Record(newEvt("c"))

	m.Back()
	m.Back() // at "a"
	if m.Index() != 0 {
		t.Fatalf("Index() = %d, want 0", m.Index())
	}
	m.Record(newEvt("d"))

	if m.Len() != 2 {
		t.Errorf("Len() = %d, want 2 (a, d)", m.Len())
	}
	if m.Index() != 1 {
		t.Errorf("Index() = %d, want 1", m.Index())
	}
	if _, ok := m.Forward(); ok {
		t.Error("Forward() after new record ok=true, want false (forward truncated)")
	}
}

func TestNavigationManager_UpdateCurrent(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("a"))
	m.UpdateCurrent("a-updated")
	cur, _ := m.Current()
	if cur.Snapshot != "a-updated" {
		t.Errorf("Snapshot = %v, want a-updated", cur.Snapshot)
	}
}

func TestNavigationManager_UpdateCurrentTitle(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("a"))
	m.UpdateCurrentTitle("A!")
	cur, _ := m.Current()
	if cur.Title != "A!" {
		t.Errorf("Title = %q, want A!", cur.Title)
	}
}

func TestNavigationManager_UpdateCurrentOnEmpty(t *testing.T) {
	m := NewNavigationManager()
	m.UpdateCurrent("noop")
	if m.Len() != 0 {
		t.Errorf("UpdateCurrent on empty created an entry: len=%d", m.Len())
	}
}

func TestNavigationManager_Cap(t *testing.T) {
	m := NewNavigationManager()
	m.SetCap(3)
	for i := range 5 {
		m.Record(newEvt(string(rune('a' + i))))
	}
	if m.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", m.Len())
	}
	if m.Index() != 2 {
		t.Fatalf("Index() = %d, want 2", m.Index())
	}
	cur, _ := m.Current()
	if cur.Title != "e" {
		t.Errorf("Current().Title = %q, want e", cur.Title)
	}
	// Oldest kept should be "c"
	m.Back()
	m.Back()
	oldest, _ := m.Current()
	if oldest.Title != "c" {
		t.Errorf("Oldest after cap = %q, want c", oldest.Title)
	}
}

func TestNavigationManager_Clear(t *testing.T) {
	m := NewNavigationManager()
	m.Record(newEvt("a"))
	m.Record(newEvt("b"))
	m.Clear()
	if m.Len() != 0 || m.Index() != -1 {
		t.Errorf("after Clear: Len=%d Index=%d, want 0 / -1", m.Len(), m.Index())
	}
}

func TestNavigationManager_DefaultCap(t *testing.T) {
	m := NewNavigationManager()
	for range DefaultNavHistoryCap + 10 {
		m.Record(newEvt("x"))
	}
	if m.Len() != DefaultNavHistoryCap {
		t.Errorf("Len() = %d, want %d", m.Len(), DefaultNavHistoryCap)
	}
}
