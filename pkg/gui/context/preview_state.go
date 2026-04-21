package context

// PreviewLink represents a detected link in the preview content. Text is
// the full link (URL or "[[wiki]]") even when our own wordwrap has broken
// the link across multiple visual lines — the renderer iterates Segments
// to paint the highlight across every wrapped piece, and FollowLink/Open
// always sees the whole link.
type PreviewLink struct {
	Text     string
	Segments []PreviewLinkSegment
}

// PreviewLinkSegment is a single on-screen span of a PreviewLink.
type PreviewLinkSegment struct {
	Line int // absolute ns.Lines index
	Col  int // start column (visible characters, 0-indexed)
	Len  int // visible length within that line
}
