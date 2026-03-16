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

// Compose dispatches to the appropriate compose method based on bookmark type.
func (p *ParentCommand) Compose(bm models.ParentBookmark) (models.Note, []models.SourceMapEntry, error) {
	var args []string
	if bm.IsFileBased() {
		args = []string{"compose", "--file", bm.File, "--strip-title", "--strip-global-tags", "--normalize-headers", "--expand-embeds"}
	} else {
		args = []string{"compose", bm.UUID, "--strip-title", "--strip-global-tags", "--normalize-headers", "--sort", "created:desc", "--expand-embeds"}
	}

	output, err := p.ruin.Execute(args...)
	if err != nil {
		return models.Note{}, nil, err
	}

	return p.parseComposeResult(output, bm.Title)
}

func (p *ParentCommand) parseComposeResult(output []byte, title string) (models.Note, []models.SourceMapEntry, error) {
	root, err := unmarshalJSON[composeResult](output)
	if err != nil {
		return models.Note{}, nil, err
	}

	note := models.Note{
		UUID:    root.UUID,
		Title:   title,
		Path:    root.Path,
		Content: root.ComposedContent,
	}

	if root.UUID != "" {
		if metaOutput, err := p.ruin.Execute("get", "--uuid", root.UUID); err == nil {
			if meta, err := unmarshalJSON[*models.Note](metaOutput); err == nil && meta != nil {
				note.Tags = meta.Tags
				note.Created = meta.Created
				note.Parent = meta.Parent
			}
		}
	}

	return note, root.SourceMap, nil
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
