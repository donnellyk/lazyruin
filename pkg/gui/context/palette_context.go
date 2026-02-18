package context

import "kvnd/lazyruin/pkg/gui/types"

// PaletteContext owns the command palette popup panel.
type PaletteContext struct {
	BaseContext
}

// NewPaletteContext creates a PaletteContext.
func NewPaletteContext() *PaletteContext {
	return &PaletteContext{
		BaseContext: NewBaseContext(NewBaseContextOpts{
			Kind:            types.TEMPORARY_POPUP,
			Key:             "palette",
			ViewNames:       []string{"palette", "paletteList"},
			PrimaryViewName: "palette",
			Focusable:       true,
			Title:           "Command Palette",
		}),
	}
}

var _ types.Context = &PaletteContext{}
