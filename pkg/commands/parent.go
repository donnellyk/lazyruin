package commands

import (
	"encoding/json"
	"strings"

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

type composeNode struct {
	UUID     string        `json:"uuid"`
	Title    string        `json:"title"`
	Path     string        `json:"path"`
	Content  string        `json:"content"`
	Children []composeNode `json:"children"`
}

func (p *ParentCommand) Compose(uuid string) ([]models.Note, error) {
	output, err := p.ruin.Execute("compose", uuid, "--strip-title", "--strip-global-tags")
	if err != nil {
		return nil, err
	}

	var root composeNode
	if err := json.Unmarshal(output, &root); err != nil {
		return nil, err
	}

	var notes []models.Note
	flattenComposeDFS(root, &notes)

	return notes, nil
}

// ComposeFlat assembles the tree into a single Note whose Content is the
// concatenation of every node's content (DFS order), separated by headings.
func (p *ParentCommand) ComposeFlat(uuid, title string) (models.Note, error) {
	notes, err := p.Compose(uuid)
	if err != nil {
		return models.Note{}, err
	}

	var sb strings.Builder
	for i, n := range notes {
		if i > 0 {
			sb.WriteString("\n")
		}
		if i > 0 && n.Title != "" {
			sb.WriteString("## ")
			sb.WriteString(n.Title)
			sb.WriteString("\n\n")
		}
		sb.WriteString(n.Content)
		sb.WriteString("\n")
	}

	return models.Note{
		UUID:    uuid,
		Title:   title,
		Content: sb.String(),
	}, nil
}

func flattenComposeDFS(node composeNode, out *[]models.Note) {
	*out = append(*out, models.Note{
		UUID:    node.UUID,
		Title:   node.Title,
		Path:    node.Path,
		Content: node.Content,
	})
	for _, child := range node.Children {
		flattenComposeDFS(child, out)
	}
}

func (p *ParentCommand) Delete(name string) error {
	_, err := p.ruin.Execute("parent", "delete", name, "--force")
	return err
}
