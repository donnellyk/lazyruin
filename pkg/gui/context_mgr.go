package gui

import (
	"sync"

	"kvnd/lazyruin/pkg/gui/types"
)

// ContextMgr manages the context stack with thread-safe access and O(1) key lookup.
type ContextMgr struct {
	mu     sync.Mutex
	stack  []types.ContextKey
	lookup map[types.ContextKey]types.Context
}

// NewContextMgr creates a ContextMgr with the default stack.
func NewContextMgr() *ContextMgr {
	return &ContextMgr{
		stack:  []types.ContextKey{"notes"},
		lookup: make(map[types.ContextKey]types.Context),
	}
}

// Register adds a context to the O(1) lookup map.
func (m *ContextMgr) Register(ctx types.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lookup[ctx.GetKey()] = ctx
}

// ContextByKey returns the context for a key, or nil if not registered.
func (m *ContextMgr) ContextByKey(key types.ContextKey) types.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lookup[key]
}

// Current returns the top of the context stack.
func (m *ContextMgr) Current() types.ContextKey {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.stack) == 0 {
		return "notes"
	}
	return m.stack[len(m.stack)-1]
}

// Previous returns the second-from-top of the context stack.
func (m *ContextMgr) Previous() types.ContextKey {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.stack) < 2 {
		return "notes"
	}
	return m.stack[len(m.stack)-2]
}

// Push appends a key to the stack.
func (m *ContextMgr) Push(key types.ContextKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stack = append(m.stack, key)
}

// Pop removes the top of the stack, keeping at least one entry.
func (m *ContextMgr) Pop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.stack) > 1 {
		m.stack = m.stack[:len(m.stack)-1]
	}
}

// Replace replaces the top of the stack.
func (m *ContextMgr) Replace(key types.ContextKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.stack) > 0 {
		m.stack[len(m.stack)-1] = key
	} else {
		m.stack = []types.ContextKey{key}
	}
}

// SetStack replaces the entire stack.
func (m *ContextMgr) SetStack(keys []types.ContextKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stack = keys
}
