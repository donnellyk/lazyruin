package helpers

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

func TestResolveTypedTag(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		item *types.CompletionItem
		want string
	}{
		{"item wins over raw", "#typed", &types.CompletionItem{Label: "#picked"}, "#picked"},
		{"nil item uses raw", "#newtag", nil, "#newtag"},
		{"raw without hash gets prefix", "newtag", nil, "#newtag"},
		{"raw with whitespace trimmed", "  #spaced  ", nil, "#spaced"},
		{"empty raw and nil item", "", nil, ""},
		{"empty item label falls back to raw", "#fromraw", &types.CompletionItem{Label: ""}, "#fromraw"},
		{"raw with dash supported", "#done-later", nil, "#done-later"},
		{"raw with trailing dash stripped", "#done-", nil, "#done"},
		{"raw junk yields empty", "#", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTypedTag(tt.raw, tt.item)
			if got != tt.want {
				t.Errorf("resolveTypedTag(%q, %+v) = %q, want %q", tt.raw, tt.item, got, tt.want)
			}
		})
	}
}
