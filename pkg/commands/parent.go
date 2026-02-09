package commands

import (
	"encoding/json"

	"kvnd/lazyruin/pkg/models"
)

type ParentCommand struct {
	ruin *RuinCommand
}

func NewParentCommand(ruin *RuinCommand) *ParentCommand {
	return &ParentCommand{ruin: ruin}
}

func (p *ParentCommand) List() ([]models.ParentBookmark, error) {
	output, err := p.ruin.Execute("parent", "list")
	if err != nil {
		return nil, err
	}

	var parents []models.ParentBookmark
	if err := json.Unmarshal(output, &parents); err != nil {
		return nil, err
	}

	return parents, nil
}

type composeResult struct {
	UUID    string `json:"uuid"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// ComposeFlat runs compose and returns the fully-assembled document as a single Note.
func (p *ParentCommand) ComposeFlat(uuid, title string) (models.Note, error) {
	output, err := p.ruin.Execute("compose", uuid, "--strip-title", "--strip-global-tags", "--content")
	if err != nil {
		return models.Note{}, err
	}

	var root composeResult
	if err := json.Unmarshal(output, &root); err != nil {
		return models.Note{}, err
	}

	return models.Note{
		UUID:    root.UUID,
		Title:   title,
		Path:    root.Path,
		Content: root.Content,
	}, nil
}

func (p *ParentCommand) Delete(name string) error {
	_, err := p.ruin.Execute("parent", "delete", name, "--force")
	return err
}
