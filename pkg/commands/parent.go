package commands

import "kvnd/lazyruin/pkg/models"

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

	return unmarshalJSON[[]models.ParentBookmark](output)
}

type composeResult struct {
	UUID            string                  `json:"uuid"`
	Title           string                  `json:"title"`
	Path            string                  `json:"path"`
	ComposedContent string                  `json:"composed_content"`
	SourceMap       []models.SourceMapEntry `json:"source_map"`
}

// ComposeFlat runs compose and returns the fully-assembled document as a single Note
// plus a source map that maps composed line ranges back to their source child notes.
func (p *ParentCommand) ComposeFlat(uuid, title string) (models.Note, []models.SourceMapEntry, error) {
	output, err := p.ruin.Execute("compose", uuid, "--strip-title", "--strip-global-tags", "--normalize-headers", "--sort", "created:desc")
	if err != nil {
		return models.Note{}, nil, err
	}

	root, err := unmarshalJSON[composeResult](output)
	if err != nil {
		return models.Note{}, nil, err
	}

	return models.Note{
		UUID:    root.UUID,
		Title:   title,
		Path:    root.Path,
		Content: root.ComposedContent,
	}, root.SourceMap, nil
}

// Save creates a parent bookmark via `parent save <name> <note>`.
func (p *ParentCommand) Save(name, noteRef string) error {
	_, err := p.ruin.Execute("parent", "save", name, noteRef, "--force")
	return err
}

func (p *ParentCommand) Delete(name string) error {
	_, err := p.ruin.Execute("parent", "delete", name, "--force")
	return err
}

// ChildInfo represents a child note returned by parent children.
type ChildInfo struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// Children returns the children of a note via `parent children`.
func (p *ParentCommand) Children(noteRef string) ([]ChildInfo, error) {
	output, err := p.ruin.Execute("parent", "children", noteRef, "--recursive")
	if err != nil {
		return nil, err
	}
	return unmarshalJSON[[]ChildInfo](output)
}

// TreeNode represents a node in the parent tree.
type TreeNode struct {
	UUID     string     `json:"uuid"`
	Title    string     `json:"title"`
	Children []TreeNode `json:"children,omitempty"`
}

// Tree returns the parent-child tree for a note.
func (p *ParentCommand) Tree(noteRef string) (*TreeNode, error) {
	output, err := p.ruin.Execute("parent", "tree", noteRef)
	if err != nil {
		return nil, err
	}
	return unmarshalJSON[*TreeNode](output)
}
