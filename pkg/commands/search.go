package commands

import (
	"kvnd/lazyruin/pkg/models"
	"strconv"
)

type SearchOptions struct {
	Sort            string
	Limit           int
	IncludeContent  bool
	StripGlobalTags bool
	StripTitle      bool
	Everything      bool
	Link            bool
}

// applyDisplayFlags adds the common display flags (--content, --strip-global-tags,
// --strip-title) to the builder based on the options. This centralizes the logic
// that was previously duplicated across Search, get, Today, and Queries.Run.
func (o SearchOptions) applyDisplayFlags(b *ArgBuilder) *ArgBuilder {
	return b.
		AddIf(o.IncludeContent, "--content").
		AddIf(o.StripGlobalTags, "--strip-global-tags").
		AddIf(o.StripTitle, "--strip-title")
}

type SearchCommand struct {
	ruin *RuinCommand
}

func NewSearchCommand(ruin *RuinCommand) *SearchCommand {
	return &SearchCommand{ruin: ruin}
}

func (s *SearchCommand) Search(query string, opts SearchOptions) ([]models.Note, error) {
	b := NewArgBuilder("search").
		AddIf(query != "", query).
		AddIf(opts.Sort != "", "-s", opts.Sort).
		AddIf(opts.Limit > 0, "-l", strconv.Itoa(opts.Limit))
	opts.applyDisplayFlags(b).
		AddIf(opts.Everything, "--everything").
		AddIf(opts.Link, "--link")

	return ExecuteAndUnmarshal[[]models.Note](s.ruin, b.Build()...)
}

// Get fetches a single note by UUID with the given search options.
func (s *SearchCommand) Get(uuid string, opts SearchOptions) (*models.Note, error) {
	return s.get([]string{"--uuid", uuid}, opts)
}

// GetByTitle fetches a single note by title (case-insensitive substring match).
func (s *SearchCommand) GetByTitle(title string, opts SearchOptions) (*models.Note, error) {
	return s.get([]string{"--title", title}, opts)
}

// GetByPath fetches a single note by path (substring match).
func (s *SearchCommand) GetByPath(path string, opts SearchOptions) (*models.Note, error) {
	return s.get([]string{"--path", path}, opts)
}

func (s *SearchCommand) get(filter []string, opts SearchOptions) (*models.Note, error) {
	b := NewArgBuilder("get").Add(filter...)
	opts.applyDisplayFlags(b)

	return ExecuteAndUnmarshal[*models.Note](s.ruin, b.Build()...)
}

func (s *SearchCommand) Today() ([]models.Note, error) {
	return ExecuteAndUnmarshal[[]models.Note](s.ruin, "today", "--content")
}
