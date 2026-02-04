package models

import "time"

type Note struct {
	UUID    string    `json:"uuid"`
	Path    string    `json:"path"`
	Title   string    `json:"title"`
	Content string    `json:"content,omitempty"`
	Tags    []string  `json:"tags"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func (n *Note) ShortDate() string {
	return n.Created.Format("Jan 02")
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
	return result
}
