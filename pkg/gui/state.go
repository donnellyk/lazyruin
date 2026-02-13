package gui

import "kvnd/lazyruin/pkg/models"

type ContextKey string

const (
	NotesContext        ContextKey = "notes"
	QueriesContext      ContextKey = "queries"
	TagsContext         ContextKey = "tags"
	PreviewContext      ContextKey = "preview"
	SearchContext       ContextKey = "search"
	SearchFilterContext ContextKey = "searchFilter"
	CaptureContext      ContextKey = "capture"
	PickContext         ContextKey = "pick"
	PaletteContext      ContextKey = "palette"
)

// NotesTab represents the sub-tabs within the Notes panel
type NotesTab string

const (
	NotesTabAll    NotesTab = "all"
	NotesTabToday  NotesTab = "today"
	NotesTabRecent NotesTab = "recent"
)

// QueriesTab represents the sub-tabs within the Queries panel
type QueriesTab string

const (
	QueriesTabQueries QueriesTab = "queries"
	QueriesTabParents QueriesTab = "parents"
)

// TagsTab represents the sub-tabs within the Tags panel
type TagsTab string

const (
	TagsTabAll    TagsTab = "all"
	TagsTabGlobal TagsTab = "global"
	TagsTabInline TagsTab = "inline"
)

type PreviewMode int

const (
	PreviewModeCardList PreviewMode = iota
	PreviewModePickResults
)

// CaptureParentInfo tracks the parent selected via > completion in the capture dialog.
type CaptureParentInfo struct {
	UUID  string
	Title string // display title for footer (e.g. "Parent / Child")
}

type GuiState struct {
	Notes                *NotesState
	Queries              *QueriesState
	Tags                 *TagsState
	Parents              *ParentsState
	Preview              *PreviewState
	Dialog               *DialogState
	CurrentContext       ContextKey
	PreviousContext      ContextKey
	SearchQuery          string
	SearchMode           bool
	CaptureMode          bool
	CaptureParent        *CaptureParentInfo
	SearchCompletion     *CompletionState
	CaptureCompletion    *CompletionState
	PickMode             bool
	PickCompletion       *CompletionState
	PickQuery            string
	PickAnyMode          bool
	PickSeedHash         bool
	PaletteMode          bool
	Palette              *PaletteState
	InputPopupMode       bool
	InputPopupCompletion *CompletionState
	InputPopupSeedDone   bool
	InputPopupConfig     *InputPopupConfig
	Initialized          bool
	lastWidth            int
	lastHeight           int
}

type NotesState struct {
	Items         []models.Note
	SelectedIndex int
	CurrentTab    NotesTab
}

type QueriesState struct {
	Items         []models.Query
	SelectedIndex int
	CurrentTab    QueriesTab
}

type ParentsState struct {
	Items         []models.ParentBookmark
	SelectedIndex int
}

type TagsState struct {
	Items         []models.Tag
	SelectedIndex int
	CurrentTab    TagsTab
}

// PreviewLink represents a detected link in the preview content.
type PreviewLink struct {
	Text string // display text (wiki-link target or URL)
	Line int    // absolute line number in the rendered preview
	Col  int    // start column (visible characters, 0-indexed)
	Len  int    // visible length of the link text
}

type PreviewState struct {
	Mode              PreviewMode
	Cards             []models.Note
	SelectedCardIndex int
	ScrollOffset      int
	CursorLine        int // highlighted line in multi-card modes (-1 = no cursor)
	ShowFrontmatter   bool
	ShowTitle         bool
	ShowGlobalTags    bool
	CardLineRanges    [][2]int // [startLine, endLine) for each card
	HeaderLines       []int    // absolute line numbers containing markdown headers
	RenderMarkdown    bool     // true to render markdown with glamour
	PickResults       []models.PickResult
	Links            []PreviewLink // detected links in current render
	HighlightedLink  int           // index into Links; -1 = none, auto-cleared each render
	renderedLink     int           // snapshot of HighlightedLink used during current render
	TemporarilyMoved map[int]bool  // card indices temporarily moved
}

// InputPopupConfig holds the configuration for the generic input popup with completion.
type InputPopupConfig struct {
	Title    string
	Footer   string
	Seed     string                                       // pre-filled text (e.g. ">" or "#")
	Triggers func() []CompletionTrigger                   // provides triggers referencing current completion state
	OnAccept func(raw string, item *CompletionItem) error // raw text and selected item (nil if none)
}

// PaletteCommand represents a single command in the command palette.
type PaletteCommand struct {
	Name     string
	Category string
	Key      string
	OnRun    func() error
	Context  ContextKey // empty = always available
}

// PaletteState holds the runtime state of the command palette.
type PaletteState struct {
	Commands      []PaletteCommand
	Filtered      []PaletteCommand
	SelectedIndex int
	FilterText    string
	OriginContext ContextKey
}

func NewGuiState() *GuiState {
	return &GuiState{
		Notes: &NotesState{
			CurrentTab: NotesTabAll,
		},
		Queries: &QueriesState{
			CurrentTab: QueriesTabQueries,
		},
		Tags:                 &TagsState{CurrentTab: TagsTabAll},
		Parents:              &ParentsState{},
		Preview:              &PreviewState{RenderMarkdown: true, HighlightedLink: -1},
		SearchCompletion:     NewCompletionState(),
		CaptureCompletion:    NewCompletionState(),
		PickCompletion:       NewCompletionState(),
		InputPopupCompletion: NewCompletionState(),
		CurrentContext:       NotesContext,
	}
}
