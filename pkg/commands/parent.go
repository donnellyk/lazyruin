package commands

import "kvnd/lazyruin/pkg/models"

type ParentCommand struct {
	ruin *RuinCommand
}

func NewParentCommand(ruin *RuinCommand) *ParentCommand {
	return &ParentCommand{ruin: ruin}
}

func (p *ParentCommand) List() ([]models.ParentBookmark, error) {
	return ExecuteAndUnmarshal[[]models.ParentBookmark](p.ruin, "parent", "list")
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

	result, err := ExecuteAndUnmarshal[composeResult](p.ruin, args...)
	if err != nil {
		return models.Note{}, nil, err
	}

	return p.buildComposeNote(result, bm.Title)
}

func (p *ParentCommand) buildComposeNote(root composeResult, title string) (models.Note, []models.SourceMapEntry, error) {
	note := models.Note{
		UUID:    root.UUID,
		Title:   title,
		Path:    root.Path,
		Content: root.ComposedContent,
	}

	if err := p.enrichNoteMetadata(&note); err != nil {
		return note, root.SourceMap, nil
	}

	return note, root.SourceMap, nil
}

// enrichNoteMetadata fetches and applies metadata (tags, created, parent) for a note with a UUID.
func (p *ParentCommand) enrichNoteMetadata(note *models.Note) error {
	if note.UUID == "" {
		return nil
	}

	meta, err := ExecuteAndUnmarshal[*models.Note](p.ruin, "get", "--uuid", note.UUID)
	if err != nil {
		return err
	}
	if meta == nil {
		return nil
	}

	note.Tags = meta.Tags
	note.Created = meta.Created
	note.Parent = meta.Parent
	return nil
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
	return ExecuteAndUnmarshal[[]ChildInfo](p.ruin, "parent", "children", noteRef, "--recursive")
}

// TreeNode represents a node in the parent tree.
type TreeNode struct {
	UUID     string     `json:"uuid"`
	Title    string     `json:"title"`
	Children []TreeNode `json:"children,omitempty"`
}

// Tree returns the parent-child tree for a note.
func (p *ParentCommand) Tree(noteRef string) (*TreeNode, error) {
	return ExecuteAndUnmarshal[*TreeNode](p.ruin, "parent", "tree", noteRef)
}
