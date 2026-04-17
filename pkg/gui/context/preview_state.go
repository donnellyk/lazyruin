package context

// PreviewLink represents a detected link in the preview content.
type PreviewLink struct {
	Text string // display text (wiki-link target or URL)
	Line int    // absolute line number in the rendered preview
	Col  int    // start column (visible characters, 0-indexed)
	Len  int    // visible length of the link text
}
