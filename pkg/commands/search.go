package commands

import (
	"encoding/json"
	"kvnd/lazyruin/pkg/models"
	"strconv"
	"strings"
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

// DefaultContentOptions returns SearchOptions with content included
// Content is fetched in full; stripping is done at render time based on UI toggles
func DefaultContentOptions() SearchOptions {
	return SearchOptions{
		IncludeContent: true,
	}
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

	var notes []models.Note
	if err := json.Unmarshal(output, &notes); err != nil {
		return nil, err
	}

	return notes, nil
}

func (s *SearchCommand) Today() ([]models.Note, error) {
	output, err := s.ruin.Execute("today", "--content")
	if err != nil {
		return nil, err
	}

	var notes []models.Note
	if err := json.Unmarshal(output, &notes); err != nil {
		return []models.Note{}, nil // Empty results are not an error
	}

	return notes, nil
}

func (s *SearchCommand) Yesterday() ([]models.Note, error) {
	output, err := s.ruin.Execute("yesterday", "--content")
	if err != nil {
		return nil, err
	}

	var notes []models.Note
	if err := json.Unmarshal(output, &notes); err != nil {
		return []models.Note{}, nil // Empty results are not an error
	}

	return notes, nil
}

func (s *SearchCommand) ByTag(tag string) ([]models.Note, error) {
	if !strings.HasPrefix(tag, "#") {
		tag = "#" + tag
	}
	opts := DefaultContentOptions()
	return s.Search(tag, opts)
}

// Recent returns notes from the last 7 days
func (s *SearchCommand) Recent(limit int) ([]models.Note, error) {
	opts := DefaultContentOptions()
	opts.Sort = "created:desc"
	opts.Limit = limit
	return s.Search("created:7d", opts)
}

// All returns all notes sorted by creation date
func (s *SearchCommand) All(limit int) ([]models.Note, error) {
	opts := DefaultContentOptions()
	opts.Sort = "created:desc"
	opts.Limit = limit
	// Use a wide date range since empty query is not allowed
	return s.Search("created:10000d", opts)
}
