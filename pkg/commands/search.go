package commands

import (
	"encoding/json"
	"kvnd/lazyruin/pkg/models"
	"strconv"
	"strings"
)

type SearchOptions struct {
	Sort  string
	Limit int
}

type SearchCommand struct {
	ruin *RuinCommand
}

func NewSearchCommand(ruin *RuinCommand) *SearchCommand {
	return &SearchCommand{ruin: ruin}
}

func (s *SearchCommand) Search(query string, opts SearchOptions) ([]models.Note, error) {
	args := []string{"search", query} // Building slice that will be concatted to form command

	if opts.Sort != "" {
		args = append(args, "-s", opts.Sort)
	}
	if opts.Limit > 0 {
		args = append(args, "-l", strconv.Itoa(opts.Limit))
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
	output, err := s.ruin.Execute("today")
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
	output, err := s.ruin.Execute("yesterday")
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
	return s.Search(tag, SearchOptions{})
}
