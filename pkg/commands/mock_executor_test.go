package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"kvnd/lazyruin/pkg/models"
)

// MockExecutor provides canned responses for testing.
type MockExecutor struct {
	vaultPath string
	notes     []models.Note
	tags      []models.Tag
	queries   []models.Query
	parents   []models.ParentBookmark
	compose   []byte // raw JSON for compose tree
	err       error
}

// NewMockExecutor creates a new mock executor.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		vaultPath: "/mock/vault",
	}
}

// WithNotes sets the notes to return for search commands.
func (m *MockExecutor) WithNotes(notes ...models.Note) *MockExecutor {
	m.notes = notes
	return m
}

// WithTags sets the tags to return for tags list.
func (m *MockExecutor) WithTags(tags ...models.Tag) *MockExecutor {
	m.tags = tags
	return m
}

// WithQueries sets the queries to return for query list.
func (m *MockExecutor) WithQueries(queries ...models.Query) *MockExecutor {
	m.queries = queries
	return m
}

// WithParents sets the parent bookmarks to return.
func (m *MockExecutor) WithParents(parents ...models.ParentBookmark) *MockExecutor {
	m.parents = parents
	return m
}

// WithCompose sets the raw JSON for compose tree responses.
func (m *MockExecutor) WithCompose(data []byte) *MockExecutor {
	m.compose = data
	return m
}

// WithError sets an error to return.
func (m *MockExecutor) WithError(err error) *MockExecutor {
	m.err = err
	return m
}

// Execute returns canned JSON responses based on the command.
func (m *MockExecutor) Execute(args ...string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	cmd := args[0]

	switch cmd {
	case "today", "yesterday":
		return json.Marshal(m.notes)

	case "search":
		// Filter notes by tag if searching by tag
		if len(args) > 1 && strings.HasPrefix(args[1], "#") {
			tag := strings.TrimPrefix(args[1], "#")
			var filtered []models.Note
			for _, n := range m.notes {
				for _, t := range n.Tags {
					if t == tag {
						filtered = append(filtered, n)
						break
					}
				}
			}
			return json.Marshal(filtered)
		}
		return json.Marshal(m.notes)

	case "tags":
		if len(args) > 1 && args[1] == "list" {
			return json.Marshal(m.tags)
		}
		return []byte("{}"), nil

	case "query":
		if len(args) > 1 {
			switch args[1] {
			case "list":
				return json.Marshal(m.queries)
			case "run":
				return json.Marshal(m.notes)
			case "save", "delete":
				return []byte("{}"), nil
			}
		}
		return []byte("{}"), nil

	case "parent":
		if len(args) > 1 {
			switch args[1] {
			case "list":
				return json.Marshal(m.parents)
			case "delete":
				return []byte("{}"), nil
			}
		}
		return []byte("{}"), nil

	case "compose":
		if m.compose != nil {
			return m.compose, nil
		}
		return []byte("{}"), nil

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

// VaultPath returns the mock vault path.
func (m *MockExecutor) VaultPath() string {
	return m.vaultPath
}
