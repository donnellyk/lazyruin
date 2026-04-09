package commands

import "github.com/donnellyk/lazyruin/pkg/models"

type TagsCommand struct {
	ruin *RuinCommand
}

func NewTagsCommand(ruin *RuinCommand) *TagsCommand {
	return &TagsCommand{ruin: ruin}
}

func (t *TagsCommand) List() ([]models.Tag, error) {
	return ExecuteAndUnmarshal[[]models.Tag](t.ruin, "tags", "list")
}

func (t *TagsCommand) Rename(oldName, newName string) error {
	_, err := t.ruin.Execute("tags", "rename", oldName, newName, "-f")
	return err
}

func (t *TagsCommand) Delete(name string) error {
	_, err := t.ruin.Execute("tags", "delete", name, "-f")
	return err
}
