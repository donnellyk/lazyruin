package gui

import "kvnd/lazyruin/pkg/models"

type ContextKey string

const (
	NotesContext   ContextKey = "notes"
	QueriesContext ContextKey = "queries"
	TagsContext    ContextKey = "tags"
	PreviewContext ContextKey = "preview"
	SearchContext  ContextKey = "search"
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
	CurrentContext  ContextKey
	PreviousContext ContextKey
	SearchQuery     string
	SearchMode      bool
}

type NotesState struct {
	Items         []models.Note
	SelectedIndex int
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
}

func NewGuiState() *GuiState {
	return &GuiState{
		Notes:          &NotesState{},
		Queries:        &QueriesState{},
		Tags:           &TagsState{},
		Preview:        &PreviewState{},
		CurrentContext: NotesContext,
	}
}
