package commands

import (
	"kvnd/lazyruin/pkg/models"
	"strconv"
)

type SearchOptions struct {
	Sort            string
	Limit           int
	IncludeContent  bool
	StripGlobalTags bool
	StripTitle      bool
}

type SearchCommand struct {
	ruin *RuinCommand
}

func NewSearchCommand(ruin *RuinCommand) *SearchCommand {
	return &SearchCommand{ruin: ruin}
}

func (s *SearchCommand) Search(query string, opts SearchOptions) ([]models.Note, error) {
	args := []string{"search", query}

	if opts.Sort != "" {
		args = append(args, "-s", opts.Sort)
	}
	if opts.Limit > 0 {
		args = append(args, "-l", strconv.Itoa(opts.Limit))
	}
	if opts.IncludeContent {
		args = append(args, "--content")
	}
	if opts.StripGlobalTags {
		args = append(args, "--strip-global-tags")
	}
	if opts.StripTitle {
		args = append(args, "--strip-title")
	}

	output, err := s.ruin.Execute(args...)
	if err != nil {
		return nil, err
	}

	return unmarshalJSON[[]models.Note](output)
}

func (s *SearchCommand) Today() ([]models.Note, error) {
	output, err := s.ruin.Execute("today", "--content")
	if err != nil {
		return nil, err
	}

	notes, err := unmarshalJSON[[]models.Note](output)
	if err != nil {
		return []models.Note{}, nil // Empty results are not an error
	}
	return notes, nil
}
