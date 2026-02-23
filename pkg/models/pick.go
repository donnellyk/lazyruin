package models

type PickMatch struct {
	Line    int      `json:"line"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	Done    bool     `json:"done"`
}

type PickResult struct {
	UUID    string      `json:"uuid"`
	Title   string      `json:"title,omitempty"`
	File    string      `json:"file"`
	Matches []PickMatch `json:"matches"`
}
