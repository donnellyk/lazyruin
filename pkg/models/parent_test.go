package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParentBookmark_IsFileBased(t *testing.T) {
	tests := []struct {
		name     string
		bookmark ParentBookmark
		want     bool
	}{
		{
			name:     "UUID-based bookmark",
			bookmark: ParentBookmark{Name: "docs", UUID: "abc-123", Title: "Docs Root"},
			want:     false,
		},
		{
			name:     "file-based bookmark",
			bookmark: ParentBookmark{Name: "alpha", File: "project.yml", Title: "Project Alpha"},
			want:     true,
		},
		{
			name:     "empty bookmark",
			bookmark: ParentBookmark{Name: "empty"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bookmark.IsFileBased(); got != tt.want {
				t.Errorf("IsFileBased() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParentBookmark_JSON_FileBased(t *testing.T) {
	input := `{"name":"alpha","file":"project.yml","title":"Project Alpha"}`
	var bm ParentBookmark
	if err := json.Unmarshal([]byte(input), &bm); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if bm.Name != "alpha" {
		t.Errorf("Name = %q, want %q", bm.Name, "alpha")
	}
	if bm.File != "project.yml" {
		t.Errorf("File = %q, want %q", bm.File, "project.yml")
	}
	if bm.UUID != "" {
		t.Errorf("UUID = %q, want empty", bm.UUID)
	}
	if !bm.IsFileBased() {
		t.Error("IsFileBased() = false, want true")
	}
}

func TestParentBookmark_JSON_UUIDBased(t *testing.T) {
	input := `{"name":"docs","uuid":"abc-123","title":"Documentation Root"}`
	var bm ParentBookmark
	if err := json.Unmarshal([]byte(input), &bm); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if bm.UUID != "abc-123" {
		t.Errorf("UUID = %q, want %q", bm.UUID, "abc-123")
	}
	if bm.File != "" {
		t.Errorf("File = %q, want empty", bm.File)
	}
	if bm.IsFileBased() {
		t.Error("IsFileBased() = true, want false")
	}
}

func TestParentBookmark_JSON_OmitsEmpty(t *testing.T) {
	bm := ParentBookmark{Name: "alpha", File: "project.yml", Title: "Alpha"}
	data, err := json.Marshal(bm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	s := string(data)
	if strings.Contains(s, `"uuid"`) {
		t.Errorf("JSON should omit empty uuid, got: %s", s)
	}

	bm2 := ParentBookmark{Name: "docs", UUID: "abc-123", Title: "Docs"}
	data2, err := json.Marshal(bm2)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	s2 := string(data2)
	if strings.Contains(s2, `"file"`) {
		t.Errorf("JSON should omit empty file, got: %s", s2)
	}
}
