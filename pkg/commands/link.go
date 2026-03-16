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
	args := []string{"link", "new", url}
	if opts.Title != "" {
		args = append(args, "--title", opts.Title)
	}
	if opts.Tags != "" {
		args = append(args, "--tags", opts.Tags)
	}
	if opts.Parent != "" {
		args = append(args, "--parent", opts.Parent)
	}
	if opts.NoFetch {
		args = append(args, "--no-fetch")
	}
	if opts.Comment != "" {
		args = append(args, "--comment", opts.Comment)
	}
	output, err := l.ruin.Execute(args...)
	if err != nil {
		return nil, err
	}
	return unmarshalJSON[*LinkNewResult](output)
}

func (l *LinkCommand) Resolve(url string) (*LinkResolveResult, error) {
	output, err := l.ruin.Execute("link", "resolve", url)
	if err != nil {
		return nil, err
	}
	return unmarshalJSON[*LinkResolveResult](output)
}
