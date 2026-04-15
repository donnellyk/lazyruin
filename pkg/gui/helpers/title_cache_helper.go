package helpers

import (
	"sync"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// TitleCacheHelper caches note UUID → title mappings so that parent badges
// can be resolved to titles even when the parent note is not a bookmark and
// not currently in the Notes panel (e.g., an older parent of a "today" note).
//
// The cache is populated two ways:
//   - PutNotes, called whenever a batch of notes is loaded from any source.
//   - ResolveUnknownParents, which fetches parents that aren't yet known.
type TitleCacheHelper struct {
	c *HelperCommon

	mu    sync.RWMutex
	cache map[string]string
}

func NewTitleCacheHelper(c *HelperCommon) *TitleCacheHelper {
	return &TitleCacheHelper{c: c, cache: make(map[string]string)}
}

// Get returns the cached title for a UUID.
func (h *TitleCacheHelper) Get(uuid string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t, ok := h.cache[uuid]
	return t, ok
}

// Put inserts a single UUID → title mapping. No-op on empty values.
func (h *TitleCacheHelper) Put(uuid, title string) {
	if uuid == "" || title == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache[uuid] = title
}

// PutNotes records titles from any notes with a non-empty UUID and Title.
func (h *TitleCacheHelper) PutNotes(notes []models.Note) {
	if len(notes) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, n := range notes {
		if n.UUID == "" || n.Title == "" {
			continue
		}
		h.cache[n.UUID] = n.Title
	}
}

// ResolveUnknownParents fetches titles for parent UUIDs referenced by the
// given notes that aren't already a bookmark or in the cache. Each distinct
// unknown parent is fetched once via `ruin get --uuid`; failures are ignored
// so the render can still fall back to a truncated UUID.
func (h *TitleCacheHelper) ResolveUnknownParents(notes []models.Note) {
	if len(notes) == 0 {
		return
	}

	known := make(map[string]bool)
	for _, bm := range h.c.GuiCommon().Contexts().Queries.Parents {
		if bm.UUID != "" {
			known[bm.UUID] = true
		}
	}

	h.mu.RLock()
	for uuid := range h.cache {
		known[uuid] = true
	}
	h.mu.RUnlock()

	seen := make(map[string]bool)
	for _, n := range notes {
		p := n.Parent
		if p == "" || known[p] || seen[p] {
			continue
		}
		seen[p] = true
		parent, err := h.c.RuinCmd().Search.Get(p, commands.SearchOptions{})
		if err != nil || parent == nil || parent.Title == "" {
			continue
		}
		h.Put(parent.UUID, parent.Title)
	}
}
