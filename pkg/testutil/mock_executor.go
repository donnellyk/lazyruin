package testutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/models"
)

// MockExecutor provides canned responses for testing.
type MockExecutor struct {
	vaultPath     string
	notes         []models.Note
	tags          []models.Tag
	queries       []models.Query
	parents       []models.ParentBookmark
	compose       []byte // raw JSON for compose tree
	linkJSON      []byte // raw JSON for link command responses
	versionOutput string // raw output for `ruin --version`
	err           error
	Calls         [][]string // recorded argument lists from Execute calls
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

// WithLinkJSON sets the raw JSON for link command responses.
func (m *MockExecutor) WithLinkJSON(data []byte) *MockExecutor {
	m.linkJSON = data
	return m
}

// WithVersion sets the raw output returned for `ruin --version` calls.
// The default is "ruin version 0.1.0\n" when not set.
func (m *MockExecutor) WithVersion(output string) *MockExecutor {
	m.versionOutput = output
	return m
}

// WithError sets an error to return.
func (m *MockExecutor) WithError(err error) *MockExecutor {
	m.err = err
	return m
}

// Execute returns canned JSON responses based on the command.
func (m *MockExecutor) Execute(args ...string) ([]byte, error) {
	m.Calls = append(m.Calls, args)

	if m.err != nil {
		return nil, m.err
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	if args[0] == "--version" {
		if m.versionOutput != "" {
			return []byte(m.versionOutput), nil
		}
		return []byte("ruin version 0.1.0\n"), nil
	}

	cmd := args[0]

	switch cmd {
	case "today", "yesterday":
		return json.Marshal(m.notes)

	case "search":
		// Filter notes by parent UUID
		if len(args) > 1 && strings.HasPrefix(args[1], "parent:") {
			parentUUID := strings.TrimPrefix(args[1], "parent:")
			var filtered []models.Note
			for _, n := range m.notes {
				if n.Parent == parentUUID {
					filtered = append(filtered, n)
				}
			}
			return json.Marshal(filtered)
		}
		// Filter notes by tag if searching by tag (matches both global and inline)
		if len(args) > 1 && strings.HasPrefix(args[1], "#") {
			tag := args[1]
			tagBare := strings.TrimPrefix(tag, "#")
			var filtered []models.Note
			for _, n := range m.notes {
				found := false
				for _, t := range n.Tags {
					if t == tagBare || t == tag {
						found = true
						break
					}
				}
				if !found {
					for _, t := range n.InlineTags {
						if t == tagBare || t == tag {
							found = true
							break
						}
					}
				}
				if found {
					filtered = append(filtered, n)
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

	case "get":
		return m.handleGet(args)

	case "compose":
		if m.compose != nil {
			return m.compose, nil
		}
		return []byte("{}"), nil

	case "link":
		return m.handleLink(args)

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

func (m *MockExecutor) handleGet(args []string) ([]byte, error) {
	for i, a := range args {
		switch a {
		case "--uuid":
			if i+1 < len(args) {
				uuid := args[i+1]
				for _, n := range m.notes {
					if n.UUID == uuid {
						return json.Marshal(&n)
					}
				}
			}
		case "--path":
			if i+1 < len(args) {
				sub := args[i+1]
				for _, n := range m.notes {
					if strings.Contains(n.Path, sub) {
						return json.Marshal(&n)
					}
				}
			}
		case "--title":
			if i+1 < len(args) {
				sub := strings.ToLower(args[i+1])
				for _, n := range m.notes {
					if strings.Contains(strings.ToLower(n.Title), sub) {
						return json.Marshal(&n)
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("note not found")
}

func (m *MockExecutor) handleLink(args []string) ([]byte, error) {
	if m.linkJSON != nil {
		return m.linkJSON, nil
	}
	return []byte("{}"), nil
}

// VaultPath returns the mock vault path.
func (m *MockExecutor) VaultPath() string {
	return m.vaultPath
}
