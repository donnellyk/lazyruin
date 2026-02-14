package models

import (
	"strings"
	"time"
)

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

func (n *Note) GlobalTagsString() string {
	result := ""
	for i, tag := range n.Tags {
		if i > 0 {
			result += ", "
		}
		if len(tag) > 0 && tag[0] != '#' {
			result += "#"
		}
		result += tag
	}
	return result
}

func (n *Note) TagsString() string {
	result := ""
	for i, tag := range n.Tags {
		if i > 0 {
			result += ", "
		}
		if len(tag) > 0 && tag[0] != '#' {
			result += "#"
		}
		result += tag
	}
	for _, tag := range n.InlineTags {
		if result != "" {
			result += ", "
		}
		if len(tag) > 0 && tag[0] != '#' {
			result += "#"
		}
		result += tag
	}
	return result
}
