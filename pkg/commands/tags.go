package commands

import (
	"encoding/json"
	"kvnd/lazyruin/pkg/models"
)

type TagsCommand struct {
	ruin *RuinCommand
}

func NewTagsCommand(ruin *RuinCommand) *TagsCommand {
	return &TagsCommand{ruin: ruin}
}

func (t *TagsCommand) List() ([]models.Tag, error) {
	output, err := t.ruin.Execute("tags", "list")
	if err != nil {
		return nil, err
	}

	var tags []models.Tag
	if err := json.Unmarshal(output, &tags); err != nil {
		return []models.Tag{}, nil
	}

	return tags, nil
}

func (t *TagsCommand) Rename(oldName, newName string) error {
	_, err := t.ruin.Execute("tags", "rename", oldName, newName, "-f")
	return err
}

func (t *TagsCommand) Delete(name string) error {
	_, err := t.ruin.Execute("tags", "delete", name, "-f")
	return err
}
