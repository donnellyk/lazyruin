package types

// CompletionItem represents a single suggestion in the completion dropdown.
type CompletionItem struct {
	Label              string // display text (e.g. "#project")
	InsertText         string // text to insert (e.g. "#project")
	Detail             string // right-aligned detail (e.g. "(5)")
	ContinueCompleting bool   // if true, don't add trailing space -- allows chaining into next trigger
	Value              string // opaque data (e.g. UUID) for use by accept handlers
	PrependToLine      bool   // if true, prepend InsertText to existing line content instead of replacing trigger
}

// CompletionTrigger defines a prefix that activates completion with a candidate provider.
type CompletionTrigger struct {
	Prefix     string
	Candidates func(filter string) []CompletionItem
}

// ParentDrillEntry records a parent selected during drill-down navigation.
type ParentDrillEntry struct {
	Name string
	UUID string
}

// CompletionState tracks the current state of a completion session.
type CompletionState struct {
	Active        bool
	TriggerStart  int // byte offset where the trigger token starts
	Items         []CompletionItem
	SelectedIndex int
	ParentDrill   []ParentDrillEntry // stack of drilled-into parents for > completion

	// FallbackCandidates is called when no trigger prefix matches the current token.
	// It receives the token text and its start position, returning ambient suggestions.
	// Used by the search popup for real-time date parsing.
	FallbackCandidates func(token string) []CompletionItem
}

// NewCompletionState returns an initialized CompletionState.
func NewCompletionState() *CompletionState {
	return &CompletionState{}
}

// Dismiss fully resets the completion state, hiding the suggestion dropdown.
func (s *CompletionState) Dismiss() {
	s.Active = false
	s.Items = nil
	s.SelectedIndex = 0
	s.TriggerStart = 0
	s.ParentDrill = nil
}
