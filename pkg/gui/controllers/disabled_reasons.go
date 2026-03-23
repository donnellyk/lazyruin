package controllers

import (
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"
)

// requireLinkNote returns a disabled-reason check that fails when the current
// note is nil or not a link note. The getNote parameter lets callers supply
// whatever note-fetching logic is appropriate for their context.
func requireLinkNote(getNote func() *models.Note) func() *types.DisabledReason {
	return func() *types.DisabledReason {
		note := getNote()
		if note == nil || !note.IsLink() {
			return &types.DisabledReason{Text: "Not a link note"}
		}
		return nil
	}
}
