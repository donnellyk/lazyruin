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
)

// NotesTab represents the sub-tabs within the Notes panel
type NotesTab string

const (
	NotesTabAll    NotesTab = "all"
	NotesTabToday  NotesTab = "today"
	NotesTabRecent NotesTab = "recent"
)

type PreviewMode int

const (
	PreviewModeSingleNote PreviewMode = iota
	PreviewModeCardList
)

type GuiState struct {
	Notes           *NotesState
	Queries         *QueriesState
	Tags            *TagsState
	Preview         *PreviewState
	Dialog          *DialogState
	CurrentContext  ContextKey
	PreviousContext ContextKey
	SearchQuery     string
	SearchMode      bool
	Initialized     bool
	EditFilePath    string // Path to edit after exiting main loop
}

type NotesState struct {
	Items         []models.Note
	SelectedIndex int
	CurrentTab    NotesTab
}

type QueriesState struct {
	Items         []models.Query
	SelectedIndex int
}

type TagsState struct {
	Items         []models.Tag
	SelectedIndex int
}

type PreviewState struct {
	Mode              PreviewMode
	Cards             []models.Note
	SelectedCardIndex int
	ScrollOffset      int
	ShowFrontmatter   bool
	ShowTitle         bool
	ShowGlobalTags    bool
}

func NewGuiState() *GuiState {
	return &GuiState{
		Notes: &NotesState{
			CurrentTab: NotesTabAll,
		},
		Queries:        &QueriesState{},
		Tags:           &TagsState{},
		Preview:        &PreviewState{},
		CurrentContext: NotesContext,
	}
}
