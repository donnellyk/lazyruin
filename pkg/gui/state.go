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
	PreviewModeSingleNote PreviewMode = iota
	PreviewModeCardList
	PreviewModePickResults
)

// CaptureParentInfo tracks the parent selected via > completion in the capture dialog.
type CaptureParentInfo struct {
	UUID  string
	Title string // display title for footer (e.g. "Parent / Child")
}

type GuiState struct {
	Notes           *NotesState
	Queries         *QueriesState
	Tags            *TagsState
	Parents         *ParentsState
	Preview         *PreviewState
	Dialog          *DialogState
	CurrentContext  ContextKey
	PreviousContext ContextKey
	SearchQuery        string
	SearchMode         bool
	CaptureMode        bool
	CaptureParent      *CaptureParentInfo
	SearchCompletion   *CompletionState
	CaptureCompletion  *CompletionState
	PickMode           bool
	PickCompletion     *CompletionState
	PickQuery          string
	PickAnyMode        bool
	PickSeedHash       bool
	Initialized        bool
	lastWidth       int
	lastHeight      int
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

type PreviewState struct {
	Mode              PreviewMode
	Cards             []models.Note
	SelectedCardIndex int
	ScrollOffset      int
	ShowFrontmatter   bool
	ShowTitle         bool
	ShowGlobalTags    bool
	CardLineRanges    [][2]int // [startLine, endLine) for each card
	EditMode          bool     // true when in bulk edit mode (entered from Notes 'e')
	RenderMarkdown    bool     // true to render markdown with glamour
	PickResults       []models.PickResult
}

func NewGuiState() *GuiState {
	return &GuiState{
		Notes: &NotesState{
			CurrentTab: NotesTabAll,
		},
		Queries: &QueriesState{
			CurrentTab: QueriesTabQueries,
		},
		Tags:              &TagsState{CurrentTab: TagsTabAll},
		Parents:           &ParentsState{},
		Preview:           &PreviewState{RenderMarkdown: true},
		SearchCompletion:  NewCompletionState(),
		CaptureCompletion: NewCompletionState(),
		PickCompletion:    NewCompletionState(),
		CurrentContext:     NotesContext,
	}
}
