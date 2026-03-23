package commands

type LinkNewOpts struct {
	Title   string
	Tags    string
	Parent  string
	NoFetch bool
	Comment string
}

type LinkNewResult struct {
	Path  string `json:"path"`
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type LinkResolveResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	ResolvedVia string `json:"resolved_via"`
}

type LinkCommand struct {
	ruin *RuinCommand
}

func NewLinkCommand(ruin *RuinCommand) *LinkCommand {
	return &LinkCommand{ruin: ruin}
}

func (l *LinkCommand) New(url string, opts LinkNewOpts) (*LinkNewResult, error) {
	b := NewArgBuilder("link", "new", url).
		AddIf(opts.Title != "", "--title", opts.Title).
		AddIf(opts.Tags != "", "--tags", opts.Tags).
		AddIf(opts.Parent != "", "--parent", opts.Parent).
		AddIf(opts.NoFetch, "--no-fetch").
		AddIf(opts.Comment != "", "--comment", opts.Comment)

	return ExecuteAndUnmarshal[*LinkNewResult](l.ruin, b.Build()...)
}

func (l *LinkCommand) Resolve(url string) (*LinkResolveResult, error) {
	return ExecuteAndUnmarshal[*LinkResolveResult](l.ruin, "link", "resolve", url)
}
