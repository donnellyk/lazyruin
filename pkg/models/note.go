package models

import (
	"strings"
	"time"

	"github.com/donnellyk/ruin-note-cli/pkg/notetext"
)

// HasDoneTag returns true if the line contains a #done inline tag
// (case-insensitive). Uses the shared notetext extractor so tags inside
// code spans, fenced code blocks, and markdown links are correctly ignored.
func HasDoneTag(line string) bool {
	for _, tag := range notetext.ExtractTags(line) {
		if strings.EqualFold(tag, "#done") {
			return true
		}
	}
	return false
}

type Note struct {
	UUID       string    `json:"uuid"`
	Path       string    `json:"path"`
	Title      string    `json:"title"`
	Content    string    `json:"content,omitempty"`
	Tags       []string  `json:"tags"`
	InlineTags []string  `json:"inline_tags"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	Parent     string    `json:"parent,omitempty"`
	Order      *int      `json:"order,omitempty"`
}

func (n *Note) ShortDate() string {
	if n.Created.IsZero() {
		return ""
	}
	return n.Created.Format("Jan 02")
}

// FirstLine returns the first non-empty line of the note's content.
func (n *Note) FirstLine() string {
	for line := range strings.SplitSeq(n.Content, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			return t
		}
	}
	return ""
}

// JoinDot joins non-empty parts with " · ".
func JoinDot(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, " · ")
}

// SourceMapEntry maps a line range in composed output back to its source child note.
type SourceMapEntry struct {
	UUID      string `json:"uuid"`
	Path      string `json:"path"`
	Title     string `json:"title"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

func (n *Note) IsLink() bool {
	for _, tag := range n.Tags {
		if strings.EqualFold(strings.TrimPrefix(tag, "#"), "link") {
			return true
		}
	}
	return false
}

func (n *Note) URL() string {
	if !n.IsLink() {
		return ""
	}
	line := n.FirstLine()
	if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		return line
	}
	return ""
}

// formatTag ensures a tag has a leading "#".
func formatTag(tag string) string {
	if len(tag) > 0 && tag[0] != '#' {
		return "#" + tag
	}
	return tag
}

// joinTags formats a slice of tags as comma-separated "#tag" strings.
func joinTags(tags []string) string {
	formatted := make([]string, len(tags))
	for i, tag := range tags {
		formatted[i] = formatTag(tag)
	}
	return strings.Join(formatted, ", ")
}

func (n *Note) GlobalTagsString() string {
	return joinTags(n.Tags)
}

func (n *Note) TagsString() string {
	combined := make([]string, 0, len(n.Tags)+len(n.InlineTags))
	combined = append(combined, n.Tags...)
	combined = append(combined, n.InlineTags...)
	return joinTags(combined)
}
