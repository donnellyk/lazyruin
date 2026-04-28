package commands

import (
	"encoding/json"
	"fmt"

	"github.com/donnellyk/lazyruin/pkg/models"
)

// EmbedCommand wraps `ruin embed eval` for evaluating a single dynamic embed
// string and unpacking the typed JSON envelope.
type EmbedCommand struct {
	ruin *RuinCommand
}

func NewEmbedCommand(ruin *RuinCommand) *EmbedCommand {
	return &EmbedCommand{ruin: ruin}
}

// EmbedType identifies which dynamic-embed type a result envelope describes.
type EmbedType string

const (
	EmbedTypeSearch  EmbedType = "search"
	EmbedTypeQuery   EmbedType = "query"
	EmbedTypePick    EmbedType = "pick"
	EmbedTypeCompose EmbedType = "compose"
)

// EmbedComposeResult is the `compose:` payload of an embed-eval response.
type EmbedComposeResult struct {
	ExpandedMarkdown string                 `json:"expanded_markdown"`
	SourceMap        []EmbedComposeSourceID `json:"source_map"`
}

// EmbedComposeSourceID maps a region of expanded markdown back to a source note.
type EmbedComposeSourceID struct {
	UUID      string `json:"uuid"`
	Path      string `json:"path"`
	Title     string `json:"title"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// EmbedResult is the decoded result of `ruin embed eval --json`. Exactly one
// of Notes / Picks / Compose is populated, indicated by Type.
type EmbedResult struct {
	Type    EmbedType
	Query   string
	Notes   []models.Note       // for Type == search or query
	Picks   []models.PickResult // for Type == pick
	Compose *EmbedComposeResult // for Type == compose
}

// Eval runs `ruin embed eval <embed> --json` and unmarshals the typed
// envelope into an EmbedResult. The default --json flag is appended by the
// shared executor; callers must not include it themselves.
func (e *EmbedCommand) Eval(embed string) (*EmbedResult, error) {
	out, err := e.ruin.Execute("embed", "eval", embed)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Type    EmbedType       `json:"type"`
		Query   string          `json:"query"`
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil {
		return nil, fmt.Errorf("parse embed envelope: %w", err)
	}

	res := &EmbedResult{Type: envelope.Type, Query: envelope.Query}

	// JSON `null` or absent results decode as zero-value slices/struct
	// without invoking Unmarshal, so partial responses don't fail the
	// caller.
	hasResults := len(envelope.Results) > 0 && string(envelope.Results) != "null"

	switch envelope.Type {
	case EmbedTypeSearch, EmbedTypeQuery:
		if hasResults {
			if err := json.Unmarshal(envelope.Results, &res.Notes); err != nil {
				return nil, fmt.Errorf("parse %s results: %w", envelope.Type, err)
			}
		}
	case EmbedTypePick:
		if hasResults {
			if err := json.Unmarshal(envelope.Results, &res.Picks); err != nil {
				return nil, fmt.Errorf("parse pick results: %w", err)
			}
		}
	case EmbedTypeCompose:
		if hasResults {
			var compose EmbedComposeResult
			if err := json.Unmarshal(envelope.Results, &compose); err != nil {
				return nil, fmt.Errorf("parse compose results: %w", err)
			}
			res.Compose = &compose
		}
	default:
		return nil, fmt.Errorf("unknown embed type %q", envelope.Type)
	}

	return res, nil
}
