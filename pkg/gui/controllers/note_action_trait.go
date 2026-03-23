package controllers

// NoteActionHandlersTrait provides common note-action handler methods that
// delegate to NoteActionsHelper. Embed this in any controller that needs
// addTag / removeTag / setParent / removeParent / toggleBookmark.
type NoteActionHandlersTrait struct {
	c *ControllerCommon
}

func (self *NoteActionHandlersTrait) addTag() error {
	return self.c.Helpers().NoteActions().AddGlobalTag()
}

func (self *NoteActionHandlersTrait) removeTag() error {
	return self.c.Helpers().NoteActions().RemoveTag()
}

func (self *NoteActionHandlersTrait) setParent() error {
	return self.c.Helpers().NoteActions().SetParentDialog()
}

func (self *NoteActionHandlersTrait) removeParent() error {
	return self.c.Helpers().NoteActions().RemoveParent()
}

func (self *NoteActionHandlersTrait) toggleBookmark() error {
	return self.c.Helpers().NoteActions().ToggleBookmark()
}
