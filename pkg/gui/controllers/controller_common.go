package controllers

import (
	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/helpers"
	"kvnd/lazyruin/pkg/gui/types"
)

// IGuiCommon is a type alias for the authoritative interface in types/.
type IGuiCommon = types.IGuiCommon

// IHelpers provides typed access to helper instances.
type IHelpers interface {
	Refresh() *helpers.RefreshHelper
	Notes() *helpers.NotesHelper
	NoteActions() *helpers.NoteActionsHelper
	Tags() *helpers.TagsHelper
	Queries() *helpers.QueriesHelper
	Editor() *helpers.EditorHelper
	Confirmation() *helpers.ConfirmationHelper
	Search() *helpers.SearchHelper
	Clipboard() *helpers.ClipboardHelper
	Preview() *helpers.PreviewHelper
	PreviewNav() *helpers.PreviewNavHelper
	PreviewLinks() *helpers.PreviewLinksHelper
	PreviewMutations() *helpers.PreviewMutationsHelper
	PreviewLineOps() *helpers.PreviewLineOpsHelper
	PreviewInfo() *helpers.PreviewInfoHelper
	Capture() *helpers.CaptureHelper
	Pick() *helpers.PickHelper
	InputPopup() *helpers.InputPopupHelper
	Snippet() *helpers.SnippetHelper
	Calendar() *helpers.CalendarHelper
	Contrib() *helpers.ContribHelper
	Completion() *helpers.CompletionHelper
}

// ControllerCommon provides shared dependencies for all controllers.
type ControllerCommon struct {
	guiCommon IGuiCommon
	ruinCmd   *commands.RuinCommand
	helpers   IHelpers
}

// NewControllerCommon creates a new ControllerCommon.
func NewControllerCommon(guiCommon IGuiCommon, ruinCmd *commands.RuinCommand, helpers IHelpers) *ControllerCommon {
	return &ControllerCommon{
		guiCommon: guiCommon,
		ruinCmd:   ruinCmd,
		helpers:   helpers,
	}
}

// GuiCommon returns the GUI common interface.
func (self *ControllerCommon) GuiCommon() IGuiCommon {
	return self.guiCommon
}

// RuinCmd returns the ruin command wrapper.
func (self *ControllerCommon) RuinCmd() *commands.RuinCommand {
	return self.ruinCmd
}

// Helpers returns the helpers aggregator.
func (self *ControllerCommon) Helpers() IHelpers {
	return self.helpers
}
