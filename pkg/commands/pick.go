package commands

import (
	"strings"

	"kvnd/lazyruin/pkg/models"
)

type PickCommand struct {
	ruin *RuinCommand
}

func NewPickCommand(ruin *RuinCommand) *PickCommand {
	return &PickCommand{ruin: ruin}
}

// PickOpts holds optional flags for the pick command.
type PickOpts struct {
	Any    bool
	Filter string   // --filter: note-level metadata filter
	Date   string   // positional @date: line-level date filter
	Todo   bool     // --todo flag
	All    bool     // --all flag (include done todos)
	Parent string   // --parent: scope to children of a parent note
	Notes  []string // --notes: scope to specific note UUIDs
}

func (p *PickCommand) Pick(tags []string, opts PickOpts) ([]models.PickResult, error) {
	b := NewArgBuilder("pick")
	for _, tag := range tags {
		b.Add(tag)
		if strings.EqualFold(tag, "#done") {
			b.Add("--done")
		}
	}
	b.AddIf(opts.Date != "", opts.Date).
		AddIf(opts.Any, "--any").
		AddIf(opts.Todo, "--todo").
		AddIf(opts.All, "--all").
		AddIf(opts.Filter != "", "--filter", opts.Filter).
		AddIf(opts.Parent != "", "--parent", opts.Parent).
		AddIf(len(opts.Notes) > 0, "--notes", strings.Join(opts.Notes, ","))

	return ExecuteAndUnmarshal[[]models.PickResult](p.ruin, b.Build()...)
}
