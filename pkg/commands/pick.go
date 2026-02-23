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
	args := []string{"pick"}
	for _, tag := range tags {
		args = append(args, tag)
		if strings.EqualFold(tag, "#done") {
			args = append(args, "--done")
		}
	}
	if opts.Date != "" {
		args = append(args, opts.Date)
	}
	if opts.Any {
		args = append(args, "--any")
	}
	if opts.Todo {
		args = append(args, "--todo")
	}
	if opts.All {
		args = append(args, "--all")
	}
	if opts.Filter != "" {
		args = append(args, "--filter", opts.Filter)
	}
	if opts.Parent != "" {
		args = append(args, "--parent", opts.Parent)
	}
	if len(opts.Notes) > 0 {
		args = append(args, "--notes", strings.Join(opts.Notes, ","))
	}

	output, err := p.ruin.Execute(args...)
	if err != nil {
		return nil, err
	}

	return unmarshalJSON[[]models.PickResult](output)
}
