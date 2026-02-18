package types

// InputPopupConfig holds the configuration for the generic input popup with completion.
type InputPopupConfig struct {
	Title    string
	Footer   string
	Seed     string                                       // pre-filled text (e.g. ">" or "#")
	Triggers func() []CompletionTrigger                   // provides triggers referencing current completion state
	OnAccept func(raw string, item *CompletionItem) error // raw text and selected item (nil if none)
}
