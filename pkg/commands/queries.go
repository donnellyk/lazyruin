package commands

import "github.com/donnellyk/lazyruin/pkg/models"

type QueriesCommand struct {
	ruin *RuinCommand
}

func NewQueriesCommand(ruin *RuinCommand) *QueriesCommand {
	return &QueriesCommand{ruin: ruin}
}

func (q *QueriesCommand) List() ([]models.Query, error) {
	return ExecuteAndUnmarshal[[]models.Query](q.ruin, "query", "list")
}

func (q *QueriesCommand) Run(name string, opts SearchOptions) ([]models.Note, error) {
	b := NewArgBuilder("query", "run", name)
	opts.applyDisplayFlags(b)

	return ExecuteAndUnmarshal[[]models.Note](q.ruin, b.Build()...)
}

func (q *QueriesCommand) Save(name, queryStr string) error {
	_, err := q.ruin.Execute("query", "save", name, queryStr, "-f")
	return err
}

func (q *QueriesCommand) Delete(name string) error {
	_, err := q.ruin.Execute("query", "delete", name)
	return err
}
