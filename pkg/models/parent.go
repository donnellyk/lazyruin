package models

type ParentBookmark struct {
	Name  string `json:"name"`
	UUID  string `json:"uuid,omitempty"`
	Title string `json:"title,omitempty"`
	File  string `json:"file,omitempty"`
}

func (p ParentBookmark) IsFileBased() bool {
	return p.File != ""
}
