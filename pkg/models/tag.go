package models

type Tag struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Scope []string `json:"scope"`
}
