package commands

import "kvnd/lazyruin/pkg/models"

type QueriesCommand struct {
	ruin *RuinCommand
}

func NewQueriesCommand(ruin *RuinCommand) *QueriesCommand {
	return &QueriesCommand{ruin: ruin}
}

func (q *QueriesCommand) List() ([]models.Query, error) {
	output, err := q.ruin.Execute("query", "list")
	if err != nil {
		return nil, err
	}

	return unmarshalJSON[[]models.Query](output)
}

func (q *QueriesCommand) Run(name string) ([]models.Note, error) {
	output, err := q.ruin.Execute("query", "run", name)
	if err != nil {
		return nil, err
	}

	notes, err := unmarshalJSON[[]models.Note](output)
	if err != nil {
		return []models.Note{}, nil
	}
	return notes, nil
}

func (q *QueriesCommand) Save(name, queryStr string) error {
	_, err := q.ruin.Execute("query", "save", name, queryStr, "-f")
	return err
}

func (q *QueriesCommand) Delete(name string) error {
	_, err := q.ruin.Execute("query", "delete", name)
	return err
}
