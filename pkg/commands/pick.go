package commands

import "kvnd/lazyruin/pkg/models"

type PickCommand struct {
	ruin *RuinCommand
}

func NewPickCommand(ruin *RuinCommand) *PickCommand {
	return &PickCommand{ruin: ruin}
}

func (p *PickCommand) Pick(tags []string, any bool) ([]models.PickResult, error) {
	args := []string{"pick"}
	for _, tag := range tags {
		args = append(args, tag)
	}
	if any {
		args = append(args, "--any")
	}

	output, err := p.ruin.Execute(args...)
	if err != nil {
		return nil, err
	}

	return unmarshalJSON[[]models.PickResult](output)
}
